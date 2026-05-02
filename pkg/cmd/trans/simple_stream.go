package trans

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
)

type SimpleStreamReader struct {
	reader   *bufio.Reader
	renderer *cmdutil.Renderer
	output   io.Writer
	buffer   strings.Builder
}

func NewSimpleStreamReader(reader io.Reader, renderer *cmdutil.Renderer, output io.Writer) *SimpleStreamReader {
	return &SimpleStreamReader{
		reader:   bufio.NewReader(reader),
		renderer: renderer,
		output:   output,
	}
}

func (s *SimpleStreamReader) Process() error {
	var buffer strings.Builder
	lastRenderedCount := 0
	chunkCount := 0

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					buffer.WriteString(line)
				}
				s.tryParseAndRender(buffer.String(), &lastRenderedCount)
				break
			} else {
				return err
			}
		}

		buffer.WriteString(line)
		chunkCount++

		s.tryParseAndRender(buffer.String(), &lastRenderedCount)

		if chunkCount > 1000 && lastRenderedCount == 0 {
			content := buffer.String()
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			_ = s.renderer.RenderToWriter([]cmdutil.MarkupSegment{{
				Text: fmt.Sprintf("[Debug] Received %d chunks, no segments parsed yet. Content preview: %s", chunkCount, content),
				Type: cmdutil.MarkupComment,
			}}, s.output)
		}
	}

	return nil
}

func (s *SimpleStreamReader) tryParseAndRender(content string, lastRenderedCount *int) {
	trimmed := strings.TrimSpace(content)

	if !strings.HasPrefix(trimmed, "[") {
		return
	}

	fixed := strings.TrimSuffix(trimmed, ",")
	if !strings.HasSuffix(fixed, "]") {
		fixed += "]"
	}

	var segments []TranslationSegment
	if json.Unmarshal([]byte(fixed), &segments) != nil {
		if json.Unmarshal([]byte(trimmed), &segments) != nil {
			return
		}
	}

	translationRenderer := NewTranslationRenderer(s.renderer, s.output)

	for i := *lastRenderedCount; i < len(segments); i++ {
		translationSegments := []TranslationSegment{
			{Raw: segments[i].Raw, Translation: segments[i].Translation},
		}

		if i > 0 {
			spacing := []cmdutil.MarkupSegment{{Text: "\n\n", Type: cmdutil.MarkupText}}
			_ = s.renderer.RenderToWriter(spacing, s.output)
		}

		_ = translationRenderer.renderSegments(translationSegments)
		translationRenderer.Flush()
	}

	*lastRenderedCount = len(segments)
}