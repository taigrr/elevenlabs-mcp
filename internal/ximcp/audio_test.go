package ximcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRandomHex(t *testing.T) {
	hex1, err := generateRandomHex(RandomHexLength)
	if err != nil {
		t.Fatalf("generateRandomHex failed: %v", err)
	}

	hex2, err := generateRandomHex(RandomHexLength)
	if err != nil {
		t.Fatalf("generateRandomHex failed: %v", err)
	}

	if len(hex1) != RandomHexLength {
		t.Errorf("expected length %d, got %d", RandomHexLength, len(hex1))
	}

	if len(hex2) != RandomHexLength {
		t.Errorf("expected length %d, got %d", RandomHexLength, len(hex2))
	}

	// Two random hex strings should almost certainly differ
	if hex1 == hex2 {
		t.Error("two consecutive random hex values should differ")
	}

	// Verify hex characters only
	for _, c := range hex1 {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Errorf("non-hex character in output: %c", c)
		}
	}
}

func TestGenerateFilePath(t *testing.T) {
	s := &Server{}

	path, err := s.generateFilePath()
	if err != nil {
		t.Fatalf("generateFilePath failed: %v", err)
	}

	if !strings.HasPrefix(path, AudioDirectory+string(filepath.Separator)) {
		t.Errorf("path should start with %s/, got %q", AudioDirectory, path)
	}

	if !strings.HasSuffix(path, ".mp3") {
		t.Errorf("path should end with .mp3, got %q", path)
	}

	// Two paths should differ (different timestamps or random hex)
	path2, err := s.generateFilePath()
	if err != nil {
		t.Fatalf("generateFilePath failed: %v", err)
	}
	if path == path2 {
		t.Error("two consecutive file paths should differ")
	}
}

func TestEnsureDirectoryExists(t *testing.T) {
	s := &Server{}

	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "a", "b", "c", "file.mp3")

	err := s.ensureDirectoryExists(nested)
	if err != nil {
		t.Fatalf("ensureDirectoryExists failed: %v", err)
	}

	dirPath := filepath.Dir(nested)
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestWriteAudioFile(t *testing.T) {
	s := &Server{}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.mp3")
	data := []byte("fake audio data")

	err := s.writeAudioFile(filePath, data)
	if err != nil {
		t.Fatalf("writeAudioFile failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("file content mismatch: got %q, want %q", content, data)
	}
}

func TestWriteTextFile(t *testing.T) {
	s := &Server{}

	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "test.mp3")
	text := "This is the spoken text"

	err := s.writeTextFile(audioPath, text)
	if err != nil {
		t.Fatalf("writeTextFile failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test.txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read text file: %v", err)
	}

	if string(content) != text {
		t.Errorf("text file content mismatch: got %q, want %q", content, text)
	}
}

func TestReadFileToAudioFileNotFound(t *testing.T) {
	s := &Server{}

	_, err := s.ReadFileToAudio("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestGenerateAudioNoVoice(t *testing.T) {
	s := &Server{
		currentVoice: nil,
	}

	_, err := s.GenerateAudio("test text")
	if err == nil {
		t.Error("expected error when no voice selected")
	}

	if !strings.Contains(err.Error(), "no voice selected") {
		t.Errorf("expected 'no voice selected' error, got: %v", err)
	}
}
