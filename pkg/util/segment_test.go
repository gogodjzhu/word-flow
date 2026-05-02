package util

import (
	"reflect"
	"testing"
)

func TestSegmentTextMultipleSentences(t *testing.T) {
	result := SegmentText("Hello world. How are you? I am fine.")
	expected := []string{"Hello world.", " How are you?", " I am fine."}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestSegmentTextSingleSegment(t *testing.T) {
	result := SegmentText("hello world")
	expected := []string{"hello world"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestSegmentTextAbbreviation(t *testing.T) {
	result := SegmentText("Mr. Smith went to the U.S. capital. He arrived.")
	expected := []string{"Mr. Smith went to the U.S. capital.", " He arrived."}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestSegmentTextCommonAbbreviations(t *testing.T) {
	abbrTests := []string{"Dr.", "e.g.", "i.e.", "vs.", "etc.", "Inc."}
	for _, abbr := range abbrTests {
		result := SegmentText(abbr + " something.")
		if len(result) != 1 {
			t.Errorf("abbreviation %q should not split, got %v", abbr, result)
		}
	}
}

func TestSegmentTextTrailingPunctuation(t *testing.T) {
	result := SegmentText("First. Second.")
	expected := []string{"First.", " Second."}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}