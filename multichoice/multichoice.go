package multichoice

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	list         list.Model
	packages     []string
	descriptions []string
	selected     map[int]struct{}
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func NewModel(packages, descriptions []string, title string) Model {
	// Create items from packages and descriptions
	items := make([]list.Item, len(packages))
	selected := make(map[int]struct{})

	for i := range packages {
		items[i] = item{title: packages[i], desc: descriptions[i]}
		selected[i] = struct{}{} // All packages pre-selected
	}

	// Create the list UI model
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title

	return Model{
		list:         l,
		packages:     packages,
		descriptions: descriptions,
		selected:     selected,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Return selected packages
			var selectedPackages []string
			for index := range m.selected {
				selectedPackages = append(selectedPackages, m.packages[index])
			}
			return m, tea.Quit
		case "up", "k":
			m.list.CursorUp()
		case "down", "j":
			m.list.CursorDown()
		case " ", "x":
			// Toggle selection of the current package
			index := m.list.Index()
			if _, ok := m.selected[index]; ok {
				delete(m.selected, index)
			} else {
				m.selected[index] = struct{}{}
			}
		}
	}
	return m, nil
}

var (
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))           // Pinkish for cursor
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true) // Green for selected items
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))            // White for non-selected items
	descStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))           // Light grey for description
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Bold(true) // Blue for title
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.list.Title) + "\n\n")

	// Determine the maximum width of package titles for proper alignment
	maxTitleLength := 0
	for _, listItem := range m.list.Items() {
		pkg, ok := listItem.(item)
		if !ok {
			continue
		}
		if len(pkg.Title()) > maxTitleLength {
			maxTitleLength = len(pkg.Title())
		}
	}

	// Render each item with aligned descriptions
	for i, listItem := range m.list.Items() {
		pkg, ok := listItem.(item)
		if !ok {
			continue
		}

		cursor := " " // no cursor
		if m.list.Index() == i {
			cursor = cursorStyle.Render(">")
		}

		checked := " " // not selected
		var itemStyle lipgloss.Style
		if _, ok := m.selected[i]; ok {
			checked = selectedStyle.Render("✓") // selected
			itemStyle = selectedStyle
		} else {
			itemStyle = normalStyle
		}

		// Render the package title and description with proper padding
		title := itemStyle.Render(pkg.Title())
		padding := strings.Repeat(" ", maxTitleLength-len(pkg.Title())+1) // 1 space between title and description
		b.WriteString(fmt.Sprintf("%s [%s] %s%s- %s\n", cursor, checked, title, padding, descStyle.Render(pkg.Description())))
	}

	b.WriteString("\n" + normalStyle.Render("Press Enter to confirm selection.") + "\n")

	return b.String()
}

// Run the multi-select program and return the selected packages.
func Run(packages, descriptions []string, title string) ([]string, error) {
	if len(packages) != len(descriptions) {
		return nil, fmt.Errorf("mismatched package and description count")
	}
	p := tea.NewProgram(NewModel(packages, descriptions, title))
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	// Collect selected packages.
	var selectedPackages []string
	for index := range m.(Model).selected {
		selectedPackages = append(selectedPackages, packages[index])
	}
	return selectedPackages, nil
}
