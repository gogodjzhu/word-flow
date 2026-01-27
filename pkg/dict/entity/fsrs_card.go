package entity

import (
	"time"

	"github.com/gogodjzhu/word-flow/pkg/dict/fsrs"
	"gorm.io/gorm"
)

// FSRSCard represents an FSRS card in the database
type FSRSCard struct {
	WordId        string         `gorm:"word_id;primaryKey" json:"word_id" yaml:"word_id"`
	Notebook      string         `gorm:"notebook" json:"notebook" yaml:"notebook"`
	Due           time.Time      `gorm:"due" json:"due" yaml:"due"`
	Stability     float64        `gorm:"stability" json:"stability" yaml:"stability"`
	Difficulty    float64        `gorm:"difficulty" json:"difficulty" yaml:"difficulty"`
	ElapsedDays   uint64         `gorm:"elapsed_days" json:"elapsed_days" yaml:"elapsed_days"`
	ScheduledDays uint64         `gorm:"scheduled_days" json:"scheduled_days" yaml:"scheduled_days"`
	Reps          uint64         `gorm:"reps" json:"reps" yaml:"reps"`
	Lapses        uint64         `gorm:"lapses" json:"lapses" yaml:"lapses"`
	State         int8           `gorm:"state" json:"state" yaml:"state"`
	LastReview    time.Time      `gorm:"last_review" json:"last_review" yaml:"last_review"`
	CreatedAt     time.Time      `gorm:"created_at" json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time      `gorm:"updated_at" json:"updated_at" yaml:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"deleted_at" json:"-" yaml:"-"`
}

// TableName returns the table name for FSRSCard
func (FSRSCard) TableName() string {
	return "fsrs_card"
}

// ToFSRSCard converts FSRSCard to domain fsrs.Card
func (f *FSRSCard) ToFSRSCard() *fsrs.Card {
	return &fsrs.Card{
		WordId:        f.WordId,
		Notebook:      f.Notebook,
		Due:           f.Due,
		Stability:     f.Stability,
		Difficulty:    f.Difficulty,
		ElapsedDays:   f.ElapsedDays,
		ScheduledDays: f.ScheduledDays,
		Reps:          f.Reps,
		Lapses:        f.Lapses,
		State:         fsrs.State(f.State),
		LastReview:    f.LastReview,
	}
}

// FromFSRSCard updates FSRSCard from domain fsrs.Card
func (f *FSRSCard) FromFSRSCard(card *fsrs.Card) {
	f.WordId = card.WordId
	f.Notebook = card.Notebook
	f.Due = card.Due
	f.Stability = card.Stability
	f.Difficulty = card.Difficulty
	f.ElapsedDays = card.ElapsedDays
	f.ScheduledDays = card.ScheduledDays
	f.Reps = card.Reps
	f.Lapses = card.Lapses
	f.State = int8(card.State)
	f.LastReview = card.LastReview
}
