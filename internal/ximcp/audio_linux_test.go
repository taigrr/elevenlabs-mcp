package ximcp

import (
	"context"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestPlaybackRejectsNamedPipe(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "audio.mp3")
	if err := syscall.Mkfifo(filePath, 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateAudioFilePath(filePath); err == nil {
		t.Fatal("expected named pipe to be rejected before playback can block opening it")
	}
	server := &Server{}
	result, _, err := server.play(context.Background(), nil, PlayArgs{FilePath: filePath})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected play tool to report invalid audio file")
	}
	if err := server.PlayAudio(filePath); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("expected regular file validation error, got %v", err)
	}
}
