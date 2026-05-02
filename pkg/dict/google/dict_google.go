package dict_google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/gogodjzhu/word-flow/internal/config"
	httputil "github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
)

type DictGoogle struct {
	cfg *config.GoogleConfig
}

func NewDictGoogle(cfg *config.GoogleConfig) (*DictGoogle, error) {
	return &DictGoogle{cfg: cfg}, nil
}

func parseDictResponse(body []byte, word string) (*entity.WordItem, error) {
	if len(body) == 0 {
		return nil, errors.New("empty response from Google Dictionary API")
	}
	var responseArr []interface{}
	if err := json.Unmarshal(body, &responseArr); err != nil {
		return nil, errors.Wrap(err, "failed to parse response")
	}

	wordItem := &entity.WordItem{
		ID:     entity.WordId(word),
		Word:   word,
		Source: "google",
	}

	var bdData []interface{}
	if len(responseArr) > 1 {
		bdData, _ = responseArr[1].([]interface{})
	}

	if len(bdData) > 0 {
		for _, entry := range bdData {
			entryArr, ok := entry.([]interface{})
			if !ok || len(entryArr) < 2 {
				continue
			}
			pos, _ := entryArr[0].(string)
			translations, _ := entryArr[1].([]interface{})

			var transStrs []string
			for _, t := range translations {
				if s, ok := t.(string); ok {
					transStrs = append(transStrs, s)
				}
			}

			meaning := &entity.WordMeaning{
				PartOfSpeech: abbreviatePos(pos),
				Definitions:  strings.Join(transStrs, "; "),
			}
			wordItem.WordMeanings = append(wordItem.WordMeanings, meaning)
		}
	} else {
		var translation strings.Builder
		if len(responseArr) > 0 {
			segments, _ := responseArr[0].([]interface{})
			for _, seg := range segments {
				segArr, ok := seg.([]interface{})
				if !ok || len(segArr) < 1 {
					continue
				}
				if translated, ok := segArr[0].(string); ok {
					translation.WriteString(translated)
				}
			}
		}
		if translation.Len() > 0 {
			wordItem.WordMeanings = append(wordItem.WordMeanings, &entity.WordMeaning{
				PartOfSpeech: "",
				Definitions:  translation.String(),
			})
		}
	}

	return wordItem, nil
}

func (d *DictGoogle) Search(word string) (*entity.WordItem, error) {
	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=zh-CN&dt=t&dt=bd&q=%s",
		neturl.QueryEscape(word),
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
		return parseDictResponse(body, word)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Google Dictionary API")
	}
	return result.(*entity.WordItem), nil
}

func abbreviatePos(pos string) string {
	switch strings.ToLower(strings.TrimSpace(pos)) {
	case "adjective":
		return "adj."
	case "noun":
		return "n."
	case "verb":
		return "v."
	case "adverb":
		return "adv."
	case "preposition":
		return "prep."
	case "conjunction":
		return "conj."
	case "pronoun":
		return "pron."
	case "interjection":
		return "int."
	default:
		return pos
	}
}