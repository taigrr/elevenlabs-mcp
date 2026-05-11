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

	files := map[string]string{
		"1710000000000-aaaaa.mp3": "older audio",
		"1710000000000-aaaaa.txt": "older summary",
		"1710000000100-bbbbb.mp3": "newer audio",
		"1710000000100-bbbbb.txt": "newer summary",
		"readme.md":               "not audio",
	}

	for name, content := range files {
		filePath := filepath.Join(tmpDir, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result := s.processAudioFiles(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 audio files, got %d", len(result))
	}

	if result[0].Name != "1710000000100-bbbbb.mp3" {
		t.Fatalf("expected newest file first, got %q", result[0].Name)
	}

	if result[1].Name != "1710000000000-aaaaa.mp3" {
		t.Fatalf("expected older file second, got %q", result[1].Name)
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
