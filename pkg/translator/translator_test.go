package translator

import (
	"testing"

	"github.com/gogodjzhu/word-flow/internal/config"
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
	if len(translators) != 2 {
		t.Errorf("expected 2 available translators, got %d", len(translators))
	}
}