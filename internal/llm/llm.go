package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
)

type Client struct {
	apiKey      string
	url         string
	model       string
	timeout     time.Duration
	maxTokens   int
	temperature float64
}

type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Tools       []Tool        `json:"tools"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Finish  string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamResponse struct {
	ID      string         `json:"id"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
}

type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason,omitempty"`
}

type StreamDelta struct {
	Role             string `json:"role,omitempty"`
	Content          string `json:"content,omitempty"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type TranslateResponse struct {
	InputType    string       `json:"input_type"`
	OriginalText string       `json:"original_text"`
	Phonetics    []Phonetic   `json:"phonetics,omitempty"`
	Meanings     []LLMMeaning `json:"meanings"`
}

type SegmentPair struct {
	Raw         string `json:"raw"`
	Translation string `json:"translation"`
}

type SegmentedTranslationResponse struct {
	Segments []SegmentPair `json:"segments"`
}

type Phonetic struct {
	HeadWord     string `json:"head_word"`
	LanguageCode string `json:"language_code"`
	Text         string `json:"text"`
}

type LLMMeaning struct {
	PartOfSpeech string   `json:"part_of_speech"`
	Definitions  string   `json:"definitions"`
	Examples     []string `json:"examples,omitempty"`
}

func NewClient(apiKey, url, model string, timeout time.Duration, maxTokens int, temperature float64) *Client {
	return &Client{
		apiKey:      apiKey,
		url:         url,
		model:       model,
		timeout:     timeout,
		maxTokens:   maxTokens,
		temperature: temperature,
	}
}

func (c *Client) TranslateAndExplain(text string) (*entity.WordItem, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty input text")
	}
	request := &ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{Role: "system", Content: `# 角色定义
你是一个专业的英汉翻译与释义工具。
请根据输入内容的类型进行翻译和解释，并严格遵循以下规则：

## 1. 输入处理规则

### 1.1. 如果输入是 [单词] (如 "hello", "computer")
- **提供音标**: (IPA 格式)
- **提供多个词性与释义**: (如有)
  - 词性必须使用以下标准英文缩写之一:
    'n.' (名词), 'v.' (动词), 'adj.' (形容词), 'adv.' (副词),
    'prep.' (介词), 'conj.' (连词), 'pron.' (代词), 'int.' (感叹词)
- **提供使用示例**:
  - 最多 3 个
  - 示例格式固定为: '{英文例句}\n{中文翻译}'
	- 示例必须填写在 meanings[].examples 中（禁止使用顶层 examples 字段）

### 1.2. 如果输入是 [词组] (如 "look for", "take off")
- **不提供音标**
- **提供多个词性与释义**: (如有, 词性规则同上)
- **词性固定为**: 'phr.' (词组)
- **提供使用示例**:
  - 最多 3 个
  - 示例格式同单词
	- 示例必须填写在 meanings[].examples 中（禁止使用顶层 examples 字段）

### 1.3. 如果输入是 [句子] (如 "The quick brown fox")
- **不提供音标**
- **提供完整中文翻译**
- **词性固定为**: 'sent.' (句子)
- **不提供示例**

## 2. 函数调用约束 (极其重要)

**translate_and_explain 函数的 arguments 字段必须是有效的 JSON 对象字符串，格式如下：**

{
  "input_type": "word|phrase|sentence", 
  "original_text": "原始内容",
  "phonetics": [...],
  "meanings": [...]
}

**绝对禁止：**
- arguments 不能以方括号 [] 开头或结尾
- arguments 不能是数组格式
- arguments 不能包含任何无效的 JSON 语法
- arguments 前后不能有额外的文本或标记

## 3. 输出约束 (必须严格遵守)

1.  **必须使用 'translate_and_explain' 函数返回结果**
2.  **返回内容只能是一次函数调用**
    - 禁止输出任何额外文本, 解释, 提示语或自然语言内容
3.  **返回结果必须是结构化, 可解析的数据**
    - **arguments 字段必须是有效的 JSON 对象字符串，以 { 开始，以 } 结束**
    - [句子]翻译必须是直译，**不能**意译或改写
    - 字段含义清晰, 语义稳定
    - 不得省略规则要求的字段
    - 不得新增规则未要求的字段
    - **严禁**在 JSON 对象外层包裹数组方括号 []
    - **严禁**输出 Markdown 格式标记 (如 code block 标记)
4.  **分类合规性**
    - 必须严格遵守上述 1 / 2 / 3 的分类与输出约束
    - 不得混用不同类型的规则
    - 不得为词组或句子提供音标
    - 不得为句子提供示例
	    - 词性集合必须严格使用枚举值: n./v./adj./adv./prep./conj./pron./int./phr./sent.`},
			{Role: "user", Content: fmt.Sprintf("translate and explain: %s", text)},
		},
		Tools: []Tool{
			{
				Type: "function",
				Function: Function{
					Name:        "translate_and_explain",
					Description: "翻译并解释英文内容,根据内容类型返回相应结构",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"input_type": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"word", "phrase", "sentence"},
								"description": "输入内容类型:单词/词组/句子",
							},
							"original_text": map[string]interface{}{
								"type":        "string",
								"description": "原始内容",
							},
							"phonetics": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"head_word": map[string]interface{}{
											"type": "string",
										},
										"language_code": map[string]interface{}{
											"type": "string",
										},
										"text": map[string]interface{}{
											"type": "string",
										},
									},
									"required":             []string{"head_word", "language_code", "text"},
									"additionalProperties": false,
								},
								"description": "音标信息(仅input_type=word时需要填写)",
							},
							"meanings": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"part_of_speech": map[string]interface{}{
											"type":        "string",
											"enum":        []string{"n.", "v.", "adj.", "adv.", "prep.", "conj.", "pron.", "int.", "phr.", "sent."},
											"description": "词性: word 请选择 n./v./adj./adv./prep./conj./pron./int.; phrase 固定 phr.; sentence 固定 sent.",
										},
										"definitions": map[string]interface{}{
											"type":        "string",
											"description": "释义",
										},
										"examples": map[string]interface{}{
											"type":     "array",
											"maxItems": 3,
											"items": map[string]interface{}{
												"type": "string",
											},
											"description": "示例(仅input_type=word/phrase时需要填写),格式为:{英文例句}\\n{中文翻译}",
										},
									},
									"required":             []string{"part_of_speech", "definitions"},
									"additionalProperties": false,
								},
								"description": "翻译列表",
							},
						},
						"required":             []string{"input_type", "original_text", "meanings"},
						"additionalProperties": false,
					},
				},
			},
		},
		ToolChoice:  "required",
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chat request")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	result, err := util.SendPost(c.url, headers, requestBytes, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, errors.New(fmt.Sprintf("LLM API error: status %d", response.StatusCode))
		}

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}
		defer response.Body.Close()
		println(string(bodyBytes))

		var chatResponse ChatResponse
		err = json.Unmarshal(bodyBytes, &chatResponse)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal chat response")
		}

		if len(chatResponse.Choices) == 0 {
			return nil, errors.New("no choices in LLM response")
		}

		choice := chatResponse.Choices[0]
		if len(choice.Message.ToolCalls) == 0 {
			return nil, errors.New("no tool calls in LLM response")
		}

		toolCall := choice.Message.ToolCalls[0]
		if toolCall.Function.Name != "translate_and_explain" {
			return nil, errors.New("unexpected tool call: " + toolCall.Function.Name)
		}

		// Handle both array and object formats
		var argumentsBytes = []byte(toolCall.Function.Arguments)
		var temp interface{}
		err = json.Unmarshal(argumentsBytes, &temp)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal tool call arguments")
		}

		var translateResponse TranslateResponse
		switch v := temp.(type) {
		case []interface{}:
			// Array format - take first element
			if len(v) == 0 {
				return nil, errors.New("empty array in tool call arguments")
			}
			elementBytes, err := json.Marshal(v[0])
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal array element")
			}
			err = json.Unmarshal(elementBytes, &translateResponse)
		case map[string]interface{}:
			// Object format - unmarshal directly
			err = json.Unmarshal(argumentsBytes, &translateResponse)
		default:
			return nil, errors.New("invalid tool call arguments format")
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal translate response")
		}

		return &translateResponse, nil
	})

	if err != nil {
		return nil, err
	}

	translateResp := result.(*TranslateResponse)
	return c.mapToWordItem(translateResp), nil
}

func (c *Client) mapToWordItem(resp *TranslateResponse) *entity.WordItem {
	item := &entity.WordItem{
		ID:   entity.WordId(resp.OriginalText),
		Word: resp.OriginalText,
	}

	switch resp.InputType {
	case "word":
		wordPhonetics := make([]*entity.WordPhonetic, len(resp.Phonetics))
		for i, ph := range resp.Phonetics {
			wordPhonetics[i] = &entity.WordPhonetic{
				HeadWord:     ph.HeadWord,
				LanguageCode: ph.LanguageCode,
				Text:         ph.Text,
				Audio:        "",
			}
		}
		item.WordPhonetics = wordPhonetics

		wordMeanings := make([]*entity.WordMeaning, len(resp.Meanings))
		for i, meaning := range resp.Meanings {
			wordMeanings[i] = &entity.WordMeaning{
				PartOfSpeech: meaning.PartOfSpeech,
				Definitions:  meaning.Definitions,
				Examples:     meaning.Examples,
			}
		}
		item.WordMeanings = wordMeanings

	case "phrase":
		wordMeanings := make([]*entity.WordMeaning, len(resp.Meanings))
		for i, meaning := range resp.Meanings {
			wordMeanings[i] = &entity.WordMeaning{
				PartOfSpeech: "phr.",
				Definitions:  meaning.Definitions,
				Examples:     meaning.Examples,
			}
		}
		item.WordMeanings = wordMeanings

	case "sentence":
		wordMeanings := make([]*entity.WordMeaning, len(resp.Meanings))
		for i, meaning := range resp.Meanings {
			wordMeanings[i] = &entity.WordMeaning{
				PartOfSpeech: "sent.",
				Definitions:  meaning.Definitions,
				Examples:     meaning.Examples,
			}
		}
		item.WordMeanings = wordMeanings
	}

	return item
}

func (c *Client) TranslateWithStream(text string, disableStream bool, out io.Writer, ref bool) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", errors.New("empty input text")
	}

	promptBuilder := NewPromptBuilder()
	systemPrompt := promptBuilder.BuildTranslationPrompt(ref)

	request := &ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
		Stream:      !disableStream,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal translate request")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}

	if disableStream {
		return c.handleNonStreamingTranslate(headers, requestBytes)
	} else {
		return c.handleStreamingTranslate(headers, requestBytes, out)
	}
}

func (c *Client) handleNonStreamingTranslate(headers map[string]string, requestBytes []byte) (string, error) {
	responseProcessor := NewResponseProcessor()

	result, err := util.SendPostWithTimeout(c.url, headers, requestBytes, c.timeout, func(response *http.Response) (interface{}, error) {
		return responseProcessor.ProcessResponse(response)
	})

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func (c *Client) handleStreamingTranslate(headers map[string]string, requestBytes []byte, out io.Writer) (string, error) {
	streamProcessor := NewStreamProcessor(out)

	err := util.SendPostStreamWithTimeout(c.url, headers, requestBytes, c.timeout, func(response *http.Response) error {
		_, err := streamProcessor.ProcessStream(response)
		return err
	})

	if err != nil {
		return "", err
	}

	return "", nil
}
