package tui_textinput

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	title        string
	textInput    textinput.Model
	callbackFunc func(string)
}

func NewModel(title, placeholder string, callbackFunc func(string)) tea.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	//ti.CharLimit = 156
	//ti.Width = 20

	return model{
		title:        title,
		textInput:    ti,
		callbackFunc: callbackFunc,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.callbackFunc(m.textInput.Value())
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.callbackFunc("")
			return m, tea.Quit
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.title,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
