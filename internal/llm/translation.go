package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// PromptBuilder constructs translation prompts
type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (pb *PromptBuilder) BuildTranslationPrompt(ref bool) string {
	basePrompt := "你是一个专业的文本翻译工具。请将以下英文文本翻译成中文。"

	if ref {
		return basePrompt + `
要求：
1. 请将原文按段落合理分段，每段包含完整的意思
2. 输出格式必须是 JSON 数组，每个元素包含 'raw' 和 'translation' 字段
3. 'raw' 字段包含原文段落，'translation' 字段包含对应的中文翻译
4. 保持段落完整性，不要在句子中间断开
5. 只输出有效的 JSON 数组，不要包含任何其他内容
6. 示例格式：[{"raw": "原文段落1", "translation": "翻译段落1"}, {"raw": "原文段落2", "translation": "翻译段落2"}]
7. 重要：请按顺序输出完整的 JSON 数组，确保格式正确`
	}

	return basePrompt + `
要求：
1. 保持原始格式，包括换行、空格、制表符
2. 准确翻译内容，不要添加额外解释
3. 不要改变文本结构，只进行语言转换
4. 直接输出翻译结果，不要包含任何其他内容`
}

// StreamProcessor handles streaming translation responses
type StreamProcessor struct {
	output io.Writer
}

func NewStreamProcessor(output io.Writer) *StreamProcessor {
	return &StreamProcessor{output: output}
}

func (sp *StreamProcessor) ProcessStream(response *http.Response) (string, error) {
	defer response.Body.Close()

	if err := sp.validateResponse(response); err != nil {
		return "", err
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(response.Body)

	for scanner.Scan() {
		done, err := sp.processStreamLine(scanner.Text(), &fullContent)
		if err != nil {
			continue
		}
		if done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "failed to read streaming response")
	}

	return fullContent.String(), nil
}

func (sp *StreamProcessor) validateResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("LLM API error: status %d", response.StatusCode))
	}
	return nil
}

func (sp *StreamProcessor) processStreamLine(line string, fullContent *strings.Builder) (bool, error) {
	if !strings.HasPrefix(line, "data: ") {
		return false, nil
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "[DONE]" {
		return true, nil
	}

	var streamResp StreamResponse
	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return false, err
	}

	if len(streamResp.Choices) == 0 {
		return false, nil
	}

	delta := streamResp.Choices[0].Delta
	if delta.Content != "" {
		if sp.output != nil {
			fmt.Fprint(sp.output, delta.Content)
		} else {
			fmt.Print(delta.Content)
		}
		fullContent.WriteString(delta.Content)
	}

	if streamResp.Choices[0].FinishReason != nil {
		if sp.output != nil {
			fmt.Fprintln(sp.output)
		} else {
			fmt.Println()
		}
		return true, nil
	}

	return false, nil
}

type ResponseProcessor struct{}

func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{}
}

func (rp *ResponseProcessor) ProcessResponse(response *http.Response) (string, error) {
	defer response.Body.Close()

	if err := rp.validateResponse(response); err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	chatResponse, err := rp.parseResponse(bodyBytes)
	if err != nil {
		return "", err
	}

	if err := rp.validateChatResponse(chatResponse); err != nil {
		return "", err
	}

	return chatResponse.Choices[0].Message.Content, nil
}

func (rp *ResponseProcessor) validateResponse(response *http.Response) error {
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("LLM API error: status %d", response.StatusCode))
	}
	return nil
}

func (rp *ResponseProcessor) parseResponse(bodyBytes []byte) (*ChatResponse, error) {
	var chatResponse ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResponse); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal chat response")
	}
	return &chatResponse, nil
}

func (rp *ResponseProcessor) validateChatResponse(chatResponse *ChatResponse) error {
	if len(chatResponse.Choices) == 0 {
		return errors.New("no choices in LLM response")
	}
	return nil
}
