package config

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func configDir() string {
	var path string
	if a := os.Getenv("WORDFLOW_HOME"); a != "" {
		path = a
	} else {
		d, _ := os.UserHomeDir()
		path = filepath.Join(d, ".config", "wordflow")
	}
	return path
}

func InitConfig(configFilename string) error {
	_, err := os.Stat(configFilename)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("config file already exists: %s", configFilename))
	}
	// create config dir
	conf := defaultConfig
	conf.Common.ConfigFilename = configFilename
	return conf.Save()
}

func ReadConfig() (*Config, error) {
	return ReadConfigSpecified("")
}

func ReadConfigSpecified(configFilename string) (*Config, error) {
	var root Config
	if configFilename == "" {
		configFilename = defaultConfig.Common.ConfigFilename
	}
	bytes, err := os.ReadFile(configFilename)
	if err != nil {
		if os.IsNotExist(err) {
			err = InitConfig(configFilename)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed init config file: %s", configFilename))
			}
			// read again
			bytes, err = os.ReadFile(configFilename)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed read config file: %s", configFilename))
			}
		} else {
			return nil, errors.Wrap(err, fmt.Sprintf("failed read config file: %s", configFilename))
		}
	}
	// overwrite default config
	err = yaml.Unmarshal(bytes, &root)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed unmarshal config file: %s", configFilename))
	}
	return &root, nil
}

func (c *Config) ToString() (string, error) {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return "", errors.Wrap(err, "failed marshal config")
	}
	return string(bytes), nil
}

func (c *Config) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "failed marshal config")
	}
	dir := filepath.Dir(c.Common.ConfigFilename)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed create config dir: %s", dir))
	}
	return os.WriteFile(c.Common.ConfigFilename, bytes, 0644)
}

type Config struct {
	Version  string          `yaml:"version"`
	Common   *CommonConfig   `yaml:"common"`
	Dict     *DictConfig     `yaml:"dict"`
	Notebook *NotebookConfig `yaml:"notebook"`
}

type CommonConfig struct {
	BasePath       string `yaml:"basePath"`
	ConfigFilename string `yaml:"configFilename"`
}

type DictConfig struct {
	Default    string                 `yaml:"default"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

func (dc *DictConfig) GetConfigForEndpoint(endpoint string) (DictEndpointConfig, error) {
	return globalMapper.MapToEndpointConfig(endpoint, dc.Parameters)
}

const (
	DictConfigEcdictDbfilename = "ecdict.dbfilename"
)

const (
	DictConfigMWebsterKey = "mwebster.key"
)

const (
	DictConfigLLMApiKey      = "llm.api_key"
	DictConfigLLMUrl         = "llm.url"
	DictConfigLLMModel       = "llm.model"
	DictConfigLLMTimeout     = "llm.timeout"
	DictConfigLLMMaxTokens   = "llm.max_tokens"
	DictConfigLLMTemperature = "llm.temperature"
)

type NotebookConfig struct {
	Default    string                 `yaml:"default"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

const (
	NotebookConfigNotebookBasepath = "notebook.basepath"
	NotebookConfigFSRSMaxReviews   = "fsrs.max_reviews_per_session"
	NotebookConfigFSRSLimit        = "fsrs.new_cards_per_day"
)

var defaultConfig = Config{
	Version: "0.1",
	Common: &CommonConfig{
		BasePath:       configDir(),
		ConfigFilename: filepath.Join(configDir(), "config.yaml"),
	},
	Dict: &DictConfig{
		Default: "youdao",
		Parameters: map[string]interface{}{
			DictConfigEcdictDbfilename: filepath.Join(configDir(), "stardict.db"),
			DictConfigLLMTimeout:       "30s",
			DictConfigLLMMaxTokens:     2000,
			DictConfigLLMTemperature:   0.3,
		},
	},
	Notebook: &NotebookConfig{
		Default: "default",
		Parameters: map[string]interface{}{
			NotebookConfigNotebookBasepath: filepath.Join(configDir(), "notebooks"),
			NotebookConfigFSRSMaxReviews:   50,
			NotebookConfigFSRSLimit:        20,
		},
	},
}
