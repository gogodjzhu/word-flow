package tui_list

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type OptionEntity interface {
	Entity() interface{}
	Title() string
	Description() string
}

type item struct {
	entity interface{}
	title  string
	desc   string
}

func NewOption(entity OptionEntity) OptionEntity {
	return item{
		entity: entity.Entity(),
		title:  entity.Title(),
		desc:   entity.Description(),
	}
}

type CallbackFunc struct {
	Keys             []string
	ShortDescription string
	FullDescription  string
	Callback         func(selectedOption OptionEntity) []OptionEntity
}

func (i item) Entity() interface{} { return i.entity }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type app struct {
	list             list.Model
	key2callbackFunc map[string]func(option item) []item
}

func NewApp(title string, options []OptionEntity, callbacks []CallbackFunc) tea.Model {
	var items []list.Item
	for _, option := range options {
		items = append(items, option.(item))
	}

	// Setup list
	groceryList := list.New(items, newItemDelegate(callbacks), 0, 0)
	groceryList.Title = title
	groceryList.AdditionalFullHelpKeys = func() []key.Binding {
		var bindings []key.Binding
		for _, callback := range callbacks {
			binding := key.NewBinding(
				key.WithKeys(callback.Keys...),
				key.WithHelp(callback.Keys[0], callback.ShortDescription))
			bindings = append(bindings, binding)
		}
		return bindings
	}
	return app{
		list: groceryList,
	}
}

func newItemDelegate(callbacks []CallbackFunc) list.ItemDelegate {
	key2Callback := make(map[string]func(option OptionEntity) []OptionEntity)
	for _, callback := range callbacks {
		for _, key := range callback.Keys {
			// if duplicate
			if _, ok := key2Callback[key]; ok {
				panic("duplicate app key:" + key)
			}
			key2Callback[key] = callback.Callback
		}
	}
	d := list.NewDefaultDelegate()
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var selectedOption item
		if i, ok := m.SelectedItem().(item); ok {
			selectedOption = i
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if callback, ok := key2Callback[msg.String()]; ok {
				if updateOptions := callback(selectedOption); updateOptions != nil {
					var items []list.Item
					for _, opt := range updateOptions {
						items = append(items, opt.(item))
					}
					return m.SetItems(items)
				}
			}
		}

		return nil
	}
	return d
}

func (a app) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		a.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if a.list.FilterState() == list.Filtering {
			break
		}
	}

	var cmds []tea.Cmd

	// This will also call our delegate's update function.
	newListModel, cmd := a.list.Update(msg)
	a.list = newListModel
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a app) View() string {
	return docStyle.Render(a.list.View())
}
