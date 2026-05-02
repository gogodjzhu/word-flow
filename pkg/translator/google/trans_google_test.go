package google

import (
	"encoding/json"
	"testing"
)

func TestParseTranslateResponse_ValidResponse(t *testing.T) {
	inner := []interface{}{
		[]interface{}{
			[]interface{}{"你好", "hello", nil, nil, float64(10)},
		},
		nil,
		"en",
	}
	body, _ := json.Marshal(inner)
	segments, err := parseTranslateResponse(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].translation != "你好" {
		t.Errorf("expected translation '你好', got %q", segments[0].translation)
	}
	if segments[0].original != "hello" {
		t.Errorf("expected original 'hello', got %q", segments[0].original)
	}
}

func TestParseTranslateResponse_MultipleSegments(t *testing.T) {
	response := [][]interface{}{
		{"世界", "world", nil, nil, float64(10)},
		{"好的", "good", nil, nil, float64(10)},
	}
	inner := []interface{}{response, nil, "en"}
	body, _ := json.Marshal(inner)
	segments, err := parseTranslateResponse(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
	if segments[0].translation != "世界" {
		t.Errorf("expected translation '世界', got %q", segments[0].translation)
	}
	if segments[0].original != "world" {
		t.Errorf("expected original 'world', got %q", segments[0].original)
	}
	if segments[1].translation != "好的" {
		t.Errorf("expected translation '好的', got %q", segments[1].translation)
	}
	if segments[1].original != "good" {
		t.Errorf("expected original 'good', got %q", segments[1].original)
	}
}

func TestParseTranslateResponse_EmptyBody(t *testing.T) {
	segments, err := parseTranslateResponse([]byte{})
	if err != nil {
		t.Fatalf("expected no error for empty body, got %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for empty body, got %d", len(segments))
	}
}

func TestParseTranslateResponse_EmptyArray(t *testing.T) {
	segments, err := parseTranslateResponse([]byte(`[]`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for empty array, got %d", len(segments))
	}
}

func TestParseTranslateResponse_InvalidJSON(t *testing.T) {
	_, err := parseTranslateResponse([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseTranslateResponse_ShortSegment(t *testing.T) {
	inner := []interface{}{
		[]interface{}{"only translation"},
		nil,
		"en",
	}
	body, _ := json.Marshal(inner)
	segments, err := parseTranslateResponse(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for short segment array, got %d", len(segments))
	}
}