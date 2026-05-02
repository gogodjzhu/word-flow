package dict_google

import (
	"encoding/json"
	"testing"
)

func TestAbbreviatePos(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"adjective", "adj."},
		{"noun", "n."},
		{"verb", "v."},
		{"adverb", "adv."},
		{"preposition", "prep."},
		{"conjunction", "conj."},
		{"pronoun", "pron."},
		{"interjection", "int."},
		{"unknown", "unknown"},
		{"", ""},
		{"  Noun  ", "n."},
	}
	for _, tt := range tests {
		result := abbreviatePos(tt.input)
		if result != tt.expected {
			t.Errorf("abbreviatePos(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseDictResponse_WithBdData(t *testing.T) {
	response := []interface{}{
		nil,
		[]interface{}{
			[]interface{}{"noun", []interface{}{"名词", "事物"}},
			[]interface{}{"verb", []interface{}{"行动"}},
		},
		"en",
	}
	body, _ := json.Marshal(response)
	wordItem, err := parseDictResponse(body, "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wordItem.Word != "test" {
		t.Errorf("expected word 'test', got %q", wordItem.Word)
	}
	if len(wordItem.WordMeanings) != 2 {
		t.Fatalf("expected 2 meanings, got %d", len(wordItem.WordMeanings))
	}
	if wordItem.WordMeanings[0].PartOfSpeech != "n." {
		t.Errorf("expected part of speech 'n.', got %q", wordItem.WordMeanings[0].PartOfSpeech)
	}
	if wordItem.WordMeanings[0].Definitions != "名词; 事物" {
		t.Errorf("expected definitions '名词; 事物', got %q", wordItem.WordMeanings[0].Definitions)
	}
	if wordItem.WordMeanings[1].PartOfSpeech != "v." {
		t.Errorf("expected part of speech 'v.', got %q", wordItem.WordMeanings[1].PartOfSpeech)
	}
	if wordItem.WordMeanings[1].Definitions != "行动" {
		t.Errorf("expected definitions '行动', got %q", wordItem.WordMeanings[1].Definitions)
	}
}

func TestParseDictResponse_AdjectiveAbbreviation(t *testing.T) {
	response := []interface{}{
		nil,
		[]interface{}{
			[]interface{}{"adjective", []interface{}{"形容词"}},
		},
	}
	body, _ := json.Marshal(response)
	wordItem, err := parseDictResponse(body, "big")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(wordItem.WordMeanings) != 1 {
		t.Fatalf("expected 1 meaning, got %d", len(wordItem.WordMeanings))
	}
	if wordItem.WordMeanings[0].PartOfSpeech != "adj." {
		t.Errorf("expected part of speech 'adj.', got %q", wordItem.WordMeanings[0].PartOfSpeech)
	}
}

func TestParseDictResponse_FallbackTranslation(t *testing.T) {
	response := []interface{}{
		[]interface{}{
			[]interface{}{"你好", "hello", nil, nil, float64(10)},
		},
		nil,
		"en",
	}
	body, _ := json.Marshal(response)
	wordItem, err := parseDictResponse(body, "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(wordItem.WordMeanings) != 1 {
		t.Fatalf("expected 1 meaning, got %d", len(wordItem.WordMeanings))
	}
	if wordItem.WordMeanings[0].PartOfSpeech != "" {
		t.Errorf("expected empty part of speech, got %q", wordItem.WordMeanings[0].PartOfSpeech)
	}
	if wordItem.WordMeanings[0].Definitions != "你好" {
		t.Errorf("expected definitions '你好', got %q", wordItem.WordMeanings[0].Definitions)
	}
}

func TestParseDictResponse_EmptyBody(t *testing.T) {
	_, err := parseDictResponse([]byte{}, "test")
	if err == nil {
		t.Error("expected error for empty body, got nil")
	}
}

func TestParseDictResponse_InvalidJSON(t *testing.T) {
	_, err := parseDictResponse([]byte(`not json`), "test")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}