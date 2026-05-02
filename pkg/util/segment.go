package util

import (
	"regexp"
	"strings"
)

var abbreviations = map[string]bool{
	"mr": true, "mrs": true, "ms": true, "dr": true, "prof": true,
	"rev": true, "gen": true, "rep": true, "sen": true, "st": true,
	"jr": true, "sr": true, "inc": true, "corp": true, "ltd": true,
	"co": true, "vs": true, "etc": true, "e.g": true, "i.e": true,
	"a.m": true, "p.m": true, "u.s": true, "u.k": true,
	"approx": true, "dept": true, "est": true, "gov": true,
	"misc": true, "tech": true, "vol": true, "no": true,
}

var sentenceEndRe = regexp.MustCompile(`[.!?]\s+`)

func SegmentText(text string) []string {
	if text == "" {
		return nil
	}

	matches := sentenceEndRe.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return []string{text}
	}

	var splitPoints []int
	for _, loc := range matches {
		punctIdx := loc[0]
		word := extractWordBefore(text, punctIdx)
		if abbreviations[strings.ToLower(word)] {
			continue
		}
		splitPoints = append(splitPoints, punctIdx+1)
	}

	if len(splitPoints) == 0 {
		return []string{text}
	}

	var result []string
	prev := 0
	for _, sp := range splitPoints {
		result = append(result, text[prev:sp])
		prev = sp
	}
	if prev < len(text) {
		result = append(result, text[prev:])
	}
	return result
}

func extractWordBefore(text string, punctIdx int) string {
	i := punctIdx - 1
	for i >= 0 {
		ch := text[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '.' {
			i--
		} else {
			break
		}
	}
	return text[i+1 : punctIdx]
}