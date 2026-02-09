package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit represents a git commit
type Commit struct {
	Hash    string
	Subject string
	Author  string
	Date    string
}

func (c Commit) FilterValue() string {
	return c.Subject
}

// getCommits retrieves all non-fixup commits on the current branch since main
func getCommits() ([]Commit, error) {
	// Find the merge base with main
	mergeBaseCmd := exec.Command("git", "merge-base", "HEAD", "main")
	mergeBase, err := mergeBaseCmd.Output()
	if err != nil {
		// If main doesn't exist, try master
		mergeBaseCmd = exec.Command("git", "merge-base", "HEAD", "master")
		mergeBase, err = mergeBaseCmd.Output()
		if err != nil {
			return nil, fmt.Errorf("could not find merge base with main/master: %w", err)
		}
	}

	base := strings.TrimSpace(string(mergeBase))

	// Get commits since the merge base, excluding fixup commits
	// Format: hash|subject|author name|date
	cmd := exec.Command("git", "log",
		"--format=%H|%s|%an|%ar",
		"--no-merges",
		fmt.Sprintf("%s..HEAD", base),
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		// Skip fixup and squash commits
		subject := parts[1]
		if strings.HasPrefix(subject, "fixup!") || strings.HasPrefix(subject, "squash!") {
			continue
		}

		commits = append(commits, Commit{
			Hash:    parts[0],
			Subject: parts[1],
			Author:  parts[2],
			Date:    parts[3],
		})
	}

	return commits, nil
}

// createFixupCommit creates a fixup commit for the given commit hash
func createFixupCommit(hash string) error {
	cmd := exec.Command("git", "commit", "--fixup", hash)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create fixup commit: %w\nOutput: %s", err, output)
	}
	return nil
}
