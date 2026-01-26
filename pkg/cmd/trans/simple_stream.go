package trans

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gogodjzhu/word-flow/internal/llm"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
)

// SimpleStreamReader handles streaming JSON by reading complete response
type SimpleStreamReader struct {
	reader   *bufio.Reader
	renderer *cmdutil.Renderer
	output   io.Writer
	buffer   strings.Builder
}

// NewSimpleStreamReader creates a new simple stream reader
func NewSimpleStreamReader(reader io.Reader, renderer *cmdutil.Renderer, output io.Writer) *SimpleStreamReader {
	return &SimpleStreamReader{
		reader:   bufio.NewReader(reader),
		renderer: renderer,
		output:   output,
	}
}

// Process reads and processes the streaming response with real-time rendering
func (s *SimpleStreamReader) Process() error {
	var buffer strings.Builder
	lastRenderedCount := 0
	chunkCount := 0

	// Read the streaming response
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					buffer.WriteString(line)
				}
				// Try final parse at EOF
				s.tryParseAndRender(buffer.String(), &lastRenderedCount)
				break
			} else {
				return err
			}
		}

		buffer.WriteString(line)
		chunkCount++

		// Try to parse and render after each chunk
		s.tryParseAndRender(buffer.String(), &lastRenderedCount)

		// Safety check: if we've received too many chunks without parsing,
		// there might be an issue with the response format
		if chunkCount > 1000 && lastRenderedCount == 0 {
			// Try to output raw content for debugging
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

	var segments []llm.SegmentPair
	if json.Unmarshal([]byte(fixed), &segments) != nil {
		if json.Unmarshal([]byte(trimmed), &segments) != nil {
			return
		}
	}

	translationRenderer := NewTranslationRenderer(s.renderer, s.output)

	for i := *lastRenderedCount; i < len(segments); i++ {
		llmSegment := segments[i]
		translationSegments := []TranslationSegment{
			{Raw: llmSegment.Raw, Translation: llmSegment.Translation},
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
