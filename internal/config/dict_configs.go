package config

import (
	"errors"
)

type DictEndpointConfig interface {
	Validate() error
}

type YoudaoConfig struct{}

func (c *YoudaoConfig) Validate() error {
	return nil
}

type EtymonlineConfig struct{}

func (c *EtymonlineConfig) Validate() error {
	return nil
}

type EcdictConfig struct {
	DBFilename string `yaml:"db_filename,omitempty"`
}

func (c *EcdictConfig) Validate() error {
	if c.DBFilename == "" {
		return errors.New("ecdict.db_filename is required when ecdict is the default dictionary")
	}
	return nil
}

type MWebsterConfig struct {
	Key string `yaml:"key,omitempty"`
}

func (c *MWebsterConfig) Validate() error {
	if c.Key == "" {
		return errors.New("mwebster.key is required when mwebster is the default dictionary")
	}
	return nil
}

type LLMConfig struct {
	ApiKey      string   `yaml:"api_key,omitempty"`
	URL         string   `yaml:"url,omitempty"`
	Model       string   `yaml:"model,omitempty"`
	Timeout     Duration `yaml:"timeout,omitempty"`
	MaxTokens   int      `yaml:"max_tokens,omitempty"`
	Temperature float64  `yaml:"temperature,omitempty"`
}

func (c *LLMConfig) Validate() error {
	if c.ApiKey == "" {
		return errors.New("llm.api_key is required. Set it via: wordflow config set dict.llm.api_key <key> or env var WORDFLOW_DICT_LLM_API_KEY")
	}
	if c.URL == "" {
		return errors.New("llm.url is required. Set it via: wordflow config set dict.llm.url <url> or env var WORDFLOW_DICT_LLM_URL")
	}
	if c.Model == "" {
		return errors.New("llm.model is required. Set it via: wordflow config set dict.llm.model <model> or env var WORDFLOW_DICT_LLM_MODEL")
	}
	if c.MaxTokens <= 0 {
		return errors.New("llm.max_tokens must be positive")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("llm.temperature must be between 0 and 2")
	}
	return nil
}

type GoogleConfig struct{}

func (c *GoogleConfig) Validate() error {
	return nil
}