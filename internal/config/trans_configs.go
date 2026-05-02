package config

import "errors"

type TransEndpointConfig interface {
	Validate() error
}

type TransLLMConfig struct {
	ApiKey      string   `yaml:"api_key,omitempty"`
	URL         string   `yaml:"url,omitempty"`
	Model       string   `yaml:"model,omitempty"`
	Timeout     Duration `yaml:"timeout,omitempty"`
	MaxTokens   int      `yaml:"max_tokens,omitempty"`
	Temperature float64  `yaml:"temperature,omitempty"`
}

func (c *TransLLMConfig) Validate() error {
	if c.ApiKey == "" {
		return errors.New("trans.llm.api_key is required. Set it via: wordflow config set trans.llm.api_key <key> or env var WORDFLOW_TRANS_LLM_API_KEY")
	}
	if c.URL == "" {
		return errors.New("trans.llm.url is required. Set it via: wordflow config set trans.llm.url <url> or env var WORDFLOW_TRANS_LLM_URL")
	}
	if c.Model == "" {
		return errors.New("trans.llm.model is required. Set it via: wordflow config set trans.llm.model <model> or env var WORDFLOW_TRANS_LLM_MODEL")
	}
	if c.MaxTokens <= 0 {
		return errors.New("trans.llm.max_tokens must be positive")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("trans.llm.temperature must be between 0 and 2")
	}
	return nil
}

type TransGoogleConfig struct{}

func (c *TransGoogleConfig) Validate() error {
	return nil
}