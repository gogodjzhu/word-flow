package util

import (
	"reflect"
	"testing"
)

func TestBatchSegmentsSingleBatch(t *testing.T) {
	segments := []string{"ab", "cd", "ef"}
	result := BatchSegments(segments, 100)
	expected := [][]string{{"ab", "cd", "ef"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestBatchSegmentsMultipleBatches(t *testing.T) {
	segments := []string{"ab", "cd", "ef"}
	result := BatchSegments(segments, 4)
	expected := [][]string{{"ab", "cd"}, {"ef"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestBatchSegmentsSingleExceedingThreshold(t *testing.T) {
	segments := []string{"abcdef", "gh"}
	result := BatchSegments(segments, 4)
	expected := [][]string{{"abcdef"}, {"gh"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestBatchSegmentsEmpty(t *testing.T) {
	result := BatchSegments(nil, 10)
	if result != nil {
		t.Errorf("got %v, want nil", result)
	}
}