package unichoice

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	choices []string
	cursor  int
	title   string
}

func New(choices []string, title string) Model {

	return Model{
		choices: choices,
		cursor:  0,
		title:   title,
	}
}

func (m Model) Data() string {
	return m.choices[m.cursor]
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}

		case "j", "down":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "enter":
			return m, tea.Quit
		}
	}

	return m, nil
}

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))           // Pinkish for cursor
	normalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))            // White for non-selected items
	descStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))           // Light grey for description
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true) // Blue for title
)

func (m Model) View() string {
	var s string

	// Optional title for the list
	s += titleStyle.Render("Select an option:") + "\n\n"

	// Loop over the choices and render them
	for i, choice := range m.choices {
		// Determine if the current item is the selected one
		cursor := " "        // No arrow by default
		style := normalStyle // Default style for non-selected items

		if i == m.cursor {
			// Highlight the current cursor with an arrow and a different style
			cursor = "→"
			style = cursorStyle
		}

		s += style.Render(cursor+" "+choice) + "\n"
	}

	// Display the quit instruction at the bottom
	s += "\n" + descStyle.Render("(j/down, k/up)")

	return s
}

func Run(choices []string, title string) (string, error) {
	p := tea.NewProgram(New(choices, title))
	tm, err := p.Run()
	if err != nil {
		return "", err
	}
	m := tm.(Model)
	return m.Data(), nil
}
