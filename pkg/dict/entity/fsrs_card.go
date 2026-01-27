package entity

import (
	"time"

	"github.com/gogodjzhu/word-flow/pkg/dict/fsrs"
	"gorm.io/gorm"
)

// FSRSCard represents an FSRS card in the database
type FSRSCard struct {
	WordId        string         `gorm:"word_id;primaryKey" json:"word_id"`
	Notebook      string         `gorm:"notebook" json:"notebook"`
	Due           time.Time      `gorm:"due" json:"due"`
	Stability     float64        `gorm:"stability" json:"stability"`
	Difficulty    float64        `gorm:"difficulty" json:"difficulty"`
	ElapsedDays   uint64         `gorm:"elapsed_days" json:"elapsed_days"`
	ScheduledDays uint64         `gorm:"scheduled_days" json:"scheduled_days"`
	Reps          uint64         `gorm:"reps" json:"reps"`
	Lapses        uint64         `gorm:"lapses" json:"lapses"`
	State         int8           `gorm:"state" json:"state"`
	LastReview    time.Time      `gorm:"last_review" json:"last_review"`
	CreatedAt     time.Time      `gorm:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"deleted_at" json:"-"`
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
