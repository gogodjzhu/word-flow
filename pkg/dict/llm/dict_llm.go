package dict_llm

import (
	"fmt"

	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/llm"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
)

type LLMClient interface {
	TranslateAndExplain(text string) (*entity.WordItem, error)
}

type DictLLM struct {
	client LLMClient
}

func NewDictLLM(config *config.LLMConfig) (*DictLLM, error) {
	if config == nil {
		return nil, errors.New("llm config is required")
	}

	// 配置已经验证过，直接使用
	client := llm.NewClient(
		config.ApiKey,
		config.URL,
		config.Model,
		config.Timeout,
		config.MaxTokens,
		config.Temperature,
	)

	return &DictLLM{
		client: client,
	}, nil
}

func (d *DictLLM) Search(word string) (*entity.WordItem, error) {
	if word == "" {
		return nil, errors.New("empty word to search")
	}

	result, err := d.client.TranslateAndExplain(word)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to translate and explain word: %s", word))
	}

	if result == nil {
		return nil, errors.New("empty result from LLM")
	}

	result.Source = "llm"
	return result, nil
}

func (d *DictLLM) SetClient(client LLMClient) {
	d.client = client
}
