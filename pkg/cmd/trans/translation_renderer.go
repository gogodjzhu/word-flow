package trans

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
)

// TranslationSegment represents a translation segment pair for --ref mode
type TranslationSegment struct {
	Raw         string `json:"raw"`
	Translation string `json:"translation"`
}

// TranslationRenderer handles translation-specific rendering logic
type TranslationRenderer struct {
	renderer *cmdutil.Renderer
	output   io.Writer
}

func NewTranslationRenderer(renderer *cmdutil.Renderer, output io.Writer) *TranslationRenderer {
	return &TranslationRenderer{
		renderer: renderer,
		output:   output,
	}
}

// RenderTranslationWithRef parses JSON translation and renders with --ref format
func (tr *TranslationRenderer) RenderTranslationWithRef(translation string) error {
	translation = strings.TrimSpace(translation)
	if strings.HasPrefix(translation, "[") && strings.HasSuffix(translation, "]") {
		segments, err := tr.parseTranslationSegments(translation)
		if err != nil {
			return tr.renderer.RenderText(translation, tr.output)
		}
		return tr.renderSegments(segments)
	}
	return tr.renderer.RenderText(translation, tr.output)
}

// parseTranslationSegments parses JSON string into TranslationSegment slice
func (tr *TranslationRenderer) parseTranslationSegments(translation string) ([]TranslationSegment, error) {
	var segments []TranslationSegment
	if err := json.Unmarshal([]byte(translation), &segments); err != nil {
		return nil, err
	}
	return segments, nil
}

// renderSegments converts TranslationSegments to MarkupSegments and renders them
func (tr *TranslationRenderer) renderSegments(segments []TranslationSegment) error {
	for i, segment := range segments {
		if i > 0 {
			segmentSpacing := []cmdutil.MarkupSegment{{Text: "\n\n", Type: cmdutil.MarkupText}}
			if err := tr.renderer.RenderToWriter(segmentSpacing, tr.output); err != nil {
				return err
			}
		}

		if segment.Raw != "" {
			rawSegments := []cmdutil.MarkupSegment{{Text: segment.Raw, Type: cmdutil.MarkupText}}
			if err := tr.renderer.RenderToWriter(rawSegments, tr.output); err != nil {
				return err
			}
		}

		spacing := []cmdutil.MarkupSegment{{Text: "\n", Type: cmdutil.MarkupText}}
		if err := tr.renderer.RenderToWriter(spacing, tr.output); err != nil {
			return err
		}

		if segment.Translation != "" {
			translationSegments := []cmdutil.MarkupSegment{{Text: segment.Translation, Type: cmdutil.MarkupComment}}
			if err := tr.renderer.RenderToWriter(translationSegments, tr.output); err != nil {
				return err
			}
		}
	}
	return nil
}

// Flush output if possible
func (tr *TranslationRenderer) Flush() {
	if flusher, ok := tr.output.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}
}
