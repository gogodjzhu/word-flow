package dict_etymonline

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
)

const Host = "https://www.etymonline.com"

type DictEtymonline struct {
}

func NewDictEtymonline(config *config.EtymonlineConfig) (*DictEtymonline, error) {
	return &DictEtymonline{}, nil
}

func (d *DictEtymonline) Search(word string) (*entity.WordItem, error) {
	url := Host + "/word/" + word
	pattern, _ := regexp.Compile(`\(([^)]+)\)`)
	result, err := util.SendGet(url, nil, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, errors.New("failed to sendGet")
		}
		defer response.Body.Close()
		doc, e := goquery.NewDocumentFromReader(response.Body)
		if e != nil {
			return nil, errors.Wrap(e, "failed to parse response body")
		}
		meanings := make([]*entity.WordMeaning, 0)
		h := doc.Find("section.prose-lg").Nodes
		for _, defineNode := range h {
			if defineNode.Data == "section" {
				wordMeaning := &entity.WordMeaning{}
				pas := goquery.NewDocumentFromNode(defineNode)
				for idx, pass := range pas.Children().Nodes {
					if idx == 0 {
						partOfSpeech := pattern.FindStringSubmatch(strings.TrimSpace(goquery.NewDocumentFromNode(pass).Text()))
						if len(partOfSpeech) < 2 {
							continue
						}
						wordMeaning.PartOfSpeech = partOfSpeech[1]
					}
					if idx >= 1 {
						var definition string
						for _, passd := range goquery.NewDocumentFromNode(pass).Children().Nodes {
							if passd.Data == "blockquote" {
								definition += " > " + strings.TrimSpace(goquery.NewDocumentFromNode(passd).Text()) + "\n"
							} else {
								definition += strings.TrimSpace(goquery.NewDocumentFromNode(passd).Text()) + "\n"
							}
						}
						wordMeaning.Definitions += definition
					}
				}
				meanings = append(meanings, wordMeaning)
			}
		}
		return &entity.WordItem{
			ID:           entity.WordId(word),
			Word:         word,
			Source:       "etymonline",
			WordMeanings: meanings,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.WordItem), nil
}
