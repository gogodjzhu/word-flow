package dict_etymonline

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
	"net/http"
	"regexp"
	"strings"
)

const Host = "https://www.etymonline.com"

type DictEtymonline struct {
}

func NewDictEtymonline(params map[string]interface{}) (*DictEtymonline, error) {
	return &DictEtymonline{}, nil
}

func (d *DictEtymonline) Search(word string) (*entity.WordItem, error) {
	url := Host + "/word/" + word
	pattern, _ := regexp.Compile(`(.*)`)
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
		h := doc.Find("div.ant-col-xs-24").Children().Nodes
		for _, defineNode := range h {
			for _, attribute := range defineNode.Attr {
				if attribute.Key == "class" && strings.HasPrefix(attribute.Val, "word--") {
					subDoc := goquery.NewDocumentFromNode(defineNode)
					var defineWord string
					for _, node := range subDoc.Children().Children().Children().Nodes {
						for _, a := range node.Attr {
							if a.Key == "class" && strings.HasPrefix(a.Val, "word__name--") {
								defineWord = goquery.NewDocumentFromNode(node).Text()
								break
							}
						}
					}
					partOfSpeechStr := pattern.FindString(defineWord)
					word = strings.TrimSpace(strings.ReplaceAll(defineWord, partOfSpeechStr, ""))
					partOfSpeech := partOfSpeechStr[1 : len(partOfSpeechStr)-1]

					var defineParas string
					for _, definePara := range subDoc.Find("section").Children().Nodes {
						defineParaDoc := goquery.NewDocumentFromNode(definePara)
						if defineParaDoc.Is("blockquote") {
							defineParas += "> "
						}
						defineParas += strings.TrimSpace(defineParaDoc.Text()) + "\n"
					}
					defineParas = strings.TrimRight(defineParas, "\n")
					meanings = append(meanings, &entity.WordMeaning{
						PartOfSpeech: partOfSpeech,
						Definitions:  defineParas,
					})
					break
				}
			}
		}
		return &entity.WordItem{
			ID:           entity.WordId(word),
			Word:         word,
			WordMeanings: meanings,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.WordItem), nil
}
