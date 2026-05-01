package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `version: v1
dict:
  default: youdao
  youdao: {}
  llm:
    api_key: test-key
    url: "https://api.example.com/v1/chat/completions"
    model: test-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
  ecdict: {}
  etymonline: {}
  mwebster: {}
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

	if cfg.Version != "v1" {
		t.Errorf("expected version v1, got %q", cfg.Version)
	}
	if cfg.Dict.Default != "youdao" {
		t.Errorf("expected dict default youdao, got %q", cfg.Dict.Default)
	}
	if cfg.Dict.LLM.ApiKey != "test-key" {
		t.Errorf("expected LLM api_key test-key, got %q", cfg.Dict.LLM.ApiKey)
	}
	if cfg.Dict.LLM.Model != "test-model" {
		t.Errorf("expected LLM model test-model, got %q", cfg.Dict.LLM.Model)
	}
	if time.Duration(cfg.Dict.LLM.Timeout) != 30*time.Second {
		t.Errorf("expected LLM timeout 30s, got %v", time.Duration(cfg.Dict.LLM.Timeout))
	}
	if cfg.Dict.LLM.MaxTokens != 2000 {
		t.Errorf("expected LLM max_tokens 2000, got %d", cfg.Dict.LLM.MaxTokens)
	}
	if cfg.Dict.LLM.Temperature != 0.3 {
		t.Errorf("expected LLM temperature 0.3, got %f", cfg.Dict.LLM.Temperature)
	}
	if cfg.Notebook.Default != "default" {
		t.Errorf("expected notebook default 'default', got %q", cfg.Notebook.Default)
	}
	if cfg.Notebook.Settings.MaxReviews != 50 {
		t.Errorf("expected max_reviews 50, got %d", cfg.Notebook.Settings.MaxReviews)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
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

	if cfg.Dict.LLM == nil {
		t.Error("expected LLM config to be initialized with defaults")
	}
	if cfg.Dict.LLM.Timeout != Duration(30*time.Second) {
		t.Errorf("expected default timeout 30s, got %v", cfg.Dict.LLM.Timeout)
	}
	if cfg.Dict.LLM.MaxTokens != 2000 {
		t.Errorf("expected default max_tokens 2000, got %d", cfg.Dict.LLM.MaxTokens)
	}
	if cfg.Dict.LLM.Temperature != 0.3 {
		t.Errorf("expected default temperature 0.3, got %f", cfg.Dict.LLM.Temperature)
	}
}

func TestDynamicDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `version: v1
dict:
  default: youdao
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

	if cfg.Notebook.Settings.BasePath == "" {
		t.Error("expected basepath to be dynamically set")
	}
	if cfg.Dict.Ecdict.DBFilename == "" {
		t.Error("expected ecdict db_filename to be dynamically set")
	}
}

func TestEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `version: v1
dict:
  default: youdao
  llm:
    api_key: file-key
    url: "https://file-url.com"
    model: file-model
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("WORDFLOW_DICT_LLM_API_KEY", "env-key")
	os.Setenv("WORDFLOW_DICT_DEFAULT", "llm")
	defer os.Unsetenv("WORDFLOW_DICT_LLM_API_KEY")
	defer os.Unsetenv("WORDFLOW_DICT_DEFAULT")

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Dict.LLM.ApiKey != "env-key" {
		t.Errorf("expected env var override, got %q", cfg.Dict.LLM.ApiKey)
	}
	if cfg.Dict.Default != "llm" {
		t.Errorf("expected env var override for default, got %q", cfg.Dict.Default)
	}
	if cfg.Dict.LLM.Model != "file-model" {
		t.Errorf("expected file value for model, got %q", cfg.Dict.LLM.Model)
	}
}

func TestEnvOverrideDuration(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := `version: v1
dict:
  default: youdao
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("WORDFLOW_DICT_LLM_TIMEOUT", "60s")
	defer os.Unsetenv("WORDFLOW_DICT_LLM_TIMEOUT")

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if time.Duration(cfg.Dict.LLM.Timeout) != 60*time.Second {
		t.Errorf("expected timeout 60s from env, got %v", time.Duration(cfg.Dict.LLM.Timeout))
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		configYAML     string
		expectedErr    string
		envVars        map[string]string
		cleanupEnvVars []string
	}{
		{
			name: "youdao no validation needed",
			endpoint: "youdao",
			configYAML: `version: v1
dict:
  default: youdao
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
		},
		{
			name:     "llm missing api_key",
			endpoint: "llm",
			configYAML: `version: v1
dict:
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
			expectedErr: "llm.api_key is required",
		},
		{
			name:     "llm missing url",
			endpoint: "llm",
			configYAML: `version: v1
dict:
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
			expectedErr: "llm.url is required",
		},
		{
			name:     "llm invalid temperature",
			endpoint: "llm",
			configYAML: `version: v1
dict:
  default: llm
  llm:
    api_key: test-key
    url: "https://api.example.com"
    model: test-model
    timeout: 30s
    max_tokens: 2000
    temperature: 5.0
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "llm.temperature must be between 0 and 2",
		},
		{
			name:     "wrong version",
			endpoint: "youdao",
			configYAML: `version: "0.1"
dict:
  default: youdao
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
    new_cards_per_day: 20
`,
			expectedErr: "unsupported config version",
		},
		{
			name:     "llm with env var api_key",
			endpoint: "llm",
			configYAML: `version: v1
dict:
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
			envVars:        map[string]string{"WORDFLOW_DICT_LLM_API_KEY": "env-key"},
			cleanupEnvVars: []string{"WORDFLOW_DICT_LLM_API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for _, k := range tt.cleanupEnvVars {
					os.Unsetenv(k)
				}
			}()

			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(tt.configYAML), 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := LoadConfig(configFile)
			if err != nil {
				t.Fatal(err)
			}

			err = cfg.Validate(tt.endpoint)
			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedErr)
				} else if !contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	if err := InitConfig(configFile); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !contains(content, "version: v1") {
		t.Error("expected version v1 in init config")
	}
	if !contains(content, "default: youdao") {
		t.Error("expected default youdao in init config")
	}
	if !contains(content, "timeout: 30s") {
		t.Error("expected timeout in init config")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}