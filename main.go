package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// flags
	debug := flag.Bool("debug", false, "enable debug logs")
	// todo: do I need a description here?
	// todo: adding a comment
	flag.BoolVar(debug, "d", false, "enable debug logs (shorthand)")

	inline := flag.Bool("inline", false, "use inline mode instead of fullscreen")
	// todo: do I need a description here?
	flag.BoolVar(inline, "i", false, "use inline mode (shorthand)")

	flag.Parse()

	cleanup := setupDebug(*debug)
	defer cleanup()

	// Get commits
	commits, err := getCommits()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching commits: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("No commits found on this branch")
		os.Exit(0)
	}

	staged, err := hasStagedChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking working tree: %v\n", err)
		os.Exit(1)
	}

	if !staged {
		fmt.Println("No staged changes. Stage your changes with 'git add' first.")
		os.Exit(0)
	}

	// Initialize the TUI model
	m := initialModel(commits)

	program := getProgram(*inline, m)

	// Run the program
	finalModel, err := program.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Print output after TUI exits
	m, ok := finalModel.(model)
	if ok && m.selected {
		selectedItem := m.list.SelectedItem()
		if selectedItem != nil {
			commit := selectedItem.(Commit)
			fmt.Printf("\n✓ Created fixup commit for %s (%s)\n\n",
				commit.Hash[:hashLength],
				commit.Subject)
			fmt.Println("Run 'git log --oneline' to see your commits.")
			fmt.Println("Run 'git rebase -i --autosquash <base>' to squash fixups.")
		}
	}
}

func setupDebug(debug bool) func() {
	if debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal: ", err)
			os.Exit(1)
		}
		log.Println("debug mode enabled")

		return func() {
			f.Close()
		}

	} else {
		log.SetOutput(io.Discard)
		return func() {}
	}

}

func getProgram(inline bool, m model) *tea.Program {
	if inline {
		log.Println("Starting in inline mode")
		return tea.NewProgram(m)
	} else {
		log.Println("Starting in fullscreen mode")
		return tea.NewProgram(m, tea.WithAltScreen())
	}
}
