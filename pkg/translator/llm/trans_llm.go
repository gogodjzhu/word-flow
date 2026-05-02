package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/llm"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/translator/types"
	"github.com/pkg/errors"
)

type TranslatorLLM struct {
	cfg *config.TransLLMConfig
}

func NewTranslatorLLM(cfg *config.TransLLMConfig) *TranslatorLLM {
	return &TranslatorLLM{cfg: cfg}
}

func (t *TranslatorLLM) Translate(text string, out io.Writer, opts *types.TransOptions) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("empty input text")
	}

	promptBuilder := llm.NewPromptBuilder()
	systemPrompt := promptBuilder.BuildTranslationPrompt(opts.Ref)

	request := &llm.ChatRequest{
		Model: t.cfg.Model,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
		Stream:      !opts.NoStream,
		Temperature: t.cfg.Temperature,
		MaxTokens:   t.cfg.MaxTokens,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "failed to marshal translate request")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + t.cfg.ApiKey,
	}

	if opts.NoStream {
		return t.handleNonStreaming(headers, requestBytes, out)
	}
	return t.handleStreaming(headers, requestBytes, out)
}

func (t *TranslatorLLM) handleNonStreaming(headers map[string]string, requestBytes []byte, out io.Writer) error {
	responseProcessor := llm.NewResponseProcessor()

	result, err := util.SendPostWithTimeout(t.cfg.URL, headers, requestBytes, time.Duration(t.cfg.Timeout), func(response *http.Response) (interface{}, error) {
		return responseProcessor.ProcessResponse(response)
	})
	if err != nil {
		return err
	}

	translation := result.(string)
	fmt.Fprint(out, translation)
	return nil
}

func (t *TranslatorLLM) handleStreaming(headers map[string]string, requestBytes []byte, out io.Writer) error {
	streamProcessor := llm.NewStreamProcessor(out)

	err := util.SendPostStreamWithTimeout(t.cfg.URL, headers, requestBytes, time.Duration(t.cfg.Timeout), func(response *http.Response) error {
		_, err := streamProcessor.ProcessStream(response)
		return err
	})
	return err
}