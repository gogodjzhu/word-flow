package dict_chatgpt

import (
	"encoding/json"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

// DictChatgpt is the implementation of Dict interface for chatgpt,
// apiDoc: https://learn.microsoft.com/zh-cn/azure/ai-services/openai/reference
type DictChatgpt struct {
	resourceName string
	deploymentId string
	apiVersion   string
	key          string
}

func NewDictChatgpt(params map[string]interface{}) (*DictChatgpt, error) {
	var ok bool
	dict := &DictChatgpt{}
	if dict.resourceName, ok = params[config.DictConfigChatgptResource].(string); !ok {
		return nil, errors.New("resourceName is required")
	}
	if dict.deploymentId, ok = params[config.DictConfigChatgptDeploymentid].(string); !ok {
		return nil, errors.New("deploymentId is required")
	}
	if dict.apiVersion, ok = params[config.DictConfigChatgptApiversion].(string); !ok {
		return nil, errors.New("apiVersion is required")
	}
	if dict.key, ok = params[config.DictConfigChatgptKey].(string); !ok {
		return nil, errors.New("key is required")
	}
	return dict, nil
}

func (d *DictChatgpt) Search(word string) (*entity.WordItem, error) {
	// POST https://{your-resource-name}.openai.azure.com/openai/deployments/{deployment-id}/completions?api-version={api-version}
	url := "https://" + d.resourceName + ".openai.azure.com/openai/deployments/" + d.deploymentId + "/chat/completions?api-version=" + d.apiVersion
	req := chatGptCompletionRequest{
		Messages: []chatGptCompletionRequestMessage{
			{
				Role:    "user",
				Content: "Print the definition of phrase: " + word,
			},
		},
		Functions: []chatGptCompletionRequestFunction{
			{
				Name:        "print_phrase_definition",
				Description: "Print the definition of phrase",
				Parameters: chatGptCompletionRequestFunctionParameters{
					Type: "object",
					Properties: map[string]chatGptCompletionRequestFunctionParametersProperty{
						"phrase": {
							Type:        "string",
							Description: "The phrase itself",
						},
						"phonetic": {
							Type:        "string",
							Description: "Phonetic symbols of phrase, or 'NaN' if no explicit definition, eg: 'sɜːrvɪsɪz', 'ˌiːkəˈnɑːmɪk'",
						},
						"definition": {
							Type:        "string",
							Description: "The definition of the given phrase, or 'NaN' if no explicit definition",
						},
						"origin": {
							Type:        "string",
							Description: "The origin of the given phrase, or 'NaN' if no explicit definition",
						},
						"example": {
							Type:        "string",
							Description: "An example of given phrase, or 'NaN' if no explicit definition",
						},
					},
					Required: []string{"phrase", "phonetic", "definition", "origin", "example"},
				},
			},
		},
		FunctionCall: chatGptCompletionRequestFunctionCall{
			Name: "print_phrase_definition",
		},
	}
	type printPhraseDefinitionParameters struct {
		Phrase     string `json:"phrase"`
		Phonetic   string `json:"phonetic"`
		Definition string `json:"definition"`
		Origin     string `json:"origin"`
		Example    string `json:"example"`
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	result, err := util.SendPost(url, map[string]string{
		"api-key": d.key,
	}, bs, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, errors.New("failed to sendGet")
		}
		defer response.Body.Close()
		var resp chatGptCompletionResponse
		err := json.NewDecoder(response.Body).Decode(&resp)
		if err != nil {
			return nil, err
		}
		var wordItem *entity.WordItem
		for _, choice := range resp.Choices {
			if choice.Message.FunctionCall.Name == "print_phrase_definition" {
				wordItem = &entity.WordItem{
					ID:   entity.WordId(word),
					Word: word,
				}
				var parameters printPhraseDefinitionParameters
				err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &parameters)
				if err != nil {
					return nil, err
				}
				// origin
				if parameters.Origin != "NaN" {
					wordItem.Origin = parameters.Origin
				}
				// meaning
				meaning := entity.WordMeaning{}
				if parameters.Definition != "NaN" {
					meaning.Definitions = parameters.Definition
				}
				if parameters.Example != "NaN" {
					meaning.Examples = []string{parameters.Example}
				}
				wordItem.WordMeanings = []*entity.WordMeaning{&meaning}
				// phonetic
				phonetic := entity.WordPhonetic{}
				if parameters.Phonetic != "NaN" {
					p := parameters.Phonetic
					p = strings.TrimPrefix(p, "/")
					p = strings.TrimPrefix(p, "[")
					p = strings.TrimSuffix(p, "/")
					p = strings.TrimSuffix(p, "]")
					phonetic.LanguageCode = "en/us"
					phonetic.Text = p
				}
				wordItem.WordPhonetics = []*entity.WordPhonetic{&phonetic}
				break
			}
		}
		return wordItem, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.WordItem), nil
}

type chatGptCompletionRequest struct {
	Messages     []chatGptCompletionRequestMessage    `json:"messages"`
	Functions    []chatGptCompletionRequestFunction   `json:"functions"`
	FunctionCall chatGptCompletionRequestFunctionCall `json:"function_call"`
}
type chatGptCompletionRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatGptCompletionRequestFunction struct {
	Name        string                                     `json:"name"`
	Description string                                     `json:"description"`
	Parameters  chatGptCompletionRequestFunctionParameters `json:"parameters"`
}

type chatGptCompletionRequestFunctionParameters struct {
	Type       string                                                        `json:"type"`
	Properties map[string]chatGptCompletionRequestFunctionParametersProperty `json:"properties"`
	Required   []string                                                      `json:"required"`
}

type chatGptCompletionRequestFunctionParametersProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type chatGptCompletionRequestFunctionCall struct {
	Name string `json:"name"`
}

type chatGptCompletionResponse struct {
	Id                  string                                        `json:"id"`
	Object              string                                        `json:"object"`
	Created             int64                                         `json:"created"`
	Model               string                                        `json:"model"`
	PromptFilterResults []chatGptCompletionResponsePromptFilterResult `json:"prompt_filter_results"`
	Choices             []chatGptCompletionResponseChoice             `json:"choices"`
	Usage               chatGptCompletionResponseUsage                `json:"usage"`
}

type chatGptCompletionResponsePromptFilterResult struct {
}

type chatGptCompletionResponseChoice struct {
	Index                int                                                 `json:"index"`
	FinishReason         string                                              `json:"finish_reason"`
	Message              chatGptCompletionResponseChoiceMessage              `json:"message"`
	ContentFilterResults chatGptCompletionResponseChoiceContentFilterResults `json:"content_filter_results"`
}

type chatGptCompletionResponseChoiceMessage struct {
	Role         string                                             `json:"role"`
	FunctionCall chatGptCompletionResponseChoiceMessageFunctionCall `json:"function_call"`
}

type chatGptCompletionResponseChoiceMessageFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatGptCompletionResponseChoiceContentFilterResults struct {
}

type chatGptCompletionResponseUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
