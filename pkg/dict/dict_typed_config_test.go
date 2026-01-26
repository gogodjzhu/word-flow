package dict

import (
	"testing"

	"github.com/gogodjzhu/word-flow/internal/config"
)

func TestNewDict_WithTypedConfigs(t *testing.T) {
	tests := []struct {
		name       string
		dictConfig *config.DictConfig
		wantErr    bool
	}{
		{
			name: "youdao config",
			dictConfig: &config.DictConfig{
				Default:    "youdao",
				Parameters: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "ecdict config",
			dictConfig: &config.DictConfig{
				Default: "ecdict",
				Parameters: map[string]interface{}{
					"ecdict.dbfilename": "/tmp/test.db",
				},
			},
			wantErr: false,
		},
		{
			name: "mwebster config",
			dictConfig: &config.DictConfig{
				Default: "mwebster",
				Parameters: map[string]interface{}{
					"mwebster.key": "test-api-key",
				},
			},
			wantErr: false,
		},
		{
			name: "llm config with all parameters",
			dictConfig: &config.DictConfig{
				Default: "llm",
				Parameters: map[string]interface{}{
					"llm.api_key":     "test-api-key",
					"llm.url":         "https://test.com/v1/chat/completions",
					"llm.model":       "gpt-4",
					"llm.timeout":     "60s",
					"llm.max_tokens":  4000,
					"llm.temperature": 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "llm config with defaults",
			dictConfig: &config.DictConfig{
				Default:    "llm",
				Parameters: map[string]interface{}{},
			},
			wantErr: true, // Should fail due to missing required api_key
		},
		{
			name: "etymonline config",
			dictConfig: &config.DictConfig{
				Default:    "etymonline",
				Parameters: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "unknown endpoint",
			dictConfig: &config.DictConfig{
				Default: "unknown",
				Parameters: map[string]interface{}{
					"test.key": "value",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dict, err := NewDict(tt.dictConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && dict == nil {
				t.Error("NewDict() returned nil dict when wantErr was false")
			}
		})
	}
}

func TestNewDict_LLMConfigTypeConversion(t *testing.T) {
	// Test that various parameter types are correctly converted
	tests := []struct {
		name       string
		parameters map[string]interface{}
		wantErr    bool
	}{
		{
			name: "integer max_tokens",
			parameters: map[string]interface{}{
				"llm.api_key":    "test-key",
				"llm.max_tokens": 3000,
			},
			wantErr: false,
		},
		{
			name: "string max_tokens",
			parameters: map[string]interface{}{
				"llm.api_key":    "test-key",
				"llm.max_tokens": "3000",
			},
			wantErr: false,
		},
		{
			name: "float temperature",
			parameters: map[string]interface{}{
				"llm.api_key":     "test-key",
				"llm.temperature": 0.8,
			},
			wantErr: false,
		},
		{
			name: "string temperature",
			parameters: map[string]interface{}{
				"llm.api_key":     "test-key",
				"llm.temperature": "0.8",
			},
			wantErr: false,
		},
		{
			name: "timeout as duration string",
			parameters: map[string]interface{}{
				"llm.api_key": "test-key",
				"llm.timeout": "45s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dictConfig := &config.DictConfig{
				Default:    "llm",
				Parameters: tt.parameters,
			}

			dict, err := NewDict(dictConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && dict == nil {
				t.Error("NewDict() returned nil dict when wantErr was false")
			}
		})
	}
}
