package inputs

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	err       error
	userInput string
	title     string
}

func initialModel(defaultValue, title string) model {
	ti := textinput.New()
	ti.Placeholder = defaultValue
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30

	return model{
		textInput: ti,
		title:     title,
		userInput: defaultValue,
		err:       nil,
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
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.title,
		m.textInput.View(),
		"Press Enter to submit",
	) + "\n"
}

func Run(defaultValue, title string) (string, error) {
	initModel := initialModel(defaultValue, title)
	p := tea.NewProgram(initModel)

	teaModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := teaModel.(model); ok {
		return m.textInput.Value(), nil
	}

	return "", fmt.Errorf("unexpected model type")
}
