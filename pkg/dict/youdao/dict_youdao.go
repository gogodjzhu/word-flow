package dict_youdao

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/internal/util"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
)

const Host = "https://dict.youdao.com"

type DictYoudao struct {
}

func NewDictYoudao(config *config.YoudaoConfig) (*DictYoudao, error) {
	return &DictYoudao{}, nil
}

func (d *DictYoudao) Search(word string) (*entity.WordItem, error) {
	params := neturl.Values{}
	params.Add("q", word)
	url := Host + "/search?" + params.Encode()
	result, err := util.SendGet(url, nil, func(response *http.Response) (interface{}, error) {
		if response.StatusCode != 200 {
			return nil, errors.New("failed to sendGet")
		}
		defer response.Body.Close()
		doc, e := goquery.NewDocumentFromReader(response.Body)
		if e != nil {
			return nil, errors.Wrap(e, "failed to parse response body")
		}
		keyword := strings.TrimSpace(doc.Find("span.keyword").Text())
		if len(strings.TrimSpace(keyword)) == 0 {
			return nil, buzz_error.InvalidInput("Invalid word: " + word)
		}
		trans := strings.TrimSpace(doc.Find("#phrsListTab > div.trans-container > ul").Text())
		enPhonetic := strings.TrimSpace(doc.Find("#phrsListTab > h2 > div > span:nth-child(1)").Text())
		usPhonetic := strings.TrimSpace(doc.Find("#phrsListTab > h2 > div > span:nth-child(2)").Text())
		var usages []string
		doc.Find("#bilingual > ul > li").Each(func(i int, selection *goquery.Selection) {
			var exampleEn, exampleCh string
			selection.Find("p").Each(func(ii int, selection *goquery.Selection) {
				switch ii {
				case 0:
					exampleEn = strings.TrimSpace(selection.Text())
				case 1:
					exampleCh = strings.TrimSpace(selection.Text())
				default:
				}
			})
			if exampleEn != "" && exampleCh != "" {
				usages = append(usages, exampleEn+"\n"+exampleCh)
			}
		})
		if len(usages) > 3 {
			usages = usages[:3]
		}
		if formatWordMeanings(trans) == nil || len(formatWordMeanings(trans)) == 0 {
			// normalize trans: collapse whitespace for a clean single-line definition
			return &entity.WordItem{
				ID:     entity.WordId(keyword),
				Word:   keyword,
				Source: "youdao",
				//WordPhonetics: formatPhonetic(keyword, enPhonetic, usPhonetic),
				WordMeanings: []*entity.WordMeaning{
					{
						PartOfSpeech: "",
						Definitions:  renderTrans(trans),
					},
				},
			}, nil
		} else {
			return &entity.WordItem{
				ID:            entity.WordId(keyword),
				Word:          keyword,
				Source:        "youdao",
				WordPhonetics: formatPhonetic(keyword, enPhonetic, usPhonetic),
				WordMeanings:  formatWordMeanings(trans),
				Examples:      usages,
			}, nil
		}
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.WordItem), nil
}

func formatPhonetic(word, enPhonetic, usPhonetic string) []*entity.WordPhonetic {
	if strings.Contains(usPhonetic, "英") {
		enPhonetic, usPhonetic = usPhonetic, enPhonetic
	}
	// extract text inside [text]
	if enPhonetic != "" {
		enPhonetic = enPhonetic[strings.Index(enPhonetic, "[")+1 : strings.Index(enPhonetic, "]")]
	}
	if usPhonetic != "" {
		usPhonetic = usPhonetic[strings.Index(usPhonetic, "[")+1 : strings.Index(usPhonetic, "]")]
	}
	return []*entity.WordPhonetic{
		{
			LanguageCode: "en",
			Text:         enPhonetic,
			Audio:        fmt.Sprintf("https://dict.youdao.com/dictvoice?audio=%s&type=1", word),
		},
		{
			LanguageCode: "us",
			Text:         usPhonetic,
			Audio:        fmt.Sprintf("https://dict.youdao.com/dictvoice?audio=%s&type=2", word),
		},
	}
}

func formatWordMeanings(str string) []*entity.WordMeaning {
	if strings.TrimSpace(str) == "" {
		return nil
	}
	partOfSpeeches := []string{"n. ", "v. ", "adj. ", "adv. ", "pron. ", "prep. ", "conj. ", "art. ", "num. ", "int. ", "abbr. ", "aux. ", "vi. ", "vt. "}
	meanings := make([]*entity.WordMeaning, 0)
	var wordMeaning *entity.WordMeaning
	for _, term := range util.SplitWorker(str, partOfSpeeches) {
		if len(strings.TrimSpace(term)) == 0 {
			continue
		}
		if util.Contains(partOfSpeeches, term) {
			if wordMeaning != nil { // part of speech is always the first element, if not empty, it means a new word meaning
				meanings = append(meanings, wordMeaning)
			}
			wordMeaning = &entity.WordMeaning{}
			wordMeaning.PartOfSpeech = strings.TrimSpace(term)
		} else {
			if wordMeaning == nil { // not yet found part of speech
				continue
			}
			var trimTerm string
			for _, t := range util.SplitWorker(term, []string{"\n"}) {
				t = strings.TrimSpace(t)
				if len(t) != 0 {
					trimTerm += t + "\n"
				}
			}
			wordMeaning.Definitions += strings.TrimSpace(trimTerm)
		}
	}
	if wordMeaning != nil && wordMeaning.PartOfSpeech != "" && wordMeaning.Definitions != "" {
		meanings = append(meanings, wordMeaning)
	}
	return meanings
}

// renderTrans collapses all whitespace (including newlines and tabs) into single spaces
// and trims leading/trailing spaces for a clean one-line definition.
func renderTrans(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
