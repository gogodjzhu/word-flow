package entity

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/fatih/color"
	"strings"
)

type WordItem struct {
	ID            string          `json:"id" yaml:"id"`
	Word          string          `json:"word" yaml:"word"`
	Origin        string          `json:"origin" yaml:"-"`
	WordPhonetics []*WordPhonetic `json:"word_phonetics" yaml:"-"`
	WordMeanings  []*WordMeaning  `json:"word_meanings" yaml:"-"`
	Usages        []string        `json:"usages" yaml:"usages"`
}

type WordPhonetic struct {
	HeadWord     string `json:"head_word" yaml:"headword"`
	LanguageCode string `json:"language_code" yaml:"language_code"`
	Text         string `json:"text" yaml:"text"`
	Audio        string `json:"audio" yaml:"audio"`
}

type WordMeaning struct {
	PartOfSpeech string   `json:"part_of_speech" yaml:"part_of_speech"`
	Definitions  string   `json:"definitions" yaml:"definitions"`
	Examples     []string `json:"examples" yaml:"examples"`
}

type WordNote struct {
	WordItemId     string `json:"word_id" yaml:"word_id"`
	Word           string `json:"word" yaml:"word"`
	LookupTimes    int    `json:"lookup_times" yaml:"lookup_times"`
	CreateTime     int64  `json:"create_time" yaml:"create_time"`
	LastLookupTime int64  `json:"last_lookup_time" yaml:"last_lookup_time"`
}

func WordId(word string) string {
	hash := md5.Sum([]byte(word))      // Compute the MD5 hash of the string
	return hex.EncodeToString(hash[:]) // Convert the hash to a hex string
}

func (f *WordItem) RenderString() string {
	red := color.New(color.FgRed).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgHiGreen).SprintFunc()

	var str string
	str += red(f.Word)
	str += "\n"
	if len(f.WordPhonetics) > 0 {
		for _, phonetic := range f.WordPhonetics {
			if len(strings.TrimSpace(phonetic.Text)) > 0 {
				str += green(phonetic.LanguageCode) + " [" + phonetic.Text + "] "
			}
		}
		str += "\n"
		for _, phonetic := range f.WordPhonetics {
			if len(strings.TrimSpace(phonetic.Audio)) > 0 {
				str += gray(phonetic.Audio) + "\n"
			}
		}
	}
	for _, meaning := range f.WordMeanings {
		if len(meaning.PartOfSpeech) > 0 {
			str += cyan(meaning.PartOfSpeech)
		}
		if len(meaning.Definitions) > 0 {
			str += meaning.Definitions
		}
		for _, s := range meaning.Examples {
			str += "\n" + gray("eg. "+s)
		}
		str += "\n"
	}
	for _, usage := range f.Usages {
		str += gray("eg. "+usage) + "\n"
	}
	return str
}

func (f *WordItem) RawString() string {
	var str string
	for _, meaning := range f.WordMeanings {
		if len(meaning.PartOfSpeech) > 0 {
			str += meaning.PartOfSpeech
		}
		if len(meaning.Definitions) > 0 {
			str += meaning.Definitions + ";"
		}
	}
	return str
}
