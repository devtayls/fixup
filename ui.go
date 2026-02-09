package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

// keyMap defines the key bindings
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Quit   key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "create fixup"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	list     list.Model
	selected bool
	err      error
}

// initialModel creates the initial model with commits
func initialModel(commits []Commit) model {
	// Convert []Commit to []list.Item
	items := make([]list.Item, len(commits))
	for i, commit := range commits {
		items[i] = commit
	}

	// Create the delegate
	delegate := commitDelegate{width: 120} // Default width, will be updated

	// Create the list
	l := list.New(items, delegate, 120, 20) // width, height - will be updated by WindowSizeMsg
	l.Title = "Select a commit to fixup"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return model{
		list: l,
	}
}

// Init is called when the program starts
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("Update called with message type: %T", msg)

	switch msg := msg.(type) {
	// Update window width
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4) // for margins
		log.Printf("Window resized: %dx%d", msg.Width, msg.Height)

	// Update key press
	case tea.KeyMsg:
		log.Printf("Key pressed: %s", msg.String())

		// Handle Enter specially (create fixup commit)
		if key.Matches(msg, keys.Select) {
			// Get selected item
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				commit := selectedItem.(Commit)
				if err := createFixupCommit(commit.Hash); err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.selected = true
				return m, tea.Quit
			}
		}

		// Handle Quit
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

	}
	// Let the list handle all other messages (navigation, etc.)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func wrapText(text string, maxWidth int) []string {
	// Guard against invalid widths
	if maxWidth <= 0 {
		return []string{text}
	}

	// If commit message is less than max width, don't change anything:
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	for len(text) > maxWidth {
		// greedily grab words till max width
		chunk := text[:maxWidth]
		wordBoundary := strings.LastIndex(chunk, " ")

		// Default to maxwidth (charbased) if wordBoundary found, use that instead.
		breakPoint := maxWidth
		if wordBoundary != -1 {
			breakPoint = wordBoundary
		}

		singleLine := text[:breakPoint]

		// syntax to grab from breakPoint to the end
		remaining := text[breakPoint:]
		remaining = strings.TrimSpace(remaining)

		// Append the lines to our array
		lines = append(lines, singleLine)

		// re-assign text for next loop
		text = remaining
	}

	if len(text) > 0 {
		lines = append(lines, text)
	}

	return lines
}

type commitDelegate struct {
	width int // terminal width for wrapping
}

func (d commitDelegate) Height() int {
	return 1 // base height
}

func (d commitDelegate) Spacing() int {
	return 0
}

func (d commitDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d commitDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	commit, ok := listItem.(Commit)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	log.Printf("Delegate rendering commit %d: %s (selected: %v)", index, commit.Subject, isSelected)

	// Choose style and prefix based on selection
	var style lipgloss.Style
	var prefix string
	if isSelected {
		style = selectedStyle
		prefix = "> "
	} else {
		style = normalStyle
		prefix = "  "
	}

	// space for subject minus prefix
	preSubjectContent := 10
	rightMargin := m.Width() / 20
	availableSubjectSpace := m.Width() - (rightMargin + preSubjectContent)

	wrappedSubject := wrapText(commit.Subject, availableSubjectSpace)

	// Format: > hash (7 chars) subject (author, date)
	shortHash := commit.Hash[:7]
	line := fmt.Sprintf("%s%s %s", prefix, shortHash, wrappedSubject[0])
	fmt.Fprintln(w, style.Render(line))

	if len(wrappedSubject) > 1 {
		for i := 1; i < len(wrappedSubject); i++ {
			indentation := "          "
			line := fmt.Sprintf("%s%s", indentation, wrappedSubject[i])
			fmt.Fprintln(w, style.Render(line))
			fmt.Fprintln(w, "\n")
		}
	}

	if isSelected {
		info := fmt.Sprintf("     %s, %s", commit.Author, commit.Date)
		fmt.Fprintln(w, infoStyle.Render(info))
	}

}

// View renders the UI
func (m model) View() string {
	log.Printf("View() called")

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	if m.selected {
		selectedItem := m.list.SelectedItem()
		if selectedItem != nil {
			commit := selectedItem.(Commit)
			return successStyle.Render(fmt.Sprintf("✓ Created fixup commit for: %s\n", commit.Subject))
		}
	}

	// Let the list render itself!
	return m.list.View()
}
