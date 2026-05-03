package baidu

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/internal/config"
	httputil "github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/translator/types"
	"github.com/pkg/errors"
)

type TranslatorBaidu struct {
	cfg *config.TransBaiduConfig
}

func NewTranslatorBaidu(cfg *config.TransBaiduConfig) *TranslatorBaidu {
	return &TranslatorBaidu{cfg: cfg}
}

type baiduResponse struct {
	From        string `json:"from"`
	To          string `json:"to"`
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
}

func generateSign(appid, text, salt, secret string) string {
	signStr := appid + text + salt + secret
	h := md5.Sum([]byte(signStr))
	return hex.EncodeToString(h[:])
}

func callBaiduTranslate(text string, cfg *config.TransBaiduConfig) (string, error) {
	salt := strconv.FormatInt(time.Now().Unix(), 10)
	sign := generateSign(cfg.AppID, text, salt, cfg.Secret)

	apiURL := "https://api.fanyi.baidu.com/api/trans/vip/translate"

	params := url.Values{}
	params.Set("q", text)
	params.Set("from", "auto")
	params.Set("to", "zh")
	params.Set("appid", cfg.AppID)
	params.Set("salt", salt)
	params.Set("sign", sign)

	body := []byte(params.Encode())
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	result, err := httputil.SendPost(apiURL, headers, body, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status: %s", response.Status)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		var resp baiduResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, errors.Wrap(err, "failed to parse response")
		}

		if len(resp.TransResult) == 0 {
			return nil, errors.New("translation result is empty")
		}

		return resp, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to call Baidu Translate API")
	}

	baiduResp := result.(baiduResponse)
	var translated strings.Builder
	for i, item := range baiduResp.TransResult {
		translated.WriteString(item.Dst)
		if i < len(baiduResp.TransResult)-1 {
			translated.WriteString(" ")
		}
	}
	return translated.String(), nil
}

func (t *TranslatorBaidu) Translate(text string, out io.Writer, opts *types.TransOptions) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("text is empty")
	}
	if opts == nil {
		opts = &types.TransOptions{}
	}

	switch {
	case !opts.Ref && !opts.NoStream:
		translated, err := callBaiduTranslate(text, t.cfg)
		if err != nil {
			return err
		}
		fmt.Fprint(out, translated)
		return nil

	case !opts.Ref && opts.NoStream:
		translated, err := callBaiduTranslate(text, t.cfg)
		if err != nil {
			return err
		}
		fmt.Fprint(out, translated)
		return nil

	case opts.Ref && !opts.NoStream:
		translated, err := callBaiduTranslate(text, t.cfg)
		if err != nil {
			return err
		}
		refPairs := []map[string]string{
			{"raw": text, "translation": translated},
		}
		data, err := json.Marshal(refPairs)
		if err != nil {
			return errors.Wrap(err, "failed to marshal ref pairs")
		}
		out.Write(data)
		return nil

	default:
		translated, err := callBaiduTranslate(text, t.cfg)
		if err != nil {
			return err
		}
		refPairs := []map[string]string{
			{"raw": text, "translation": translated},
		}
		data, err := json.Marshal(refPairs)
		if err != nil {
			return errors.Wrap(err, "failed to marshal ref pairs")
		}
		out.Write(data)
		return nil
	}
}