package cmdutil

import (
	"io"
)

// RenderHelper provides utility methods for common rendering patterns
type RenderHelper struct {
	renderer *Renderer
	output   io.Writer
}

func NewRenderHelper(renderer *Renderer, output io.Writer) *RenderHelper {
	return &RenderHelper{
		renderer: renderer,
		output:   output,
	}
}

// RenderWithSpacing renders content with spacing before and after
func (rh *RenderHelper) RenderWithSpacing(content string, before, after string) error {
	if before != "" {
		beforeSegments := []MarkupSegment{{Text: before, Type: MarkupText}}
		if err := rh.renderer.RenderToWriter(beforeSegments, rh.output); err != nil {
			return err
		}
	}

	contentSegments := []MarkupSegment{{Text: content, Type: MarkupText}}
	if err := rh.renderer.RenderToWriter(contentSegments, rh.output); err != nil {
		return err
	}

	if after != "" {
		afterSegments := []MarkupSegment{{Text: after, Type: MarkupText}}
		if err := rh.renderer.RenderToWriter(afterSegments, rh.output); err != nil {
			return err
		}
	}
	return nil
}

// RenderComment renders text as a comment
func (rh *RenderHelper) RenderComment(content string) error {
	commentSegments := []MarkupSegment{{Text: content, Type: MarkupComment}}
	return rh.renderer.RenderToWriter(commentSegments, rh.output)
}

// Flush output if possible
func (rh *RenderHelper) Flush() {
	if flusher, ok := rh.output.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}
}
