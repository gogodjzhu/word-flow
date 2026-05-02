package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPatchYAMLFile_SetLeafValue(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `version: v1
dict:
  # Default dictionary endpoint
  default: youdao
  llm:
    timeout: 30s
    max_tokens: 2000
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "dict.llm.timeout", "60s"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "timeout: 60s") {
		t.Errorf("expected timeout to be 60s, got:\n%s", content)
	}
	if !strings.Contains(content, "# Default dictionary endpoint") {
		t.Errorf("expected comment to be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "max_tokens: 2000") {
		t.Errorf("expected max_tokens unchanged, got:\n%s", content)
	}
	if !strings.Contains(content, "default: youdao") {
		t.Errorf("expected default unchanged, got:\n%s", content)
	}
}

func TestPatchYAMLFile_MissingIntermediateNodes(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `version: v1
dict:
  default: youdao
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "trans.llm.api_key", "sk-xxx"); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}
	applyDefaults(cfg, configFile)

	if cfg.Trans.LLM.ApiKey != "sk-xxx" {
		t.Errorf("expected trans.llm.api_key to be 'sk-xxx', got %q", cfg.Trans.LLM.ApiKey)
	}
}

func TestPatchYAMLFile_TypeCoercion(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		wantIn   string
	}{
		{
			name:   "boolean true",
			key:    "notebook.settings.max_reviews_per_session",
			value:  "100",
			wantIn: "100",
		},
		{
			name:   "string value",
			key:    "dict.default",
			value:  "ecdict",
			wantIn: "ecdict",
		},
		{
			name:   "duration value",
			key:    "dict.llm.timeout",
			value:  "60s",
			wantIn: "60s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			original := `version: v1
dict:
  default: youdao
  llm:
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
`
			if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
				t.Fatal(err)
			}

			if err := PatchYAMLFile(configFile, tt.key, tt.value); err != nil {
				t.Fatal(err)
			}

			data, err := os.ReadFile(configFile)
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			if !strings.Contains(content, tt.wantIn) {
				t.Errorf("expected %q in output, got:\n%s", tt.wantIn, content)
			}
		})
	}
}

func TestPatchYAMLFile_BooleanAndFloatTypes(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	original := `version: v1
dict:
  default: youdao
  llm:
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "dict.llm.temperature", "0.7"); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}
	applyDefaults(cfg, configFile)

	if cfg.Dict.LLM.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", cfg.Dict.LLM.Temperature)
	}
}

func TestPatchYAMLFile_IntType(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	original := `version: v1
notebook:
  default: default
  settings:
    max_reviews_per_session: 50
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "notebook.settings.max_reviews_per_session", "100"); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}
	applyDefaults(cfg, configFile)

	if cfg.Notebook.Settings.MaxReviews != 100 {
		t.Errorf("expected max_reviews 100, got %d", cfg.Notebook.Settings.MaxReviews)
	}
}

func TestPatchYAMLFile_DurationType(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	original := `version: v1
dict:
  default: youdao
  llm:
    timeout: 30s
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "dict.llm.timeout", "60s"); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}
	applyDefaults(cfg, configFile)

	if time.Duration(cfg.Dict.LLM.Timeout) != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", time.Duration(cfg.Dict.LLM.Timeout))
	}
}

func TestPatchYAMLFile_CommentedSectionSurvives(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `version: v1
dict:
  default: youdao
  llm:
    # timeout for LLM calls
    timeout: 30s
trans:
  default: google
  google: {}
  # llm:
  #   api_key: ""
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "dict.llm.timeout", "60s"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "timeout: 60s") {
		t.Errorf("expected timeout updated to 60s")
	}
	if !strings.Contains(content, "# timeout for LLM calls") {
		t.Errorf("expected comment preserved")
	}
	if !strings.Contains(content, "# llm:") {
		t.Errorf("expected commented-out llm section preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "#   api_key:") {
		t.Errorf("expected commented-out api_key preserved, got:\n%s", content)
	}
}

func TestPatchYAMLFile_CreateNextToCommentedSection(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `version: v1
trans:
  default: google
  google: {}
  # llm:
  #   api_key: ""
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "trans.llm.api_key", "sk-test"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "api_key: sk-test") {
		t.Errorf("expected api_key to be set, got:\n%s", content)
	}
	if !strings.Contains(content, "# llm:") {
		t.Errorf("expected commented-out llm to be preserved, got:\n%s", content)
	}

	cfg, err := ReadConfigSpecified(configFile)
	if err != nil {
		t.Fatal(err)
	}
	applyDefaults(cfg, configFile)

	if cfg.Trans.LLM.ApiKey != "sk-test" {
		t.Errorf("expected trans.llm.api_key to be 'sk-test', got %q", cfg.Trans.LLM.ApiKey)
	}
}

func TestPatchYAMLFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `: invalid yaml {{`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	err := PatchYAMLFile(configFile, "dict.default", "ecdict")
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestPatchYAMLFile_FileNotExist(t *testing.T) {
	err := PatchYAMLFile("/nonexistent/path/config.yaml", "dict.default", "ecdict")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestPatchYAMLFile_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `version: v1
dict:
  default: youdao
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	err := PatchYAMLFile(configFile, "nonexistent.field", "value")
	if err == nil {
		t.Error("expected error for invalid key, got nil")
	}
}

func TestPatchYAMLFile_PreservesOtherFieldsAndComments(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	original := `# Wordflow configuration
version: v1

dict:
  # Dictionary endpoint options: youdao, llm, ecdict
  default: youdao

  llm:
    # Required for LLM dictionary
    # api_key: ""
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3

trans:
  default: google
  google: {}
`
	if err := os.WriteFile(configFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PatchYAMLFile(configFile, "dict.llm.timeout", "60s"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "# Wordflow configuration") {
		t.Errorf("expected top comment preserved")
	}
	if !strings.Contains(content, "# Dictionary endpoint options: youdao, llm, ecdict") {
		t.Errorf("expected dict comment preserved")
	}
	if !strings.Contains(content, "# Required for LLM dictionary") {
		t.Errorf("expected llm comment preserved")
	}
	if !strings.Contains(content, "# api_key: \"\"") {
		t.Errorf("expected commented api_key preserved, got:\n%s", content)
	}
	if !strings.Contains(content, "timeout: 60s") {
		t.Errorf("expected timeout updated to 60s")
	}
	if !strings.Contains(content, "version: v1") {
		t.Errorf("expected version preserved")
	}
	if !strings.Contains(content, "max_tokens: 2000") {
		t.Errorf("expected max_tokens preserved")
	}
	if !strings.Contains(content, "temperature: 0.3") {
		t.Errorf("expected temperature preserved")
	}
}