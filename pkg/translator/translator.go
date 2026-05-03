package translator

import (
	"io"

	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	trans_baidu "github.com/gogodjzhu/word-flow/pkg/translator/baidu"
	trans_google "github.com/gogodjzhu/word-flow/pkg/translator/google"
	trans_llm "github.com/gogodjzhu/word-flow/pkg/translator/llm"
	"github.com/gogodjzhu/word-flow/pkg/translator/types"
	"github.com/pkg/errors"
)

type TransOptions = types.TransOptions

type Translator interface {
	Translate(text string, out io.Writer, opts *TransOptions) error
}

type Endpoint string

const (
	TransLLM    Endpoint = "llm"
	TransGoogle Endpoint = "google"
	TransBaidu  Endpoint = "baidu"
)

type TranslatorInfo struct {
	Name        string
	Description string
}

func AvailableTranslators() []TranslatorInfo {
	return []TranslatorInfo{
		{
			Name:        string(TransLLM),
			Description: "AI-powered translation using Large Language Models, requires LLM API key and endpoint.",
		},
		{
			Name:        string(TransGoogle),
			Description: "[Free] Google Translate API for translation.",
		},
		{
			Name:        string(TransBaidu),
			Description: "[API] Baidu Translate API, requires app_id and secret.",
		},
	}
}

func NewTranslator(cfg *config.TransConfig) (Translator, error) {
	endpointConfig, err := cfg.GetEndpointConfig(cfg.Default)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get endpoint config")
	}

	if err := endpointConfig.Validate(); err != nil {
		return nil, errors.Wrap(err, "config validation failed")
	}

	switch Endpoint(cfg.Default) {
	case TransLLM:
		return trans_llm.NewTranslatorLLM(cfg.LLM), nil
	case TransGoogle:
		return trans_google.NewTranslatorGoogle(cfg.Google), nil
	case TransBaidu:
		return trans_baidu.NewTranslatorBaidu(cfg.Baidu), nil
	default:
		return nil, buzz_error.InvalidEndpoint(cfg.Default)
	}
}