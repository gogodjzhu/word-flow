package translator

import (
	"testing"

	"github.com/gogodjzhu/word-flow/internal/config"
	trans_baidu "github.com/gogodjzhu/word-flow/pkg/translator/baidu"
	trans_google "github.com/gogodjzhu/word-flow/pkg/translator/google"
	trans_llm "github.com/gogodjzhu/word-flow/pkg/translator/llm"
)

func TestNewTranslator_Google(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "google",
		Google:  &config.TransGoogleConfig{},
	}
	trans, err := NewTranslator(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := trans.(*trans_google.TranslatorGoogle); !ok {
		t.Errorf("expected *TranslatorGoogle, got %T", trans)
	}
}

func TestNewTranslator_GoogleWithConfig(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "google",
		Google:  &config.TransGoogleConfig{},
	}
	trans, err := NewTranslator(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trans == nil {
		t.Error("expected non-nil translator")
	}
}

func TestNewTranslator_LLM(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "llm",
		LLM:     &config.TransLLMConfig{ApiKey: "test", URL: "http://test", Model: "test", MaxTokens: 100, Temperature: 0.3},
	}
	trans, err := NewTranslator(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := trans.(*trans_llm.TranslatorLLM); !ok {
		t.Errorf("expected *TranslatorLLM, got %T", trans)
	}
}

func TestNewTranslator_LLMValidationError(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "llm",
		LLM:     &config.TransLLMConfig{},
	}
	_, err := NewTranslator(cfg)
	if err == nil {
		t.Error("expected validation error for missing api_key, got nil")
	}
}

func TestNewTranslator_UnknownEndpoint(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "unknown",
	}
	_, err := NewTranslator(cfg)
	if err == nil {
		t.Error("expected error for unknown endpoint, got nil")
	}
}

func TestNewTranslator_Baidu(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "baidu",
		Baidu:   &config.TransBaiduConfig{AppID: "test_app_id", Secret: "test_secret"},
	}
	trans, err := NewTranslator(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := trans.(*trans_baidu.TranslatorBaidu); !ok {
		t.Errorf("expected *TranslatorBaidu, got %T", trans)
	}
}

func TestNewTranslator_BaiduValidationError(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "baidu",
		Baidu:   &config.TransBaiduConfig{AppID: "", Secret: ""},
	}
	_, err := NewTranslator(cfg)
	if err == nil {
		t.Error("expected validation error for missing app_id and secret, got nil")
	}
}

func TestNewTranslator_BaiduValidationErrorAppIDOnly(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "baidu",
		Baidu:   &config.TransBaiduConfig{AppID: "test_app_id", Secret: ""},
	}
	_, err := NewTranslator(cfg)
	if err == nil {
		t.Error("expected validation error for missing secret, got nil")
	}
}

func TestNewTranslator_BaiduValidationErrorSecretOnly(t *testing.T) {
	cfg := &config.TransConfig{
		Default: "baidu",
		Baidu:   &config.TransBaiduConfig{AppID: "", Secret: "test_secret"},
	}
	_, err := NewTranslator(cfg)
	if err == nil {
		t.Error("expected validation error for missing app_id, got nil")
	}
}

func TestAvailableTranslators(t *testing.T) {
	translators := AvailableTranslators()
	names := make(map[string]bool)
	for _, ti := range translators {
		names[ti.Name] = true
	}
	if !names["llm"] {
		t.Error("expected 'llm' translator to be available")
	}
	if !names["google"] {
		t.Error("expected 'google' translator to be available")
	}
	if !names["baidu"] {
		t.Error("expected 'baidu' translator to be available")
	}
	if len(translators) != 3 {
		t.Errorf("expected 3 available translators, got %d", len(translators))
	}
}