package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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

// model represents the state of the TUI
type model struct {
	commits  []Commit
	cursor   int
	selected bool
	err      error
	width    int
	height   int
}

// initialModel creates the initial model with commits
func initialModel(commits []Commit) model {
	return model{
		commits: commits,
		cursor:  0,
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		log.Printf("Window resized: %dx%d", m.width, m.height)

	case tea.KeyMsg:
		log.Printf("Key pressed: %s", msg.String())
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.commits)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Select):
			// Create fixup commit for selected commit
			commit := m.commits[m.cursor]
			if err := createFixupCommit(commit.Hash); err != nil {
				m.err = err
				return m, tea.Quit
			}
			m.selected = true
			return m, tea.Quit
		}
	}

	return m, nil
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

// View renders the UI
func (m model) View() string {
	log.Printf("View() called - cursor at: %d", m.cursor)

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	if m.selected {
		commit := m.commits[m.cursor]
		return successStyle.Render(fmt.Sprintf("✓ Created fixup commit for: %s\n", commit.Subject))
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Select a commit to fixup"))
	b.WriteString("\n\n")

	// Commit list
	for i, commit := range m.commits {
		var prefix string
		style := normalStyle
		if i == m.cursor {
			style = selectedStyle
			prefix = "> "
		} else {
			prefix = "  "
		}

		// space for subject minus prefix
		preSubjectContent := 10
		rightMargin := m.width / 20
		availableSubjectSpace := m.width - (rightMargin + preSubjectContent)

		wrappedSubject := wrapText(commit.Subject, availableSubjectSpace)

		// Format: > hash (7 chars) subject (author, date)
		shortHash := commit.Hash[:7]
		line := fmt.Sprintf("%s%s %s", prefix, shortHash, wrappedSubject[0])
		b.WriteString(style.Render(line))
		b.WriteString("\n")

		if len(wrappedSubject) > 1 {
			for i := 1; i < len(wrappedSubject); i++ {
				indentation := "          "
				line := fmt.Sprintf("%s%s", indentation, wrappedSubject[i])
				b.WriteString(style.Render(line))
				b.WriteString("\n")

			}

		}

		// Show author and date for selected commit
		if i == m.cursor {
			info := fmt.Sprintf("    %s, %s", commit.Author, commit.Date)
			b.WriteString(infoStyle.Render(info))
			b.WriteString("\n")
		}
	}

	// Help text
	b.WriteString(helpStyle.Render("↑/↓ or j/k: navigate • enter: create fixup • q: quit"))

	return b.String()
}
