package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateForTrans_ValidatesTransNotDict(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `version: v1
dict:
  default: youdao
  llm:
    api_key: dict-key
    url: "https://dict.example.com"
    model: dict-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
trans:
  default: google
  google: {}
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateForTrans(cfg); err != nil {
		t.Errorf("expected no error for google trans, got %v", err)
	}
}

func TestValidateForTrans_GoogleNoApiKey(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `version: v1
dict:
  default: youdao
trans:
  default: google
  google: {}
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateForTrans(cfg); err != nil {
		t.Errorf("expected no error for google trans without api key, got %v", err)
	}
}

func TestValidateForTrans_LLMRequiresFields(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectedErr string
	}{
		{
			name: "llm missing api_key",
			configYAML: `version: v1
dict:
  default: youdao
trans:
  default: llm
  llm:
    url: "https://api.example.com"
    model: test-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "trans.llm.api_key is required",
		},
		{
			name: "llm missing url",
			configYAML: `version: v1
dict:
  default: youdao
trans:
  default: llm
  llm:
    api_key: test-key
    model: test-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "trans.llm.url is required",
		},
		{
			name: "llm missing model",
			configYAML: `version: v1
dict:
  default: youdao
trans:
  default: llm
  llm:
    api_key: test-key
    url: "https://api.example.com"
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "trans.llm.model is required",
		},
		{
			name: "llm valid config",
			configYAML: `version: v1
dict:
  default: youdao
trans:
  default: llm
  llm:
    api_key: test-key
    url: "https://api.example.com"
    model: test-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(tt.configYAML), 0644); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfig(configFile)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateForTrans(cfg)
			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedErr)
				} else if !containsSubstr(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

func TestGetEndpointConfig(t *testing.T) {
	tc := &TransConfig{
		Default: "llm",
		LLM:     &TransLLMConfig{ApiKey: "key", URL: "http://test", Model: "m", MaxTokens: 100, Temperature: 0.3},
		Google:  &TransGoogleConfig{},
	}

	llmCfg, err := tc.GetEndpointConfig("llm")
	if err != nil {
		t.Fatalf("expected no error for llm, got %v", err)
	}
	if _, ok := llmCfg.(*TransLLMConfig); !ok {
		t.Errorf("expected *TransLLMConfig, got %T", llmCfg)
	}

	googleCfg, err := tc.GetEndpointConfig("google")
	if err != nil {
		t.Fatalf("expected no error for google, got %v", err)
	}
	if _, ok := googleCfg.(*TransGoogleConfig); !ok {
		t.Errorf("expected *TransGoogleConfig, got %T", googleCfg)
	}

	_, err = tc.GetEndpointConfig("unknown")
	if err == nil {
		t.Error("expected error for unknown endpoint, got nil")
	}
}

func TestTransConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `version: v1
dict:
  default: youdao
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Trans.Default != "google" {
		t.Errorf("expected trans default 'google', got %q", cfg.Trans.Default)
	}
	if cfg.Trans.LLM == nil {
		t.Error("expected trans LLM config to be initialized with defaults")
	}
	if cfg.Trans.LLM.Timeout != Duration(30*time.Second) {
		t.Errorf("expected default trans LLM timeout 30s, got %v", cfg.Trans.LLM.Timeout)
	}
	if cfg.Trans.LLM.MaxTokens != 2000 {
		t.Errorf("expected default trans LLM max_tokens 2000, got %d", cfg.Trans.LLM.MaxTokens)
	}
	if cfg.Trans.LLM.Temperature != 0.3 {
		t.Errorf("expected default trans LLM temperature 0.3, got %f", cfg.Trans.LLM.Temperature)
	}
}