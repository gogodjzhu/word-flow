package fsrs

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs"
)

// Scheduler wraps the FSRS algorithm for vocabulary exam
type Scheduler struct {
	params fsrs.Parameters
}

// NewScheduler creates a new FSRS scheduler with default parameters
func NewScheduler() *Scheduler {
	return &Scheduler{
		params: fsrs.DefaultParam(),
	}
}

// NewSchedulerWithParams creates a new FSRS scheduler with custom parameters
func NewSchedulerWithParams(params fsrs.Parameters) *Scheduler {
	return &Scheduler{
		params: params,
	}
}

// Card represents a vocabulary card with FSRS scheduling data
type Card struct {
	WordId        string    `json:"word_id" yaml:"word_id"`
	Notebook      string    `json:"notebook" yaml:"notebook"`
	Due           time.Time `json:"due" yaml:"due"`
	Stability     float64   `json:"stability" yaml:"stability"`
	Difficulty    float64   `json:"difficulty" yaml:"difficulty"`
	ElapsedDays   uint64    `json:"elapsed_days" yaml:"elapsed_days"`
	ScheduledDays uint64    `json:"scheduled_days" yaml:"scheduled_days"`
	Reps          uint64    `json:"reps" yaml:"reps"`
	Lapses        uint64    `json:"lapses" yaml:"lapses"`
	State         State     `json:"state" yaml:"state"`
	LastReview    time.Time `json:"last_review" yaml:"last_review"`
}

// State represents the learning state of a card
type State int8

const (
	New State = iota
	Learning
	Review
	Relearning
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case New:
		return "New"
	case Learning:
		return "Learning"
	case Review:
		return "Review"
	case Relearning:
		return "Relearning"
	default:
		return "Unknown"
	}
}

// Rating represents the user's rating for a card review
type Rating int8

const (
	Again Rating = iota + 1 // 1 - Complete failure
	Hard                    // 2 - Difficult recall
	Good                    // 3 - Moderate effort
	Easy                    // 4 - Very easy
)

// String returns the string representation of the rating
func (r Rating) String() string {
	switch r {
	case Again:
		return "Again"
	case Hard:
		return "Hard"
	case Good:
		return "Good"
	case Easy:
		return "Easy"
	default:
		return "Unknown"
	}
}

// ToFSRSCard converts our Card to FSRS Card
func (c *Card) ToFSRSCard() fsrs.Card {
	return fsrs.Card{
		Due:           c.Due,
		Stability:     c.Stability,
		Difficulty:    c.Difficulty,
		ElapsedDays:   c.ElapsedDays,
		ScheduledDays: c.ScheduledDays,
		Reps:          c.Reps,
		Lapses:        c.Lapses,
		State:         fsrs.State(c.State),
		LastReview:    c.LastReview,
	}
}

// FromFSRSCard updates our Card from FSRS Card
func (c *Card) FromFSRSCard(fsrsCard fsrs.Card) {
	c.Due = fsrsCard.Due
	c.Stability = fsrsCard.Stability
	c.Difficulty = fsrsCard.Difficulty
	c.ElapsedDays = fsrsCard.ElapsedDays
	c.ScheduledDays = fsrsCard.ScheduledDays
	c.Reps = fsrsCard.Reps
	c.Lapses = fsrsCard.Lapses
	c.State = State(fsrsCard.State)
	c.LastReview = fsrsCard.LastReview
}

// NewCard creates a new card for a given word
func NewCard(wordId, notebook string) *Card {
	fsrsCard := fsrs.NewCard()

	return &Card{
		WordId:        wordId,
		Notebook:      notebook,
		Due:           fsrsCard.Due,
		Stability:     fsrsCard.Stability,
		Difficulty:    fsrsCard.Difficulty,
		ElapsedDays:   fsrsCard.ElapsedDays,
		ScheduledDays: fsrsCard.ScheduledDays,
		Reps:          fsrsCard.Reps,
		Lapses:        fsrsCard.Lapses,
		State:         State(fsrsCard.State),
		LastReview:    fsrsCard.LastReview,
	}
}

// IsDue checks if the card is due for review
func (c *Card) IsDue() bool {
	return time.Now().After(c.Due) || time.Now().Equal(c.Due)
}

// IsNew checks if the card is in the new state
func (c *Card) IsNew() bool {
	return c.State == New
}

// IsLearning checks if the card is in the learning state
func (c *Card) IsLearning() bool {
	return c.State == Learning
}

// IsReview checks if the card is in the review state
func (c *Card) IsReview() bool {
	return c.State == Review
}

// Repeat calculates the next state for all possible ratings
func (s *Scheduler) Repeat(card *Card, now time.Time) map[Rating]*Card {
	fsrsCard := card.ToFSRSCard()
	schedulingCards := s.params.Repeat(fsrsCard, now)

	result := make(map[Rating]*Card)
	for rating, schedulingInfo := range schedulingCards {
		newCard := &Card{
			WordId:   card.WordId,
			Notebook: card.Notebook,
		}
		newCard.FromFSRSCard(schedulingInfo.Card)
		result[Rating(rating)] = newCard
	}

	return result
}

// Next calculates the next state for a specific rating
func (s *Scheduler) Next(card *Card, now time.Time, rating Rating) *Card {
	fsrsCard := card.ToFSRSCard()
	schedulingCards := s.params.Repeat(fsrsCard, now)
	schedulingInfo := schedulingCards[fsrs.Rating(rating)]

	newCard := &Card{
		WordId:   card.WordId,
		Notebook: card.Notebook,
	}
	newCard.FromFSRSCard(schedulingInfo.Card)

	return newCard
}
