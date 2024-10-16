package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil/tui/tui_list"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil/tui/tui_result"
	"github.com/gogodjzhu/word-flow/pkg/cmdutil/tui/tui_textinput"
	"os"
)

func main() {
	//TestResult()
	//TestTextInput()
	TestList()
}

func TestResult() {
	red := color.New(color.FgRed).SprintFunc()
	p := tea.NewProgram(tui_result.NewModel([]string{"a", "b", "c"}, red("please select..."), func(s string) {
		fmt.Println("selected:", s)
	}))

	// Run returns the model as a tea.Model.
	_, err := p.Run()
	if err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
}

func TestTextInput() {
	p := tea.NewProgram(tui_textinput.NewModel("Please input...", "placeholder", func(s string) {
		fmt.Println("You input:", s)
	}))
	if _, err := p.Run(); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
}

func TestList() {
	options := []tui_list.OptionEntity{
		tui_list.NewOption(MyFruitEntity{
			Name: "apple",
			Desc: "apple is good",
		}),
		tui_list.NewOption(MyFruitEntity{
			Name: "orange",
			Desc: "orange is good",
		}),
		tui_list.NewOption(MyFruitEntity{
			Name: "banana",
			Desc: "banana is good",
		}),
	}
	p := tea.NewProgram(tui_list.NewApp("test", options, []tui_list.CallbackFunc{
		{
			Keys: []string{"h"},
			Callback: func(option tui_list.OptionEntity) []tui_list.OptionEntity {
				return options
			},
			ShortDescription: "hKey",
			FullDescription:  "h for a key full description",
		},
		{
			Keys: []string{"n"},
			Callback: func(option tui_list.OptionEntity) []tui_list.OptionEntity {
				return options
			},
			ShortDescription: "nKey",
			FullDescription:  "n for a key full description",
		},
		{
			Keys: []string{"a"},
			Callback: func(option tui_list.OptionEntity) []tui_list.OptionEntity {
				options = append(options, tui_list.NewOption(MyFruitEntity{
					Name: "a",
					Desc: "a is good",
				}))
				return options
			},
			ShortDescription: "aKey",
			FullDescription:  "a for a key full description",
		},
		{
			Keys: []string{"x"},
			Callback: func(option tui_list.OptionEntity) []tui_list.OptionEntity {
				// remove option from options by title
				newOptions := make([]tui_list.OptionEntity, 0)
				for i, o := range options {
					if o.Title() != option.Title() {
						newOptions = append(newOptions, options[i])
					}
				}
				options = newOptions
				return newOptions
			},
			ShortDescription: "xKey",
			FullDescription:  "x for a key full description",
		},
	}))
	if _, err := p.Run(); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
}

type MyFruitEntity struct {
	Name interface{}
	Desc string
}

func (t MyFruitEntity) Entity() interface{} {
	return t.Name
}

func (t MyFruitEntity) Title() string {
	return t.Name.(string)
}

func (t MyFruitEntity) Description() string {
	return t.Desc
}
