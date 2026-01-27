package tui_exam

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/gogodjzhu/word-flow/pkg/dict/entity"
	"github.com/gogodjzhu/word-flow/pkg/dict/fsrs"
)

// keyMap defines key bindings for the exam interface
type keyMap struct {
	RateAgain key.Binding
	RateHard  key.Binding
	RateGood  key.Binding
	RateEasy  key.Binding
	ShowDef   key.Binding
	ShowEx    key.Binding
	Skip      key.Binding
	Quit      key.Binding
	Help      key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() keyMap {
	return keyMap{
		RateAgain: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "Again (failure)"),
		),
		RateHard: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "Hard (difficult)"),
		),
		RateGood: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "Good (moderate)"),
		),
		RateEasy: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "Easy (very easy)"),
		),
		ShowDef: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "Toggle definition"),
		),
		ShowEx: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "Toggle examples"),
		),
		Skip: key.NewBinding(
			key.WithKeys("s", "tab"),
			key.WithHelp("s/Tab", "Skip word"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/Esc", "Quit exam"),
		),
		Help: key.NewBinding(
			key.WithKeys("h", "?"),
			key.WithHelp("h/?", "Show help"),
		),
	}
}

// Styles for the exam interface
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	wordStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			MarginTop(1)

	definitionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginTop(1)

	exampleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Italic(true).
			MarginTop(1)

	ratingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginTop(2)

	selectedRatingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#FFFFFF")).
				Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

// Model represents the exam TUI model
type Model struct {
	keys       keyMap
	words      []*entity.WordNote
	currentIdx int
	showDef    bool
	showEx     bool
	showHelp   bool
	completed  int
	skipped    int
	startTime  time.Time
	scheduler  *fsrs.Scheduler
	width      int
	height     int
	quitting   bool
}

// NewModel creates a new exam model
func NewModel(words []*entity.WordNote, scheduler *fsrs.Scheduler) Model {
	return Model{
		keys:       DefaultKeyMap(),
		words:      words,
		currentIdx: 0,
		showDef:    false, // Hidden by default
		showEx:     false, // Hidden by default
		showHelp:   false,
		completed:  0,
		skipped:    0,
		startTime:  time.Now(),
		scheduler:  scheduler,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.ShowDef):
			m.showDef = !m.showDef
			return m, nil

		case key.Matches(msg, m.keys.ShowEx):
			m.showEx = !m.showEx
			return m, nil

		case key.Matches(msg, m.keys.Skip):
			m.skipped++
			m.nextWord()
			return m, nil

		case key.Matches(msg, m.keys.RateAgain):
			return m.rateWord(fsrs.Again)

		case key.Matches(msg, m.keys.RateHard):
			return m.rateWord(fsrs.Hard)

		case key.Matches(msg, m.keys.RateGood):
			return m.rateWord(fsrs.Good)

		case key.Matches(msg, m.keys.RateEasy):
			return m.rateWord(fsrs.Easy)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// rateWord processes the rating and moves to the next word
func (m Model) rateWord(rating fsrs.Rating) (tea.Model, tea.Cmd) {
	if m.currentIdx >= len(m.words) {
		return m, nil
	}

	currentWord := m.words[m.currentIdx]
	if currentWord.FSRSCard == nil {
		newCard := fsrs.NewCard(currentWord.WordItemId, "")
		currentWord.FSRSCard = &entity.FSRSCard{}
		currentWord.FSRSCard.FromFSRSCard(newCard)
	}

	// Update FSRS card if it exists
	if currentWord.FSRSCard != nil {
		card := currentWord.FSRSCard.ToFSRSCard()
		nextCard := m.scheduler.Next(card, time.Now(), rating)
		currentWord.FSRSCard.FromFSRSCard(nextCard)
		currentWord.LastRating = int(rating)
		currentWord.NextReview = nextCard.Due.Unix()
	}

	m.completed++
	m.nextWord()
	return m, nil
}

// nextWord moves to the next word
func (m *Model) nextWord() {
	m.currentIdx++
	if m.currentIdx >= len(m.words) {
		m.quitting = true
	}
}

// View renders the model
func (m Model) View() string {
	if m.quitting {
		return m.renderSummary()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if len(m.words) == 0 || m.currentIdx >= len(m.words) {
		return m.renderSummary()
	}

	// Ensure width is initialized
	if m.width <= 0 {
		return "Initializing..."
	}

	// Safety check
	if m.currentIdx >= len(m.words) {
		return m.renderSummary()
	}

	currentWord := m.words[m.currentIdx]

	var content string

	// Title with progress
	title := fmt.Sprintf("Vocabulary Review: %d/%d (%.1f%%)",
		m.currentIdx+1, len(m.words),
		float64(m.currentIdx+1)/float64(len(m.words))*100)
	content += titleStyle.Render(title) + "\n\n"

	// Current word
	content += wordStyle.Render(currentWord.Word) + "\n"

	// Separator
	separatorWidth := min(max(m.width-4, 10), 50)
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render(strings.Repeat("─", separatorWidth)) + "\n"

	// Definition (hidden by default)
	if m.showDef {
		content += definitionStyle.Render(currentWord.GetDefinition()) + "\n"
	} else {
		content += definitionStyle.Render("[Press 'd' to show Definition]") + "\n"
	}

	// Examples (hidden by default)
	if m.showEx {
		examples := currentWord.GetExamples()
		if len(examples) > 0 {
			for _, example := range examples {
				content += "\n" + exampleStyle.Render("  • "+example)
			}
			content += "\n"
		}
	} else {
		content += exampleStyle.Render("[Press 'e' to show Examples]") + "\n"
	}

	// Separator
	separatorWidth2 := min(max(m.width-4, 10), 50)
	content += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render(strings.Repeat("─", separatorWidth2)) + "\n"

	// Rating options
	content += ratingStyle.Render("How well did you know this word?") + "\n"
	ratings := []struct {
		rating fsrs.Rating
		desc   string
	}{
		{fsrs.Again, "Complete failure"},
		{fsrs.Hard, "Difficult recall"},
		{fsrs.Good, "Moderate effort"},
		{fsrs.Easy, "Very easy"},
	}

	for i, r := range ratings {
		prefix := ""
		ratingNum := i + 1
		ratingName := "Again"
		switch r.rating {
		case fsrs.Again:
			ratingName = "Again"
		case fsrs.Hard:
			ratingName = "Hard"
		case fsrs.Good:
			ratingName = "Good"
		case fsrs.Easy:
			ratingName = "Easy"
		}
		content += fmt.Sprintf("%s[%d] %s - %s\n", prefix, ratingNum, ratingName, r.desc)
	}

	// Info section
	if currentWord.FSRSCard != nil {
		lastReview := time.Unix(currentWord.FSRSCard.LastReview.Unix(), 0)
		info := fmt.Sprintf("Last reviewed: %s | Streak: %d days",
			humanize.Time(lastReview),
			m.completed)
		content += infoStyle.Render(info) + "\n"
	}

	// Help line
	content += helpStyle.Render("[1-4: Rate] [d: Definition] [e: Examples] [s: Skip] [h: Help] [q: Quit]")

	// Apply container style
	container := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return container.Render(content)
}

// renderSummary renders the session summary
func (m Model) renderSummary() string {
	duration := time.Since(m.startTime)

	content := titleStyle.Render("Session Complete") + "\n\n"

	content += fmt.Sprintf("✅ Reviewed %d words this session\n", m.completed)
	if m.skipped > 0 {
		content += fmt.Sprintf("⏭️  Skipped %d words\n", m.skipped)
	}

	if m.completed > 0 {
		successRate := float64(m.completed) / float64(m.completed+m.skipped) * 100
		content += fmt.Sprintf("📊 Success Rate: %.1f%%\n", successRate)
	}

	content += fmt.Sprintf("⏱️  Duration: %s\n", duration.Round(time.Second))

	content += "\n" + helpStyle.Render("[Enter: Continue] [q: Quit]")

	container := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return container.Render(content)
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	content := titleStyle.Render("Help - Vocabulary Review") + "\n\n"

	content += ratingStyle.Render("Rating Options:") + "\n"
	content += "  [1] Again - Complete failure, reset to learning\n"
	content += "  [2] Hard  - Difficult recall with hesitation\n"
	content += "  [3] Good  - Moderate effort, standard interval\n"
	content += "  [4] Easy  - Very easy recall, longer interval\n\n"

	content += ratingStyle.Render("Controls:") + "\n"
	content += "  [d] Toggle definition visibility\n"
	content += "  [e] Toggle examples visibility\n"
	content += "  [s] Skip current word (review later)\n"
	content += "  [h/?] Show this help\n"
	content += "  [q/Esc] Exit review session\n\n"

	content += helpStyle.Render("[Press any key to return to review]")

	container := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return container.Render(content)
}

// GetResults returns the session results
func (m Model) GetResults() ExamResults {
	var words []*entity.WordNote
	if m.words != nil && len(m.words) > 0 && m.completed > 0 {
		maxWords := min(m.completed, len(m.words))
		words = make([]*entity.WordNote, maxWords)
		copy(words, m.words[:maxWords])
	}

	return ExamResults{
		Completed: m.completed,
		Skipped:   m.skipped,
		Duration:  time.Since(m.startTime),
		Words:     words,
	}
}

// ExamResults represents the results of an exam session
type ExamResults struct {
	Completed int
	Skipped   int
	Duration  time.Duration
	Words     []*entity.WordNote
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
