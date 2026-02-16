package main

import (
	"reflect"
	"testing"
)

// TestWrapText tests the text wrapping function with various inputs
func TestWrapText(t *testing.T) {
	// Table-driven tests - the Go way!
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     []string
	}{
		{
			name:     "text fits within width",
			text:     "short text",
			maxWidth: 20,
			want:     []string{"short text"},
		},
		{
			name:     "exact width match",
			text:     "exactly twenty chars",
			maxWidth: 20,
			want:     []string{"exactly twenty chars"},
		},
		{
			name:     "wrap on word boundary",
			text:     "this is a very long commit message that needs wrapping",
			maxWidth: 20,
			want:     []string{"this is a very long", "commit message that", "needs wrapping"},
		},
		{
			name:     "single long word",
			text:     "supercalifragilisticexpialidocious",
			maxWidth: 15,
			want:     []string{"supercalifragil", "isticexpialidoc", "ious"},
		},
		{
			name:     "multiple spaces",
			text:     "word1    word2    word3",
			maxWidth: 15,
			want:     []string{"word1    word2", "word3"},
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			want:     []string{""},
		},
		{
			name:     "invalid width zero",
			text:     "some text",
			maxWidth: 0,
			want:     []string{"some text"},
		},
		{
			name:     "invalid width negative",
			text:     "some text",
			maxWidth: -5,
			want:     []string{"some text"},
		},
		{
			name:     "wrapping removes leading spaces",
			text:     "first line and second line here",
			maxWidth: 15,
			want:     []string{"first line and", "second line", "here"},
		},
		{
			name:     "text with newlines gets wrapped",
			text:     "this has a\nnewline in it",
			maxWidth: 10,
			want:     []string{"this has", "a\nnewline", "in it"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.maxWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrapText() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWrapTextEdgeCases tests specific edge cases
func TestWrapTextEdgeCases(t *testing.T) {
	t.Run("width of 1", func(t *testing.T) {
		got := wrapText("abc", 1)
		if len(got) != 3 {
			t.Errorf("expected 3 lines, got %d", len(got))
		}
	})

	t.Run("very long text", func(t *testing.T) {
		// Build a long string
		longText := ""
		for i := 0; i < 100; i++ {
			longText += "another "
		}
		got := wrapText(longText, 50)
		// Should produce multiple lines
		if len(got) < 10 {
			t.Errorf("expected many lines for long text, got %d", len(got))
		}
	})
}

// BenchmarkWrapText benchmarks the wrapping function
func BenchmarkWrapText(b *testing.B) {
	text := "This is a reasonably long commit message that will need to be wrapped across multiple lines for display"
	for i := 0; i < b.N; i++ {
		wrapText(text, 40)
	}
}

// BenchmarkWrapTextLong benchmarks with very long text
func BenchmarkWrapTextLong(b *testing.B) {
	text := ""
	for i := 0; i < 1000; i++ {
		text += "word "
	}
	for i := 0; i < b.N; i++ {
		wrapText(text, 80)
	}
}

// TestConstants ensures our layout constants are correct
func TestConstants(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{"hashLength", hashLength, 7},
		{"prefixWidth", prefixWidth, 2},
		{"hashSpacing", hashSpacing, 1},
		{"leftMargin", leftMargin, 10}, // 2 + 7 + 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.want)
			}
		})
	}

	// Test that indentSpaces matches leftMargin
	if len(indentSpaces) != leftMargin {
		t.Errorf("indentSpaces length = %d, want %d (should match leftMargin)",
			len(indentSpaces), leftMargin)
	}

	// Test that leftMargin calculation is correct
	calculatedMargin := prefixWidth + hashLength + hashSpacing
	if leftMargin != calculatedMargin {
		t.Errorf("leftMargin = %d, but calculated as %d",
			leftMargin, calculatedMargin)
	}
}

// TestCommitDelegateType tests the delegate type
func TestCommitDelegateType(t *testing.T) {
	delegate := commitDelegate{}

	if delegate.Height() != 1 {
		t.Errorf("Height() = %d, want 1", delegate.Height())
	}

	if delegate.Spacing() != 0 {
		t.Errorf("Spacing() = %d, want 0", delegate.Spacing())
	}

	// Test that Update returns nil
	if cmd := delegate.Update(nil, nil); cmd != nil {
		t.Errorf("Update() returned non-nil cmd: %v", cmd)
	}
}
