package main

import (
	"testing"
)

// TestCommitFilterValue tests the FilterValue method
func TestCommitFilterValue(t *testing.T) {
	tests := []struct {
		name    string
		commit  Commit
		want    string
	}{
		{
			name: "normal commit",
			commit: Commit{
				Hash:    "abc123",
				Subject: "Fix authentication bug",
				Author:  "John Doe",
				Date:    "2 hours ago",
			},
			want: "Fix authentication bug",
		},
		{
			name: "empty subject",
			commit: Commit{
				Hash:    "def456",
				Subject: "",
				Author:  "Jane Doe",
				Date:    "1 day ago",
			},
			want: "",
		},
		{
			name: "long subject",
			commit: Commit{
				Hash:    "ghi789",
				Subject: "This is a very long commit message that describes many things in great detail",
				Author:  "Bob Smith",
				Date:    "3 days ago",
			},
			want: "This is a very long commit message that describes many things in great detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.commit.FilterValue()
			if got != tt.want {
				t.Errorf("Commit.FilterValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCommitStruct tests the Commit struct fields
func TestCommitStruct(t *testing.T) {
	commit := Commit{
		Hash:    "abcdef1234567890",
		Subject: "Test commit",
		Author:  "Test Author",
		Date:    "now",
	}

	if commit.Hash != "abcdef1234567890" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abcdef1234567890")
	}
	if commit.Subject != "Test commit" {
		t.Errorf("Subject = %q, want %q", commit.Subject, "Test commit")
	}
	if commit.Author != "Test Author" {
		t.Errorf("Author = %q, want %q", commit.Author, "Test Author")
	}
	if commit.Date != "now" {
		t.Errorf("Date = %q, want %q", commit.Date, "now")
	}
}

// TestCommitZeroValue tests the zero value of Commit
func TestCommitZeroValue(t *testing.T) {
	var commit Commit

	if commit.Hash != "" {
		t.Errorf("Zero value Hash should be empty string, got %q", commit.Hash)
	}
	if commit.FilterValue() != "" {
		t.Errorf("Zero value FilterValue should be empty string, got %q", commit.FilterValue())
	}
}

// Note: Testing getCommits() and createFixupCommit() would require mocking
// exec.Command or using integration tests with a real git repository.
// These are examples of functions that are harder to unit test due to
// external dependencies.
