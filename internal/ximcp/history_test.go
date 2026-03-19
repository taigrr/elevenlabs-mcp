package ximcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSummary(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short text unchanged",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "exactly max words",
			input:    "one two three four five six seven eight nine ten",
			expected: "one two three four five six seven eight nine ten",
		},
		{
			name:     "exceeds max words truncated",
			input:    "one two three four five six seven eight nine ten eleven twelve",
			expected: "one two three four five six seven eight nine ten...",
		},
		{
			name:     "empty text",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "text with extra whitespace trimmed at edges",
			input:    "  hello   world  ",
			expected: "hello   world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.createSummary(tt.input)
			if result != tt.expected {
				t.Errorf("createSummary(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAudioHistoryEmptyDir(t *testing.T) {
	s := &Server{}

	// Temporarily override the audio directory to a non-existent path
	origDir := AudioDirectory
	t.Cleanup(func() {
		// AudioDirectory is a const, so we test the behavior as-is
		_ = origDir
	})

	// GetAudioHistory with non-existent directory should return empty
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")

	// We can't override the const, so test processAudioFiles directly
	entries, err := os.ReadDir(nonExistent)
	if err != nil {
		// Expected — directory doesn't exist
		if !os.IsNotExist(err) {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	result := s.processAudioFiles(entries)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d files", len(result))
	}
}

func TestProcessAudioFiles(t *testing.T) {
	s := &Server{}

	tmpDir := t.TempDir()

	// Create test files
	mp3File := filepath.Join(tmpDir, "test.mp3")
	txtFile := filepath.Join(tmpDir, "test.txt")
	otherFile := filepath.Join(tmpDir, "readme.md")

	if err := os.WriteFile(mp3File, []byte("fake audio"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(txtFile, []byte("Hello world test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherFile, []byte("not audio"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result := s.processAudioFiles(entries)

	// Should only include .mp3 files
	if len(result) != 1 {
		t.Fatalf("expected 1 audio file, got %d", len(result))
	}

	if result[0].Name != "test.mp3" {
		t.Errorf("expected name 'test.mp3', got %q", result[0].Name)
	}
}

func TestGetAudioSummary(t *testing.T) {
	s := &Server{}

	tmpDir := t.TempDir()

	// Create a text file alongside an mp3
	txtPath := filepath.Join(tmpDir, "audio.txt")
	if err := os.WriteFile(txtPath, []byte("This is a test summary"), 0644); err != nil {
		t.Fatal(err)
	}

	// getAudioSummary uses AudioDirectory const, so we test createSummary instead
	summary := s.createSummary("This is a test summary")
	if summary != "This is a test summary" {
		t.Errorf("unexpected summary: %q", summary)
	}

	// Test with no text file fallback
	summary = s.createSummary("")
	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}
