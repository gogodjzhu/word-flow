package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func configDir() string {
	if a := os.Getenv("WORDFLOW_HOME"); a != "" {
		return a
	}
	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".config", "wordflow")
}

type Duration time.Duration

func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		s := value.Value
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("cannot parse duration %q: %w", s, err)
		}
		*d = Duration(parsed)
		return nil
	default:
		return fmt.Errorf("cannot parse duration from %v", value.Kind)
	}
}

type Config struct {
	Version  string          `yaml:"version"`
	Common   *CommonConfig   `yaml:"-"`
	Dict     *DictConfig     `yaml:"dict"`
	Trans    *TransConfig    `yaml:"trans"`
	Notebook *NotebookConfig `yaml:"notebook"`
}

type TransConfig struct {
	Default string           `yaml:"default"`
	LLM     *TransLLMConfig  `yaml:"llm"`
	Google  *TransGoogleConfig `yaml:"google"`
}

func (tc *TransConfig) GetEndpointConfig(endpoint string) (TransEndpointConfig, error) {
	switch endpoint {
	case "llm":
		return tc.LLM, nil
	case "google":
		return tc.Google, nil
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

type CommonConfig struct {
	BasePath       string
	ConfigFilename string
}

type DictConfig struct {
	Default   string            `yaml:"default"`
	LLM       *LLMConfig        `yaml:"llm"`
	Youdao    *YoudaoConfig     `yaml:"youdao"`
	Ecdict    *EcdictConfig     `yaml:"ecdict"`
	Etymoline *EtymonlineConfig `yaml:"etymonline"`
	MWebster  *MWebsterConfig   `yaml:"mwebster"`
	Google    *GoogleConfig     `yaml:"google"`
}

func (dc *DictConfig) GetEndpointConfig(endpoint string) (DictEndpointConfig, error) {
	switch endpoint {
	case "youdao":
		return dc.Youdao, nil
	case "llm":
		return dc.LLM, nil
	case "ecdict":
		return dc.Ecdict, nil
	case "etymonline":
		return dc.Etymoline, nil
	case "mwebster":
		return dc.MWebster, nil
	case "google":
		return dc.Google, nil
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

type NotebookConfig struct {
	Default  string            `yaml:"default"`
	Settings *NotebookSettings `yaml:"settings"`
}

type NotebookSettings struct {
	BasePath       string `yaml:"basepath,omitempty"`
	MaxReviews     int    `yaml:"max_reviews_per_session"`
	NewCardsPerDay int    `yaml:"new_cards_per_day"`
}

func (ns *NotebookSettings) Validate() error {
	if ns.MaxReviews <= 0 {
		return errors.New("max_reviews_per_session must be positive")
	}
	if ns.NewCardsPerDay < 0 {
		return errors.New("new_cards_per_day must be non-negative")
	}
	return nil
}

const DefaultConfigVersion = "v1"

func DefaultConfig() *Config {
	dir := configDir()
	return &Config{
		Version: DefaultConfigVersion,
		Common: &CommonConfig{
			BasePath:       dir,
			ConfigFilename: filepath.Join(dir, "config.yaml"),
		},
		Dict: &DictConfig{
			Default:   "youdao",
			LLM:       &LLMConfig{Timeout: Duration(30 * time.Second), MaxTokens: 2000, Temperature: 0.3},
			Youdao:    &YoudaoConfig{},
			Ecdict:    &EcdictConfig{},
			Etymoline: &EtymonlineConfig{},
			MWebster:  &MWebsterConfig{},
		},
		Notebook: &NotebookConfig{
			Default:  "default",
			Settings: &NotebookSettings{MaxReviews: 50, NewCardsPerDay: 20},
		},
	}
}

func ConfigTemplate() string {
	return configTemplate
}

const configTemplate = `# Wordflow configuration
version: v1

dict:
  # Default dictionary endpoint. Options: youdao, llm, ecdict, etymonline, mwebster, google
  default: youdao

  youdao: {}

  llm:
    # LLM provider settings (required if dict.default is llm or using trans command)
    # api_key: ""           # Required. Set via WORDFLOW_DICT_LLM_API_KEY or wordflow config set dict.llm.api_key
    # url: ""               # Required. Full API endpoint URL, not base URL. e.g. https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
    # model: ""             # Required. LLM model name, e.g. glm-4
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3

  ecdict:
    # Local dictionary database path. Defaults to <WORDFLOW_HOME>/stardict.db if empty
    # db_filename: ""

  etymonline: {}

  mwebster:
    # Merriam-Webster API key (required if using mwebster)
    # key: ""

  google: {}

trans:
  # Default translator endpoint. Options: google, llm
  default: google

  google: {}

  # llm:
  #   # LLM provider settings for translation (required if trans.default is llm)
  #   # api_key: ""           # Required. Set via WORDFLOW_TRANS_LLM_API_KEY or wordflow config set trans.llm.api_key
  #   # url: ""               # Required. Full API endpoint URL
  #   # model: ""             # Required. LLM model name
  #   timeout: 30s
  #   max_tokens: 2000
  #   temperature: 0.3

notebook:
  default: default

  settings:
    # Notebook storage path. Defaults to <WORDFLOW_HOME>/notebooks if empty
    # basepath: ""
    max_reviews_per_session: 50
    new_cards_per_day: 20
`

func ConfigFilePath() string {
	return filepath.Join(configDir(), "config.yaml")
}

func InitConfig(configFilename string) error {
	_, err := os.Stat(configFilename)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("config file already exists: %s", configFilename))
	}
	dir := filepath.Dir(configFilename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create config dir: %s", dir))
	}
	return os.WriteFile(configFilename, []byte(configTemplate), 0644)
}

func ReadConfig() (*Config, error) {
	return ReadConfigSpecified("")
}

func ReadConfigSpecified(configFilename string) (*Config, error) {
	if configFilename == "" {
		configFilename = ConfigFilePath()
	}
	_, err := os.Stat(configFilename)
	if os.IsNotExist(err) {
		if initErr := InitConfig(configFilename); initErr != nil {
			return nil, errors.Wrap(initErr, fmt.Sprintf("failed init config file: %s", configFilename))
		}
	} else if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to stat config file: %s", configFilename))
	}

	cfg, loadErr := LoadConfig(configFilename)
	if loadErr != nil {
		return nil, loadErr
	}
	return cfg, nil
}

func applyDefaults(cfg *Config, configFilename string) {
	dir := configDir()
	if cfg.Common == nil {
		cfg.Common = &CommonConfig{}
	}
	cfg.Common.BasePath = dir
	cfg.Common.ConfigFilename = configFilename

	if cfg.Version == "" {
		cfg.Version = DefaultConfigVersion
	}
	if cfg.Dict == nil {
		cfg.Dict = &DictConfig{}
	}
	if cfg.Dict.Default == "" {
		cfg.Dict.Default = "youdao"
	}
	if cfg.Dict.LLM == nil {
		cfg.Dict.LLM = &LLMConfig{}
	}
	if cfg.Dict.LLM.Timeout == 0 {
		cfg.Dict.LLM.Timeout = Duration(30 * time.Second)
	}
	if cfg.Dict.LLM.MaxTokens == 0 {
		cfg.Dict.LLM.MaxTokens = 2000
	}
	if cfg.Dict.LLM.Temperature == 0 {
		cfg.Dict.LLM.Temperature = 0.3
	}
	if cfg.Dict.Youdao == nil {
		cfg.Dict.Youdao = &YoudaoConfig{}
	}
	if cfg.Dict.Ecdict == nil {
		cfg.Dict.Ecdict = &EcdictConfig{}
	}
	if cfg.Dict.Etymoline == nil {
		cfg.Dict.Etymoline = &EtymonlineConfig{}
	}
	if cfg.Dict.MWebster == nil {
		cfg.Dict.MWebster = &MWebsterConfig{}
	}
	if cfg.Dict.Google == nil {
		cfg.Dict.Google = &GoogleConfig{}
	}
	if cfg.Trans == nil {
		cfg.Trans = &TransConfig{}
	}
	if cfg.Trans.Default == "" {
		cfg.Trans.Default = "google"
	}
	if cfg.Trans.LLM == nil {
		cfg.Trans.LLM = &TransLLMConfig{}
	}
	if cfg.Trans.LLM.Timeout == 0 {
		cfg.Trans.LLM.Timeout = Duration(30 * time.Second)
	}
	if cfg.Trans.LLM.MaxTokens == 0 {
		cfg.Trans.LLM.MaxTokens = 2000
	}
	if cfg.Trans.LLM.Temperature == 0 {
		cfg.Trans.LLM.Temperature = 0.3
	}
	if cfg.Trans.Google == nil {
		cfg.Trans.Google = &TransGoogleConfig{}
	}
	if cfg.Notebook == nil {
		cfg.Notebook = &NotebookConfig{}
	}
	if cfg.Notebook.Default == "" {
		cfg.Notebook.Default = "default"
	}
	if cfg.Notebook.Settings == nil {
		cfg.Notebook.Settings = &NotebookSettings{}
	}
	if cfg.Notebook.Settings.MaxReviews == 0 {
		cfg.Notebook.Settings.MaxReviews = 50
	}
	if cfg.Notebook.Settings.NewCardsPerDay == 0 {
		cfg.Notebook.Settings.NewCardsPerDay = 20
	}

	if cfg.Notebook.Settings.BasePath == "" {
		cfg.Notebook.Settings.BasePath = filepath.Join(dir, "notebooks")
	}
	if cfg.Dict.Ecdict.DBFilename == "" {
		cfg.Dict.Ecdict.DBFilename = filepath.Join(dir, "stardict.db")
	}
}

func applyEnvOverrides(cfg *Config) {
	envMap := buildEnvMap()
	knownPaths := buildKnownPaths(reflect.ValueOf(cfg).Elem(), "")
	for envKey, envVal := range envMap {
		if yamlPath, ok := knownPaths[envKey]; ok {
			setValueByPath(cfg, yamlPath, envVal)
		}
	}
}

func buildEnvMap() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		if !strings.HasPrefix(key, "WORDFLOW_") {
			continue
		}
		envKey := strings.TrimPrefix(key, "WORDFLOW_")
		if len(parts) == 2 {
			result[envKey] = parts[1]
		} else {
			result[envKey] = ""
		}
	}
	return result
}

func buildKnownPaths(v reflect.Value, prefix string) map[string]string {
	result := make(map[string]string)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return result
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		field := v.Field(i)

		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		tagName := strings.SplitN(yamlTag, ",", 2)[0]
		if tagName == "" || tagName == "omitempty" {
			continue
		}

		fullPath := tagName
		if prefix != "" {
			fullPath = prefix + "." + tagName
		}
		envKey := strings.ToUpper(strings.ReplaceAll(fullPath, ".", "_"))

		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				continue
			}
			for k, v := range buildKnownPaths(field.Elem(), fullPath) {
				result[k] = v
			}
			result[envKey] = fullPath
			continue
		}

		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(Duration(0)) {
			for k, v := range buildKnownPaths(field, fullPath) {
				result[k] = v
			}
			result[envKey] = fullPath
			continue
		}

		result[envKey] = fullPath
	}
	return result
}

func setValueByPath(cfg *Config, path string, value string) {
	v := reflect.ValueOf(cfg).Elem()
	parts := strings.Split(path, ".")
	current := v

	for i, part := range parts {
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				current.Set(reflect.New(current.Type().Elem()))
			}
			current = current.Elem()
		}
		field, found := findFieldByYAMLTag(current, part)
		if !found {
			return
		}
		if i == len(parts)-1 {
			setFieldValueFromString(field, value)
			return
		}
		current = field
	}
}

func setFieldValueFromString(field reflect.Value, value string) {
	if !field.CanSet() {
		return
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		if intVal, err := strconv.Atoi(value); err == nil {
			field.SetInt(int64(intVal))
		}
	case reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	default:
		if field.Type() == reflect.TypeOf(Duration(0)) {
			if dur, err := time.ParseDuration(value); err == nil {
				field.Set(reflect.ValueOf(Duration(dur)))
			}
		}
	}
}

func findFieldByYAMLTag(v reflect.Value, tag string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		name := strings.SplitN(yamlTag, ",", 2)[0]
		if name == tag {
			return field, true
		}
	}
	return reflect.Value{}, false
}

func LoadConfig(configFilename string) (*Config, error) {
	data, err := os.ReadFile(configFilename)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read config file: %s", configFilename))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse config file: %s", configFilename))
	}

	applyDefaults(&cfg, configFilename)

	applyEnvOverrides(&cfg)

	return &cfg, nil
}

func (c *Config) Validate(activeEndpoint string) error {
	if c.Version != DefaultConfigVersion {
		return fmt.Errorf("unsupported config version: %q. Run 'wordflow config init' to regenerate your config", c.Version)
	}
	if activeEndpoint == "llm" {
		if err := c.Dict.LLM.Validate(); err != nil {
			return err
		}
	}
	if activeEndpoint == "ecdict" {
		if err := c.Dict.Ecdict.Validate(); err != nil {
			return err
		}
	}
	if activeEndpoint == "mwebster" {
		if err := c.Dict.MWebster.Validate(); err != nil {
			return err
		}
	}
	if err := c.Notebook.Settings.Validate(); err != nil {
		return err
	}
	return nil
}

func ValidateForTrans(cfg *Config) error {
	if cfg.Version != DefaultConfigVersion {
		return fmt.Errorf("unsupported config version: %q. Run 'wordflow config init' to regenerate your config", cfg.Version)
	}
	if cfg.Trans == nil {
		return errors.New("trans config is missing")
	}
	endpointConfig, err := cfg.Trans.GetEndpointConfig(cfg.Trans.Default)
	if err != nil {
		return err
	}
	return endpointConfig.Validate()
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create config dir: %s", dir))
	}
	return os.WriteFile(c.Common.ConfigFilename, bytes, 0644)
}