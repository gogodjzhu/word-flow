package cmdutil

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// MarkupType 定义通用的文本标记类型
type MarkupType int

const (
	MarkupText    MarkupType = iota
	MarkupTitle              // 主要内容/标题 (原红色)
	MarkupRef                // 引用/参考信息 (原绿色)
	MarkupNote               // 注释/说明信息 (原青色)
	MarkupComment            // 补充说明/示例 (原灰色)
)

// MarkupSegment 表示一个带标记的文本片段
type MarkupSegment struct {
	Text string
	Type MarkupType
}

// ColorScheme 定义颜色方案
type ColorScheme struct {
	Title   func(...interface{}) string
	Ref     func(...interface{}) string
	Note    func(...interface{}) string
	Comment func(...interface{}) string
}

// Renderer 负责将标记片段渲染为最终文本
type Renderer struct {
	colorScheme *ColorScheme
	enabled     bool
}

// NewRenderer 创建新的渲染器
func NewRenderer(enabled bool) *Renderer {
	colorScheme := &ColorScheme{
		Title:   color.New(color.FgRed).SprintFunc(),
		Ref:     color.New(color.FgHiGreen).SprintFunc(),
		Note:    color.New(color.FgCyan).SprintFunc(),
		Comment: color.New(color.FgHiBlack).SprintFunc(),
	}

	return &Renderer{
		colorScheme: colorScheme,
		enabled:     enabled,
	}
}

// Render 将标记片段渲染为字符串
func (r *Renderer) Render(segments []MarkupSegment) string {
	var result strings.Builder

	for _, segment := range segments {
		text := segment.Text
		if r.enabled {
			switch segment.Type {
			case MarkupTitle:
				text = r.colorScheme.Title(segment.Text)
			case MarkupRef:
				text = r.colorScheme.Ref(segment.Text)
			case MarkupNote:
				text = r.colorScheme.Note(segment.Text)
			case MarkupComment:
				text = r.colorScheme.Comment(segment.Text)
			}
		}
		result.WriteString(text)
	}

	return result.String()
}

// RenderToWriter 将标记片段渲染并写入指定的 writer
func (r *Renderer) RenderToWriter(segments []MarkupSegment, writer io.Writer) error {
	_, err := fmt.Fprint(writer, r.Render(segments))
	return err
}

// RenderText 渲染纯文本（无标记）
func (r *Renderer) RenderText(text string, writer io.Writer) error {
	segments := []MarkupSegment{{Text: text, Type: MarkupText}}
	return r.RenderToWriter(segments, writer)
}
