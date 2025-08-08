package ximcp

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/mark3labs/mcp-go/server"
	"github.com/taigrr/elevenlabs/client"
	"github.com/taigrr/elevenlabs/client/types"
)

const (
	DefaultStability       = 0.5
	DefaultSimilarityBoost = 0.5
	AudioDirectory         = ".xi"
	AudioSampleRate        = 44100
	RandomHexLength        = 5
	MaxSummaryWords        = 10
)

type Server struct {
	mcpServer    *server.MCPServer
	client       client.Client
	voices       []types.VoiceResponseModel
	currentVoice *types.VoiceResponseModel
	voicesMutex  sync.RWMutex
	playMutex    sync.Mutex
}

func NewServer(mcpServer *server.MCPServer) (*Server, error) {
	apiKey := os.Getenv("XI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("XI_API_KEY environment variable is required")
	}

	elevenClient := client.New(apiKey)

	s := &Server{
		client:    elevenClient,
		mcpServer: mcpServer,
	}

	if err := s.initializeVoices(); err != nil {
		return nil, fmt.Errorf("failed to initialize voices: %w", err)
	}

	if err := s.initializeSpeaker(); err != nil {
		return nil, fmt.Errorf("failed to initialize speaker: %w", err)
	}

	return s, nil
}

func (s *Server) initializeVoices() error {
	if err := s.refreshVoices(); err != nil {
		return err
	}
	return nil
}

func (s *Server) initializeSpeaker() error {
	sr := beep.SampleRate(AudioSampleRate)
	speaker.Init(sr, sr.N(time.Second/10))
	return nil
}

func (s *Server) refreshVoices() error {
	s.voicesMutex.Lock()
	defer s.voicesMutex.Unlock()

	voices, err := s.client.GetVoices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get voices: %w", err)
	}

	s.voices = voices
	s.setDefaultVoiceIfNeeded()
	return nil
}

func (s *Server) setDefaultVoiceIfNeeded() {
	if s.currentVoice == nil && len(s.voices) > 0 {
		s.currentVoice = &s.voices[0]
	}
}

func (s *Server) GetVoices() ([]types.VoiceResponseModel, *types.VoiceResponseModel, error) {
	if err := s.refreshVoices(); err != nil {
		return nil, nil, err
	}

	s.voicesMutex.RLock()
	defer s.voicesMutex.RUnlock()

	return s.voices, s.currentVoice, nil
}

func (s *Server) SetVoice(voiceID string) (*types.VoiceResponseModel, error) {
	s.voicesMutex.Lock()
	defer s.voicesMutex.Unlock()

	selectedVoice := s.findVoiceByID(voiceID)
	if selectedVoice == nil {
		return nil, fmt.Errorf("voice with ID '%s' not found", voiceID)
	}

	s.currentVoice = selectedVoice
	return selectedVoice, nil
}

func (s *Server) findVoiceByID(voiceID string) *types.VoiceResponseModel {
	for i, voice := range s.voices {
		if voice.VoiceID == voiceID {
			return &s.voices[i]
		}
	}
	return nil
}
