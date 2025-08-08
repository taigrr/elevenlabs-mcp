package ximcp

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/taigrr/elevenlabs/client/types"
)

func generateRandomHex(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:length]
}

func (s *Server) GenerateAudio(text string) (string, error) {
	if s.currentVoice == nil {
		return "", fmt.Errorf("no voice selected")
	}

	audioData, err := s.generateTTSAudio(text)
	if err != nil {
		return "", err
	}

	filePath, err := s.saveAudioFiles(text, audioData)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *Server) generateTTSAudio(text string) ([]byte, error) {
	return s.client.TTS(context.Background(), text, s.currentVoice.VoiceID, "", types.SynthesisOptions{
		Stability:       DefaultStability,
		SimilarityBoost: DefaultSimilarityBoost,
	})
}

func (s *Server) saveAudioFiles(text string, audioData []byte) (string, error) {
	filePath := s.generateFilePath()

	if err := s.ensureDirectoryExists(filePath); err != nil {
		return "", err
	}

	if err := s.writeAudioFile(filePath, audioData); err != nil {
		return "", err
	}

	if err := s.writeTextFile(filePath, text); err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *Server) generateFilePath() string {
	timestamp := time.Now().UnixMilli()
	randomHex := generateRandomHex(RandomHexLength)
	filename := fmt.Sprintf("%d-%s.mp3", timestamp, randomHex)
	return filepath.Join(AudioDirectory, filename)
}

func (s *Server) ensureDirectoryExists(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (s *Server) writeAudioFile(filePath string, audioData []byte) error {
	if err := os.WriteFile(filePath, audioData, 0644); err != nil {
		return fmt.Errorf("failed to write audio file: %w", err)
	}
	return nil
}

func (s *Server) writeTextFile(filePath, text string) error {
	textFilePath := strings.TrimSuffix(filePath, ".mp3") + ".txt"
	if err := os.WriteFile(textFilePath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write text file: %w", err)
	}
	return nil
}

func (s *Server) PlayAudio(filepath string) error {
	s.playMutex.Lock()
	defer s.playMutex.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode mp3: %w", err)
	}
	defer streamer.Close()

	return s.playStreamer(streamer, format)
}

func (s *Server) playStreamer(streamer beep.StreamSeekCloser, format beep.Format) error {
	resampled := beep.Resample(4, format.SampleRate, AudioSampleRate, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}

func (s *Server) PlayAudioAsync(filepath string) {
	go func() {
		if err := s.PlayAudio(filepath); err != nil {
			log.Printf("Error playing audio: %v", err)
		}
	}()
}

func (s *Server) ReadFileToAudio(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	text := string(content)
	return s.GenerateAudio(text)
}
