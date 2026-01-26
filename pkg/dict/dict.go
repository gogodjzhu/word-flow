package dict

import (
	"github.com/gogodjzhu/word-flow/internal/buzz_error"
	"github.com/gogodjzhu/word-flow/internal/config"
	dict_ecdict "github.com/gogodjzhu/word-flow/pkg/dict/ecdict"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	dict_etymonline "github.com/gogodjzhu/word-flow/pkg/dict/etymonline"
	dict_llm "github.com/gogodjzhu/word-flow/pkg/dict/llm"
	dict_mwebster "github.com/gogodjzhu/word-flow/pkg/dict/mwebster"
	dict_youdao "github.com/gogodjzhu/word-flow/pkg/dict/youdao"
	"github.com/pkg/errors"
)

type Dict interface {
	Search(word string) (*entity.WordItem, error)
}

type Endpoint string

const (
	Youdao     Endpoint = "youdao"
	Etymonline Endpoint = "etymonline"
	Ecdict     Endpoint = "ecdict"
	MWebster   Endpoint = "mwebster"
	LLM        Endpoint = "llm"
)

// AvailableEndpoints returns all supported dictionary endpoint identifiers.
func AvailableEndpoints() []string {
	return []string{
		string(Youdao),
		string(Etymonline),
		string(Ecdict),
		string(MWebster),
		string(LLM),
	}
}

func NewDict(conf *config.DictConfig) (Dict, error) {
	endpoint := conf.Default
	endpointConfig, err := conf.GetConfigForEndpoint(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config for endpoint %s", endpoint)
	}

	switch Endpoint(endpoint) {
	case Youdao:
		return dict_youdao.NewDictYoudao(endpointConfig.(*config.YoudaoConfig))
	case Etymonline:
		return dict_etymonline.NewDictEtymonline(endpointConfig.(*config.EtymonlineConfig))
	case Ecdict:
		return dict_ecdict.NewDictEcdit(endpointConfig.(*config.EcdictConfig))
	case MWebster:
		return dict_mwebster.NewDictMWebster(endpointConfig.(*config.MWebsterConfig))
	case LLM:
		return dict_llm.NewDictLLM(endpointConfig.(*config.LLMConfig))
	default:
		return nil, buzz_error.InvalidEndpoint(endpoint)
	}
}
