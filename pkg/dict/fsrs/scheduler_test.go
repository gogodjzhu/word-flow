package fsrs

import (
	"testing"
	"time"
)

func TestNewCard(t *testing.T) {
	wordId := "test-word"
	notebook := "test-notebook"

	card := NewCard(wordId, notebook)

	if card.WordId != wordId {
		t.Errorf("Expected wordId %s, got %s", wordId, card.WordId)
	}

	if card.Notebook != notebook {
		t.Errorf("Expected notebook %s, got %s", notebook, card.Notebook)
	}

	if !card.IsNew() {
		t.Error("Expected new card to be in New state")
	}

	if !card.IsDue() {
		t.Error("Expected new card to be due for review")
	}
}

func TestSchedulerRepeat(t *testing.T) {
	scheduler := NewScheduler()
	card := NewCard("test", "test")

	// Test repeat for all ratings
	now := time.Now()
	nextCards := scheduler.Repeat(card, now)

	if len(nextCards) != 4 {
		t.Errorf("Expected 4 rating options, got %d", len(nextCards))
	}

	// Check that each rating exists
	ratings := []Rating{Skip, Hard, Good, Easy}
	for _, rating := range ratings {
		if _, exists := nextCards[rating]; !exists {
			t.Errorf("Missing card for rating %v", rating)
		}
	}

	// Check that Easy has longer interval than Skip
	skipCard := nextCards[Skip]
	easyCard := nextCards[Easy]

	if easyCard.Due.Before(skipCard.Due) {
		t.Error("Expected Easy rating to have longer interval than Skip")
	}
}

func TestCardStates(t *testing.T) {
	// Test card state methods
	card := &Card{
		WordId:   "test-word",
		Notebook: "test-notebook",
		State:    Review,
	}

	if !card.IsReview() {
		t.Error("Card should be in review state")
	}

	if card.IsNew() {
		t.Error("Card should not be in new state")
	}

	if card.IsLearning() {
		t.Error("Card should not be in learning state")
	}

	// Test new card
	newCard := NewCard("new-word", "test")
	if !newCard.IsNew() {
		t.Error("New card should be in new state")
	}

	if !newCard.IsDue() {
		t.Error("New card should be due for review")
	}
}

func TestRatingString(t *testing.T) {
	tests := map[Rating]string{
		Skip: "Skip",
		Hard: "Hard",
		Good: "Good",
		Easy: "Easy",
	}

	for rating, expected := range tests {
		if rating.String() != expected {
			t.Errorf("Expected rating string %s, got %s", expected, rating.String())
		}
	}
}

func TestStateString(t *testing.T) {
	tests := map[State]string{
		New:        "New",
		Learning:   "Learning",
		Review:     "Review",
		Relearning: "Relearning",
	}

	for state, expected := range tests {
		if state.String() != expected {
			t.Errorf("Expected state string %s, got %s", expected, state.String())
		}
	}
}
