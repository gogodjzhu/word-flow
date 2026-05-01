package dict

import (
	"testing"
	"time"

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
				Default: "youdao",
				Youdao:  &config.YoudaoConfig{},
			},
			wantErr: false,
		},
		{
			name: "ecdict config",
			dictConfig: &config.DictConfig{
				Default: "ecdict",
				Ecdict:  &config.EcdictConfig{DBFilename: "/tmp/test.db"},
			},
			wantErr: false,
		},
		{
			name: "mwebster config",
			dictConfig: &config.DictConfig{
				Default:   "mwebster",
				MWebster: &config.MWebsterConfig{Key: "test-api-key"},
			},
			wantErr: false,
		},
		{
			name: "llm config with all parameters",
			dictConfig: &config.DictConfig{
				Default: "llm",
				LLM: &config.LLMConfig{
					ApiKey:      "test-api-key",
					URL:         "https://test.com/v1/chat/completions",
					Model:       "gpt-4",
					Timeout:     config.Duration(60 * time.Second),
					MaxTokens:   4000,
					Temperature: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "llm config with missing api_key",
			dictConfig: &config.DictConfig{
				Default: "llm",
				LLM: &config.LLMConfig{
					URL:         "https://test.com/v1/chat/completions",
					Model:       "gpt-4",
					Timeout:     config.Duration(30 * time.Second),
					MaxTokens:   2000,
					Temperature: 0.3,
				},
			},
			wantErr: false, // NewDict doesn't validate; validation happens separately via cfg.Validate()
		},
		{
			name: "etymonline config",
			dictConfig: &config.DictConfig{
				Default:   "etymonline",
				Etymoline: &config.EtymonlineConfig{},
			},
			wantErr: false,
		},
		{
			name: "unknown endpoint",
			dictConfig: &config.DictConfig{
				Default: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDict(tt.dictConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && d == nil {
				t.Error("NewDict() returned nil dict when wantErr was false")
			}
		})
	}
}