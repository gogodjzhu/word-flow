package cmdtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
)

func TestConfigViewAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `version: v1
dict:
  default: llm
  youdao: {}
  llm:
    api_key: my-key
    url: "https://api.example.com"
    model: gpt-4
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

	cfg, err := config.ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}

	// Test ToString (used by config view)
	out, err := cfg.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "default: llm") {
		t.Error("expected 'default: llm' in view output")
	}
	if !strings.Contains(out, "api_key: my-key") {
		t.Error("expected 'api_key: my-key' in view output")
	}

	// Test config path
	if cfg.Common.ConfigFilename != configFile {
		t.Errorf("expected config path %q, got %q", configFile, cfg.Common.ConfigFilename)
	}
}

func TestConfigGetWithEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	content := `version: v1
dict:
  default: youdao
  youdao: {}
  llm:
    api_key: file-key
    url: "https://api.example.com"
    model: gpt-4
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

	os.Setenv("WORDFLOW_DICT_LLM_API_KEY", "env-key")
	defer os.Unsetenv("WORDFLOW_DICT_LLM_API_KEY")

	cfg, err := config.ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Dict.LLM.ApiKey != "env-key" {
		t.Errorf("expected env override 'env-key', got %q", cfg.Dict.LLM.ApiKey)
	}
}

func TestConfigInitTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	if err := config.InitConfig(configFile); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "version: v1") {
		t.Error("expected version v1 in template")
	}
	if !strings.Contains(content, "timeout: 30s") {
		t.Error("expected timeout in template")
	}
	if !strings.Contains(content, "default: youdao") {
		t.Error("expected default youdao in template")
	}
	if !strings.Contains(content, "Full API endpoint URL") {
		t.Error("expected URL description mentioning 'Full API endpoint URL' in template")
	}
}

func TestConfigInitForce(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Write old config
	if err := os.WriteFile(configFile, []byte("version: \"0.1\"\nold: stuff\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Force overwrite with template
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte(config.ConfigTemplate()), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: v1") {
		t.Error("expected v1 config after force overwrite")
	}
}

func TestGlobalConfigFlag(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "custom.yaml")
	content := `version: v1
dict:
  default: ecdict
  ecdict:
    db_filename: /tmp/test.db
  youdao: {}
  etymonline: {}
  mwebster: {}
  llm:
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: my-notebook
  settings:
    max_reviews_per_session: 30
    new_cards_per_day: 10
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	f := cmdutil.NewFactory()
	f.SetConfigPath(configFile)

	cfg, err := f.Config()
	if err != nil {
		t.Fatalf("failed to load config from custom path: %v", err)
	}
	if cfg.Dict.Default != "ecdict" {
		t.Errorf("expected default 'ecdict', got %q", cfg.Dict.Default)
	}
	if cfg.Notebook.Default != "my-notebook" {
		t.Errorf("expected notebook default 'my-notebook', got %q", cfg.Notebook.Default)
	}
	if cfg.Notebook.Settings.MaxReviews != 30 {
		t.Errorf("expected max_reviews 30, got %d", cfg.Notebook.Settings.MaxReviews)
	}
	if cfg.Common.ConfigFilename != configFile {
		t.Errorf("expected config path %q, got %q", configFile, cfg.Common.ConfigFilename)
	}

	// Second call should return cached config (same pointer)
	cfg2, err := f.Config()
	if err != nil {
		t.Fatal(err)
	}
	if cfg != cfg2 {
		t.Error("expected cached config to be returned on second call")
	}
	_ = cfg2 // verify cache returns same pointer
}