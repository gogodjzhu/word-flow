package dict_ecdict

import (
	"github.com/gogodjzhu/word-flow/internal/config"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"strings"
)

type DictEcdict struct {
	db *gorm.DB
}

func NewDictEcdit(params map[string]interface{}) (*DictEcdict, error) {
	dbfilename, ok := params[config.DictConfigEcdictDbfilename].(string)
	if !ok {
		return nil, errors.New("dbfilename not found")
	}
	err := PrepareDBFile(dbfilename)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbfilename), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open db file")
	}
	return &DictEcdict{
		db: db,
	}, nil
}

func PrepareDBFile(dbfilename string) error {
	_, err := os.Stat(dbfilename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// TODO download from url
		return errors.New("db file not found")
	}
	return errors.Wrap(err, "failed to prepare db file")
}

func (d *DictEcdict) Search(word string) (*entity.WordItem, error) {
	var wordItem Word
	d.db.Raw("select * from stardict where word = ?", word).Scan(&wordItem)
	result := &entity.WordItem{
		ID:            entity.WordId(wordItem.Word),
		Word:          wordItem.Word,
		WordPhonetics: make([]*entity.WordPhonetic, 0),
		WordMeanings:  make([]*entity.WordMeaning, 0),
	}
	result.WordPhonetics = append(result.WordPhonetics, &entity.WordPhonetic{
		Text: wordItem.Phonetic,
	})
	definitions := strings.Split(wordItem.Translation, "\n")
	for _, definition := range definitions {
		var partOfSpeechStr string
		var definitionStr string
		dotPos := strings.Index(definition, ".")
		if dotPos != -1 {
			partOfSpeechStr = definition[:dotPos+1]
		}
		definitionStr = definition[dotPos+1:]
		result.WordMeanings = append(result.WordMeanings, &entity.WordMeaning{
			PartOfSpeech: partOfSpeechStr,
			Definitions:  definitionStr,
		})
	}
	return result, nil
}

type Word struct {
	gorm.Model
	Id          int64  `gorm:"column:id"`
	Word        string `gorm:"column:word"`
	ShortWorld  string `gorm:"column:sw"`
	Phonetic    string
	Definition  string
	Translation string
	Pos         string
	Collins     string
	Oxford      string
	Tag         string
	Bnc         string
	Frq         string
	Exchange    string
	Detail      string
	Audio       string
}

func (Word) TableName() string {
	return "stardict"
}
