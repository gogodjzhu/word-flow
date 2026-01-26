package config

import (
	"testing"
	"time"
)

func TestGetConfigForEndpoint_Ecdict(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]interface{}
		wantErr    bool
		expected   *EcdictConfig
	}{
		{
			name: "valid config",
			parameters: map[string]interface{}{
				"ecdict.dbfilename": "/test/path.db",
			},
			wantErr: false,
			expected: &EcdictConfig{
				DBFilename: "/test/path.db",
			},
		},
		{
			name:       "with defaults - should succeed with default path",
			parameters: map[string]interface{}{},
			wantErr:    false,
			expected: &EcdictConfig{
				DBFilename: "", // We don't test the exact default value as it depends on environment
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dictConfig := &DictConfig{
				Parameters: tt.parameters,
			}

			config, err := dictConfig.GetConfigForEndpoint("ecdict")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigForEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				ecdictConfig, ok := config.(*EcdictConfig)
				if !ok {
					t.Errorf("Expected *EcdictConfig, got %T", config)
					return
				}

				if tt.expected.DBFilename != "" && ecdictConfig.DBFilename != tt.expected.DBFilename {
					t.Errorf("DBFilename = %v, want %v", ecdictConfig.DBFilename, tt.expected.DBFilename)
				}
			}
		})
	}
}

func TestGetConfigForEndpoint_LLM(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]interface{}
		wantErr    bool
		expected   *LLMConfig
	}{
		{
			name: "valid config with all fields",
			parameters: map[string]interface{}{
				"llm.api_key":     "test-api-key",
				"llm.url":         "https://test.com/v1/chat/completions",
				"llm.model":       "gpt-4",
				"llm.timeout":     "60s",
				"llm.max_tokens":  4000,
				"llm.temperature": 0.7,
			},
			wantErr: false,
			expected: &LLMConfig{
				ApiKey:      "test-api-key",
				URL:         "https://test.com/v1/chat/completions",
				Model:       "gpt-4",
				Timeout:     60 * time.Second,
				MaxTokens:   4000,
				Temperature: 0.7,
			},
		},
		{
			name: "config with defaults",
			parameters: map[string]interface{}{
				"llm.api_key": "test-api-key",
			},
			wantErr: false,
			expected: &LLMConfig{
				ApiKey:      "test-api-key",
				URL:         "https://open.bigmodel.cn/api/paas/v4/chat/completions",
				Model:       "glm-4",
				Timeout:     30 * time.Second,
				MaxTokens:   2000,
				Temperature: 0.3,
			},
		},
		{
			name: "missing required api_key",
			parameters: map[string]interface{}{
				"llm.url": "https://test.com",
			},
			wantErr: true,
		},
		{
			name: "string max_tokens",
			parameters: map[string]interface{}{
				"llm.api_key":    "test-api-key",
				"llm.max_tokens": "3000",
			},
			wantErr: false,
			expected: &LLMConfig{
				ApiKey:      "test-api-key",
				URL:         "https://open.bigmodel.cn/api/paas/v4/chat/completions",
				Model:       "glm-4",
				Timeout:     30 * time.Second,
				MaxTokens:   3000,
				Temperature: 0.3,
			},
		},
		{
			name: "string temperature",
			parameters: map[string]interface{}{
				"llm.api_key":     "test-api-key",
				"llm.temperature": "0.8",
			},
			wantErr: false,
			expected: &LLMConfig{
				ApiKey:      "test-api-key",
				URL:         "https://open.bigmodel.cn/api/paas/v4/chat/completions",
				Model:       "glm-4",
				Timeout:     30 * time.Second,
				MaxTokens:   2000,
				Temperature: 0.8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dictConfig := &DictConfig{
				Parameters: tt.parameters,
			}

			config, err := dictConfig.GetConfigForEndpoint("llm")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigForEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				llmConfig, ok := config.(*LLMConfig)
				if !ok {
					t.Errorf("Expected *LLMConfig, got %T", config)
					return
				}

				if llmConfig.ApiKey != tt.expected.ApiKey {
					t.Errorf("ApiKey = %v, want %v", llmConfig.ApiKey, tt.expected.ApiKey)
				}
				if llmConfig.URL != tt.expected.URL {
					t.Errorf("URL = %v, want %v", llmConfig.URL, tt.expected.URL)
				}
				if llmConfig.Model != tt.expected.Model {
					t.Errorf("Model = %v, want %v", llmConfig.Model, tt.expected.Model)
				}
				if llmConfig.Timeout != tt.expected.Timeout {
					t.Errorf("Timeout = %v, want %v", llmConfig.Timeout, tt.expected.Timeout)
				}
				if llmConfig.MaxTokens != tt.expected.MaxTokens {
					t.Errorf("MaxTokens = %v, want %v", llmConfig.MaxTokens, tt.expected.MaxTokens)
				}
				if llmConfig.Temperature != tt.expected.Temperature {
					t.Errorf("Temperature = %v, want %v", llmConfig.Temperature, tt.expected.Temperature)
				}
			}
		})
	}
}

func TestGetConfigForEndpoint_Unknown(t *testing.T) {
	dictConfig := &DictConfig{
		Parameters: map[string]interface{}{
			"test.key": "value",
		},
	}

	_, err := dictConfig.GetConfigForEndpoint("unknown")
	if err == nil {
		t.Error("Expected error for unknown endpoint")
	}
}
