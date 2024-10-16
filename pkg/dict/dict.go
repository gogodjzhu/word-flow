package dict

import (
	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	dict_chatgpt "github.com/gogodjzhu/word-flow/pkg/dict/chatgpt"
	dict_ecdict "github.com/gogodjzhu/word-flow/pkg/dict/ecdict"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	dict_etymonline "github.com/gogodjzhu/word-flow/pkg/dict/etymonline"
	dict_mwebster "github.com/gogodjzhu/word-flow/pkg/dict/mwebster"
	dict_youdao "github.com/gogodjzhu/word-flow/pkg/dict/youdao"
)

type Dict interface {
	Search(word string) (*entity.WordItem, error)
}

type Endpoint string

const (
	Youdao     Endpoint = "youdao"
	Etymonline Endpoint = "etymonline"
	Ecdict     Endpoint = "ecdict"
	Chatgpt    Endpoint = "chatgpt"
	MWebster   Endpoint = "mwebster"
)

func NewDict(conf *config.DictConfig) (Dict, error) {
	switch Endpoint(conf.Default) {
	case Youdao:
		return dict_youdao.NewDictYoudao(conf.Parameters)
	case Etymonline:
		return dict_etymonline.NewDictEtymonline(conf.Parameters)
	case Ecdict:
		return dict_ecdict.NewDictEcdit(conf.Parameters)
	case Chatgpt:
		return dict_chatgpt.NewDictChatgpt(conf.Parameters)
	case MWebster:
		return dict_mwebster.NewDictMWebster(conf.Parameters)
	default:
		return nil, buzz_error.InvalidEndpoint(conf.Default)
	}
}
