package ximcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AudioFile struct {
	Name    string
	Summary string
}

func (s *Server) GetAudioHistory() ([]AudioFile, error) {
	files, err := os.ReadDir(AudioDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return []AudioFile{}, nil
		}
		return nil, fmt.Errorf("failed to read %s directory: %w", AudioDirectory, err)
	}

	return s.processAudioFiles(files), nil
}

func (s *Server) processAudioFiles(files []os.DirEntry) []AudioFile {
	var audioFiles []AudioFile

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".mp3") {
			summary := s.getAudioSummary(file.Name())
			audioFiles = append(audioFiles, AudioFile{
				Name:    file.Name(),
				Summary: summary,
			})
		}
	}

	return audioFiles
}

func (s *Server) getAudioSummary(audioFileName string) string {
	textFile := strings.TrimSuffix(audioFileName, ".mp3") + ".txt"
	textPath := filepath.Join(AudioDirectory, textFile)

	content, err := os.ReadFile(textPath)
	if err != nil {
		return "(no text summary available)"
	}

	return s.createSummary(string(content))
}

func (s *Server) createSummary(text string) string {
	text = strings.TrimSpace(text)
	words := strings.Fields(text)

	if len(words) > MaxSummaryWords {
		return strings.Join(words[:MaxSummaryWords], " ") + "..."
	}

	return text
}
