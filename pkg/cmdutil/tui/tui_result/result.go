package tui_result

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	title        string
	cursor       int
	choices      []string
	callbackFunc func(string)
}

func NewModel(choices []string, title string, fn func(string)) tea.Model {
	return model{
		title:        title,
		choices:      choices,
		callbackFunc: fn,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			// Send the choice on the channel and exit.
			choice := m.choices[m.cursor]
			m.callbackFunc(choice)
			return m, tea.Quit
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString(m.title + "\n")
	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(m.choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")
	return s.String()
}
