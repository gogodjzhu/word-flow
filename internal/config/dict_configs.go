package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

// DictEndpointConfig 基础配置接口
type DictEndpointConfig interface {
	Validate() error
	GetDefaults() interface{}
}

// YoudaoConfig 有道词典配置
type YoudaoConfig struct{}

func (c *YoudaoConfig) Validate() error {
	return nil
}

func (c *YoudaoConfig) GetDefaults() interface{} {
	return map[string]interface{}{}
}

// EtymonlineConfig Etymonline词典配置
type EtymonlineConfig struct{}

func (c *EtymonlineConfig) Validate() error {
	return nil
}

func (c *EtymonlineConfig) GetDefaults() interface{} {
	return map[string]interface{}{}
}

// EcdictConfig ECDict词典配置
type EcdictConfig struct {
	DBFilename string `yaml:"ecdict.dbfilename" json:"ecdict.dbfilename"`
}

func (c *EcdictConfig) Validate() error {
	if c.DBFilename == "" {
		return errors.New("ecdict.dbfilename is required")
	}
	return nil
}

func (c *EcdictConfig) GetDefaults() interface{} {
	var path string
	if a := os.Getenv("WORDFLOW_HOME"); a != "" {
		path = a
	} else {
		d, _ := os.UserHomeDir()
		path = filepath.Join(d, ".config", "wordflow")
	}
	return map[string]interface{}{
		"ecdict.dbfilename": filepath.Join(path, "stardict.db"),
	}
}

// MWebsterConfig Merriam-Webster词典配置
type MWebsterConfig struct {
	Key string `yaml:"mwebster.key" json:"mwebster.key"`
}

func (c *MWebsterConfig) Validate() error {
	if c.Key == "" {
		return errors.New("mwebster.key is required")
	}
	return nil
}

func (c *MWebsterConfig) GetDefaults() interface{} {
	return map[string]interface{}{}
}

// LLMConfig LLM词典配置
type LLMConfig struct {
	ApiKey      string        `yaml:"llm.api_key" json:"llm.api_key"`
	URL         string        `yaml:"llm.url" json:"llm.url"`
	Model       string        `yaml:"llm.model" json:"llm.model"`
	Timeout     time.Duration `yaml:"llm.timeout" json:"llm.timeout"`
	MaxTokens   int           `yaml:"llm.max_tokens" json:"llm.max_tokens"`
	Temperature float64       `yaml:"llm.temperature" json:"llm.temperature"`
}

func (c *LLMConfig) Validate() error {
	if c.URL == "" {
		return errors.New("llm.url is required")
	}
	if c.Model == "" {
		return errors.New("llm.model is required")
	}
	if c.ApiKey == "" {
		return errors.New("llm.api_key is required")
	}
	if c.MaxTokens <= 0 {
		return errors.New("llm.max_tokens must be positive")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("llm.temperature must be between 0 and 2")
	}
	return nil
}

func (c *LLMConfig) GetDefaults() interface{} {
	return map[string]interface{}{
		"llm.url":         "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		"llm.model":       "glm-4",
		"llm.timeout":     "30s",
		"llm.max_tokens":  2000,
		"llm.temperature": 0.3,
	}
}
