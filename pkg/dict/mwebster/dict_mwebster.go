package dict_mwebster

import (
	"encoding/json"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type DictMWebster struct {
	key string
}

func NewDictMWebster(params map[string]interface{}) (*DictMWebster, error) {
	var ok bool
	dict := &DictMWebster{}
	if dict.key, ok = params[config.DictConfigMWebsterKey].(string); !ok {
		return nil, errors.New("missing mwebster.key in params")
	}
	return dict, nil
}

func (d *DictMWebster) Search(word string) (*entity.WordItem, error) {
	url := "https://www.dictionaryapi.com/api/v3/references/collegiate/json/" + word + "?key=" + d.key
	result, err := util.SendGet(url, nil, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, errors.New("failed to sendGet")
		}
		defer response.Body.Close()
		var wordStructures []WordStructure

		if bs, err := io.ReadAll(response.Body); err != nil {
			return nil, errors.Wrap(err, "failed to read from response body")
		} else {
			if err := json.Unmarshal(bs, &wordStructures); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal response body:'"+string(bs)+"'")
			}
		}
		if len(wordStructures) == 0 {
			return nil, errors.New("word not found")
		}

		var wordItem entity.WordItem
		wordItem.ID = entity.WordId(word)
		wordItem.Word = word
		wordItem.WordPhonetics = make([]*entity.WordPhonetic, 0)
		wordItem.WordMeanings = make([]*entity.WordMeaning, 0)
		for _, wordStructure := range wordStructures {
			if wordStructure.Homograph == 0 && wordStructure.Meta.ID != word {
				continue
			} else if wordStructure.Homograph != 0 && wordStructure.Meta.ID != word+":"+strconv.Itoa(wordStructure.Homograph) {
				continue
			}
			var wordMeaning entity.WordMeaning
			wordMeaning.PartOfSpeech = wordStructure.FunctionLabel + "."
			var definitions string
			for _, shortDefinition := range wordStructure.ShortDefinitions {
				definitions += shortDefinition + "; "
			}
			wordMeaning.Definitions = definitions
			wordItem.WordMeanings = append(wordItem.WordMeanings, &wordMeaning)

			var wordPhonetic entity.WordPhonetic
			wordPhonetic.HeadWord = wordStructure.HeadwordInformation.Headword
			for _, pronunciation := range wordStructure.HeadwordInformation.Pronunciations {
				wordPhonetic.Text = pronunciation.Written
				// see https://www.dictionaryapi.com/products/json#sec-2.prs
				audioUrl := "https://media.merriam-webster.com/audio/prons/[language_code]/[country_code]/[format]/[subdirectory]/[base filename].[format]"
				audioUrl = strings.ReplaceAll(audioUrl, "[language_code]", "en")
				audioUrl = strings.ReplaceAll(audioUrl, "[country_code]", "us")
				audioUrl = strings.ReplaceAll(audioUrl, "[format]", "mp3")
				if strings.HasPrefix(pronunciation.Sound.Audio, "bix") {
					audioUrl = strings.ReplaceAll(audioUrl, "[subdirectory]", "bix")
				} else if strings.HasPrefix(pronunciation.Sound.Audio, "gg") {
					audioUrl = strings.ReplaceAll(audioUrl, "[subdirectory]", "gg")
				} else if len(pronunciation.Sound.Audio) == 1 {
					// regex: 0-9 or '_'
					if match, err := regexp.MatchString(pronunciation.Sound.Audio, "^[0-9_]$"); err != nil {
						return nil, err
					} else if match {
						audioUrl = strings.ReplaceAll(audioUrl, "[subdirectory]", "number")
					}
				}
				audioUrl = strings.ReplaceAll(audioUrl, "[subdirectory]", pronunciation.Sound.Audio[0:1])
				audioUrl = strings.ReplaceAll(audioUrl, "[base filename]", pronunciation.Sound.Audio)
				wordPhonetic.Audio = audioUrl
				break
			}
			wordItem.WordPhonetics = append(wordItem.WordPhonetics, &wordPhonetic)
		}
		return &wordItem, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.WordItem), nil
}

type WordStructure struct {
	Meta                WordStructureMeta                `json:"meta"`
	Homograph           int                              `json:"hom"`
	HeadwordInformation WordStructureHeadwordInformation `json:"hwi"`
	FunctionLabel       string                           `json:"fl"`
	ShortDefinitions    []string                         `json:"shortdef"`
}

type WordStructureMeta struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
	Sort string `json:"sort"`
}

type WordStructureHeadwordInformation struct {
	Headword       string                       `json:"hw"`
	Pronunciations []WordStructurePronunciation `json:"prs"`
}

type WordStructurePronunciation struct {
	Written string                          `json:"mw"`
	Sound   WordStructurePronunciationSound `json:"sound"`
}

type WordStructurePronunciationSound struct {
	Audio string `json:"audio"`
	Ref   string `json:"ref"`
	Stat  string `json:"stat"`
}
