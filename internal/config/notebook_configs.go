package config

import (
	"errors"
	"path/filepath"
)

// NotebookSettings notebook配置
type NotebookSettings struct {
	Default        string `yaml:"-" json:"-"`
	BasePath       string `yaml:"notebook.basepath" json:"notebook.basepath"`
	MaxReviews     int    `yaml:"fsrs.max_reviews_per_session" json:"fsrs.max_reviews_per_session"`
	NewCardsPerDay int    `yaml:"fsrs.new_cards_per_day" json:"fsrs.new_cards_per_day"`
}

func (c *NotebookSettings) Validate() error {
	if c.BasePath == "" {
		return errors.New("notebook.basepath is required")
	}
	if c.MaxReviews <= 0 {
		return errors.New("fsrs.max_reviews_per_session must be positive")
	}
	if c.NewCardsPerDay < 0 {
		return errors.New("fsrs.new_cards_per_day must be non-negative")
	}
	return nil
}

func (c *NotebookSettings) GetDefaults() interface{} {
	return map[string]interface{}{
		NotebookConfigNotebookBasepath: filepath.Join(configDir(), "notebooks"),
		NotebookConfigFSRSMaxReviews:   50,
		NotebookConfigFSRSLimit:        20,
	}
}

func (nc *NotebookConfig) GetConfig() (*NotebookSettings, error) {
	conf := &NotebookSettings{Default: nc.Default}
	mapped, err := globalMapper.MapToConfig(conf, nc.Parameters)
	if err != nil {
		return nil, err
	}
	return mapped.(*NotebookSettings), nil
}
