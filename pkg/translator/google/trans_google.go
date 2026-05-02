package google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/gogodjzhu/word-flow/internal/config"
	httputil "github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/translator/types"
	"github.com/gogodjzhu/word-flow/pkg/util"
	"github.com/pkg/errors"
)

type TranslatorGoogle struct {
	cfg *config.TransGoogleConfig
}

func NewTranslatorGoogle(cfg *config.TransGoogleConfig) *TranslatorGoogle {
	return &TranslatorGoogle{cfg: cfg}
}

type translatedSegment struct {
	translation string
	original    string
}

func parseTranslateResponse(body []byte) ([]translatedSegment, error) {
	if len(body) == 0 {
		return []translatedSegment{}, nil
	}
	var responseArr []interface{}
	if err := json.Unmarshal(body, &responseArr); err != nil {
		return nil, errors.Wrap(err, "failed to parse response")
	}
	if len(responseArr) == 0 {
		return []translatedSegment{}, nil
	}
	segments, ok := responseArr[0].([]interface{})
	if !ok {
		return []translatedSegment{}, nil
	}
	var results []translatedSegment
	for _, seg := range segments {
		segArr, ok := seg.([]interface{})
		if !ok || len(segArr) < 2 {
			continue
		}
		translated, _ := segArr[0].(string)
		original, _ := segArr[1].(string)
		results = append(results, translatedSegment{
			translation: translated,
			original:    original,
		})
	}
	return results, nil
}

func callGoogleTranslate(text string) ([]translatedSegment, error) {
	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=zh-CN&dt=t&q=%s",
		neturl.QueryEscape(text),
	)
	result, err := httputil.SendGet(url, nil, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status: %s", response.Status)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}
		return parseTranslateResponse(body)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Google Translate API")
	}
	return result.([]translatedSegment), nil
}

func (t *TranslatorGoogle) Translate(text string, out io.Writer, opts *types.TransOptions) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("text is empty")
	}
	if opts == nil {
		opts = &types.TransOptions{}
	}

	segments := util.SegmentText(text)
	batches := util.BatchSegments(segments, 500)

	switch {
	case !opts.Ref && !opts.NoStream:
		for i, batch := range batches {
			batchText := strings.Join(batch, " ")
			segResults, err := callGoogleTranslate(batchText)
			if err != nil {
				return errors.Wrap(err, "failed to translate batch")
			}
			for _, seg := range segResults {
				fmt.Fprint(out, seg.translation)
			}
			if i < len(batches)-1 {
				fmt.Fprint(out, " ")
			}
		}
		return nil

	case !opts.Ref && opts.NoStream:
		var result strings.Builder
		for i, batch := range batches {
			batchText := strings.Join(batch, " ")
			segResults, err := callGoogleTranslate(batchText)
			if err != nil {
				return errors.Wrap(err, "failed to translate batch")
			}
			for _, seg := range segResults {
				result.WriteString(seg.translation)
			}
			if i < len(batches)-1 {
				result.WriteString(" ")
			}
		}
		fmt.Fprint(out, result.String())
		return nil

	case opts.Ref && !opts.NoStream:
		fmt.Fprint(out, "[")
		first := true
		for _, batch := range batches {
			batchText := strings.Join(batch, " ")
			segResults, err := callGoogleTranslate(batchText)
			if err != nil {
				return errors.Wrap(err, "failed to translate batch")
			}
			for _, seg := range segResults {
				if !first {
					fmt.Fprint(out, ",")
				}
				first = false
				pair := map[string]string{
					"raw":         seg.original,
					"translation": seg.translation,
				}
				data, err := json.Marshal(pair)
				if err != nil {
					return errors.Wrap(err, "failed to marshal ref pair")
				}
				out.Write(data)
			}
		}
		fmt.Fprint(out, "]")
		return nil

	default:
		var allPairs []map[string]string
		for _, batch := range batches {
			batchText := strings.Join(batch, " ")
			segResults, err := callGoogleTranslate(batchText)
			if err != nil {
				return errors.Wrap(err, "failed to translate batch")
			}
			for _, seg := range segResults {
				allPairs = append(allPairs, map[string]string{
					"raw":         seg.original,
					"translation": seg.translation,
				})
			}
		}
		data, err := json.Marshal(allPairs)
		if err != nil {
			return errors.Wrap(err, "failed to marshal ref pairs")
		}
		out.Write(data)
		return nil
	}
}