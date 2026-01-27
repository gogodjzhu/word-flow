package entity

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
)

type WordItem struct {
	ID            string          `json:"id" yaml:"id"`
	Word          string          `json:"word" yaml:"word"`
	Source        string          `json:"source" yaml:"source"`
	WordPhonetics []*WordPhonetic `json:"word_phonetics" yaml:"-"`
	WordMeanings  []*WordMeaning  `json:"word_meanings" yaml:"-"`
	// mixed examples
	Examples []string `json:"examples,omitempty" yaml:"examples,omitempty"`
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
	Translation    string `json:"translation,omitempty" yaml:"translation,omitempty"`
	// FSRS fields
	FSRSCard   *FSRSCard `json:"fsrs_card,omitempty" yaml:"-"`
	LastRating int       `json:"last_rating,omitempty" yaml:"last_rating"`
	NextReview int64     `json:"next_review" yaml:"next_review"`
}

func WordId(word string) string {
	hash := md5.Sum([]byte(word))      // Compute the MD5 hash of the string
	return hex.EncodeToString(hash[:]) // Convert the hash to a hex string
}

// Format 返回标记化的文本片段，用于后续渲染
func (w *WordItem) Format() []cmdutil.MarkupSegment {
	var segments []cmdutil.MarkupSegment

	// 单词作为标题
	segments = append(segments, cmdutil.MarkupSegment{
		Text: w.Word,
		Type: cmdutil.MarkupTitle,
	})
	if len(w.Source) > 0 {
		segments = append(segments, cmdutil.MarkupSegment{
			Text: "  (" + w.Source + ")",
			Type: cmdutil.MarkupNote,
		})
	}
	segments = append(segments, cmdutil.MarkupSegment{
		Text: "\n",
		Type: cmdutil.MarkupText,
	})

	// 音标信息
	if len(w.WordPhonetics) > 0 {
		for _, phonetic := range w.WordPhonetics {
			if len(strings.TrimSpace(phonetic.Text)) > 0 {
				segments = append(segments, cmdutil.MarkupSegment{
					Text: phonetic.LanguageCode + " [" + phonetic.Text + "] ",
					Type: cmdutil.MarkupRef,
				})
			}
		}
		segments = append(segments, cmdutil.MarkupSegment{
			Text: "\n",
			Type: cmdutil.MarkupText,
		})

		for _, phonetic := range w.WordPhonetics {
			if len(strings.TrimSpace(phonetic.Audio)) > 0 {
				segments = append(segments, cmdutil.MarkupSegment{
					Text: phonetic.Audio + "\n",
					Type: cmdutil.MarkupComment,
				})
			}
		}
	}

	// 词义信息
	for _, meaning := range w.WordMeanings {
		if len(meaning.PartOfSpeech) > 0 {
			segments = append(segments, cmdutil.MarkupSegment{
				Text: meaning.PartOfSpeech,
				Type: cmdutil.MarkupNote,
			})
		}
		if len(meaning.Definitions) > 0 {
			segments = append(segments, cmdutil.MarkupSegment{
				Text: meaning.Definitions,
				Type: cmdutil.MarkupText,
			})
		}
		for _, s := range meaning.Examples {
			segments = append(segments, cmdutil.MarkupSegment{
				Text: "\n",
				Type: cmdutil.MarkupText,
			})
			segments = append(segments, cmdutil.MarkupSegment{
				Text: "eg. " + s,
				Type: cmdutil.MarkupComment,
			})
		}
		segments = append(segments, cmdutil.MarkupSegment{
			Text: "\n",
			Type: cmdutil.MarkupText,
		})
	}
	// 例句信息
	if len(w.Examples) > 0 {
		for _, example := range w.Examples {
			segments = append(segments, cmdutil.MarkupSegment{
				Text: "\n",
				Type: cmdutil.MarkupText,
			})
			segments = append(segments, cmdutil.MarkupSegment{
				Text: "eg. " + example,
				Type: cmdutil.MarkupComment,
			})
		}
	}
	// 结尾换行
	segments = append(segments, cmdutil.MarkupSegment{
		Text: "\n",
		Type: cmdutil.MarkupText,
	})

	return segments
}

func (f *WordItem) RenderString() string {
	// 使用默认渲染器保持向后兼容
	defaultRenderer := cmdutil.NewRenderer(true)
	return defaultRenderer.Render(f.Format())
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

// IsDueForReview checks if the word note is due for FSRS review
func (n *WordNote) IsDueForReview() bool {
	if n.FSRSCard == nil {
		return true // New words are always due
	}
	return time.Now().Unix() >= n.NextReview
}

// GetDefinition returns the primary definition from translation
func (n *WordNote) GetDefinition() string {
	if n.Translation != "" {
		return n.Translation
	}
	return "No definition available"
}

// GetExamples returns example sentences (placeholder for now)
func (n *WordNote) GetExamples() []string {
	// TODO: Extract examples from translation or WordItem
	return []string{}
}
