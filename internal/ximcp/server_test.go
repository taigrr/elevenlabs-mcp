package ximcp

import (
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/client/types"
)

func TestFindVoiceByID(t *testing.T) {
	s := &Server{
		voices: []types.VoiceResponseModel{
			{VoiceID: "abc123", Name: "Alice"},
			{VoiceID: "def456", Name: "Bob"},
			{VoiceID: "ghi789", Name: "Charlie"},
		},
	}

	tests := []struct {
		name     string
		voiceID  string
		expected string
		found    bool
	}{
		{"find first voice", "abc123", "Alice", true},
		{"find middle voice", "def456", "Bob", true},
		{"find last voice", "ghi789", "Charlie", true},
		{"voice not found", "zzz999", "", false},
		{"empty id", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voice := s.findVoiceByID(tt.voiceID)
			if tt.found {
				if voice == nil {
					t.Fatal("expected voice, got nil")
				}
				if voice.Name != tt.expected {
					t.Errorf("expected name %q, got %q", tt.expected, voice.Name)
				}
			} else {
				if voice != nil {
					t.Errorf("expected nil, got voice %q", voice.Name)
				}
			}
		})
	}
}

func TestSetDefaultVoiceIfNeeded(t *testing.T) {
	t.Run("sets first voice when no current", func(t *testing.T) {
		s := &Server{
			voices: []types.VoiceResponseModel{
				{VoiceID: "abc123", Name: "Alice"},
				{VoiceID: "def456", Name: "Bob"},
			},
			currentVoice: nil,
		}

		s.setDefaultVoiceIfNeeded()

		if s.currentVoice == nil {
			t.Fatal("expected current voice to be set")
		}
		if s.currentVoice.Name != "Alice" {
			t.Errorf("expected first voice 'Alice', got %q", s.currentVoice.Name)
		}
	})

	t.Run("does not override existing voice", func(t *testing.T) {
		bob := types.VoiceResponseModel{VoiceID: "def456", Name: "Bob"}
		s := &Server{
			voices: []types.VoiceResponseModel{
				{VoiceID: "abc123", Name: "Alice"},
				bob,
			},
			currentVoice: &bob,
		}

		s.setDefaultVoiceIfNeeded()

		if s.currentVoice.Name != "Bob" {
			t.Errorf("expected voice to remain 'Bob', got %q", s.currentVoice.Name)
		}
	})

	t.Run("no voices available", func(t *testing.T) {
		s := &Server{
			voices:       []types.VoiceResponseModel{},
			currentVoice: nil,
		}

		s.setDefaultVoiceIfNeeded()

		if s.currentVoice != nil {
			t.Error("expected no voice when list is empty")
		}
	})
}

func TestSetVoice(t *testing.T) {
	s := &Server{
		voices: []types.VoiceResponseModel{
			{VoiceID: "abc123", Name: "Alice"},
			{VoiceID: "def456", Name: "Bob"},
		},
	}

	t.Run("set valid voice", func(t *testing.T) {
		voice, err := s.SetVoice("def456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if voice.Name != "Bob" {
			t.Errorf("expected 'Bob', got %q", voice.Name)
		}
		if s.currentVoice.VoiceID != "def456" {
			t.Error("currentVoice not updated")
		}
	})

	t.Run("set invalid voice", func(t *testing.T) {
		_, err := s.SetVoice("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent voice")
		}
	})
}

func TestFormatVoiceList(t *testing.T) {
	current := types.VoiceResponseModel{VoiceID: "abc123", Name: "Alice", Category: "premade"}
	s := &Server{}

	voices := []types.VoiceResponseModel{
		current,
		{VoiceID: "def456", Name: "Bob", Category: "cloned"},
	}

	result := s.formatVoiceList(voices, &current)

	if result == "" {
		t.Error("expected non-empty result")
	}

	// Current voice should be marked with asterisk
	if !contains(result, "* Alice") {
		t.Error("expected current voice to be marked with asterisk")
	}

	// Other voice should not be marked
	if contains(result, "* Bob") {
		t.Error("non-current voice should not have asterisk marker")
	}

	// Both voices should appear
	if !contains(result, "Alice") || !contains(result, "Bob") {
		t.Error("both voices should appear in output")
	}

	if !contains(result, "Currently selected: Alice") {
		t.Error("expected currently selected line")
	}
}

func TestFormatVoiceListNoSelection(t *testing.T) {
	s := &Server{}

	voices := []types.VoiceResponseModel{
		{VoiceID: "abc123", Name: "Alice", Category: "premade"},
	}

	result := s.formatVoiceList(voices, nil)

	if !contains(result, "No voice currently selected") {
		t.Error("expected 'No voice currently selected' when nil")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
