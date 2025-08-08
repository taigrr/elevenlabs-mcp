package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/taigrr/elevenlabs/client"
	"github.com/taigrr/elevenlabs/client/types"
)

type ElevenLabsServer struct {
	client       client.Client
	voices       []types.VoiceResponseModel
	currentVoice *types.VoiceResponseModel
	voicesMutex  sync.RWMutex
	playMutex    sync.Mutex
}

func NewElevenLabsServer() (*ElevenLabsServer, error) {
	apiKey := os.Getenv("XI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("XI_API_KEY environment variable is required")
	}

	elevenClient := client.New(apiKey)

	s := &ElevenLabsServer{
		client: elevenClient,
	}

	// Initialize voices and set default
	if err := s.refreshVoices(); err != nil {
		return nil, fmt.Errorf("failed to initialize voices: %w", err)
	}

	// Initialize speaker for audio playback
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	return s, nil
}

func (s *ElevenLabsServer) refreshVoices() error {
	s.voicesMutex.Lock()
	defer s.voicesMutex.Unlock()

	voices, err := s.client.GetVoices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get voices: %w", err)
	}

	s.voices = voices

	// Set default voice if none selected
	if s.currentVoice == nil && len(voices) > 0 {
		s.currentVoice = &voices[0]
	}

	return nil
}

func generateRandomHex(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:length]
}

func (s *ElevenLabsServer) generateAudio(text string) (string, error) {
	if s.currentVoice == nil {
		return "", fmt.Errorf("no voice selected")
	}

	// Generate audio using TTS
	audioData, err := s.client.TTS(context.Background(), text, s.currentVoice.VoiceID, "", types.SynthesisOptions{
		Stability:       0.5,
		SimilarityBoost: 0.5,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate speech: %w", err)
	}

	// Create filename with timestamp and random hex
	timestamp := time.Now().UnixMilli()
	randomHex := generateRandomHex(5)
	filename := fmt.Sprintf("%d-%s.mp3", timestamp, randomHex)
	filePath := filepath.Join(".xi", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write audio file
	if err := os.WriteFile(filePath, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to write audio file: %w", err)
	}

	// Write text file alongside audio
	textFilePath := strings.TrimSuffix(filePath, ".mp3") + ".txt"
	if err := os.WriteFile(textFilePath, []byte(text), 0644); err != nil {
		return "", fmt.Errorf("failed to write text file: %w", err)
	}

	return filePath, nil
}

func (s *ElevenLabsServer) playAudio(filepath string) error {
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

	resampled := beep.Resample(4, format.SampleRate, 44100, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}

func (s *ElevenLabsServer) playAudioAsync(filepath string) {
	go func() {
		if err := s.playAudio(filepath); err != nil {
			log.Printf("Error playing audio: %v", err)
		}
	}()
}

func (s *ElevenLabsServer) setupTools(mcpServer *server.MCPServer) {
	// Say tool
	sayTool := mcp.Tool{
		Name:        "say",
		Description: "Convert text to speech, save as MP3 file, and play the audio",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"text": map[string]any{
					"type":        "string",
					"description": "Text to convert to speech",
				},
			},
			Required: []string{"text"},
		},
	}

	mcpServer.AddTool(sayTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := request.RequireString("text")
		if err != nil {
			return nil, err
		}

		filepath, err := s.generateAudio(text)
		if err != nil {
			return nil, err
		}

		// Play audio asynchronously
		s.playAudioAsync(filepath)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Audio generated, saved to %s, and playing", filepath),
				},
			},
		}, nil
	})

	// Read tool
	readTool := mcp.Tool{
		Name:        "read",
		Description: "Read a text file and convert it to speech, saving as MP3",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "Path to the text file to read and convert to speech",
				},
			},
			Required: []string{"file_path"},
		},
	}

	mcpServer.AddTool(readTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return nil, err
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		text := string(content)
		audioPath, err := s.generateAudio(text)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("File '%s' converted to speech and saved to: %s", filePath, audioPath),
				},
			},
		}, nil
	})

	// Play tool
	playTool := mcp.Tool{
		Name:        "play",
		Description: "Play an audio file",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "Path to the audio file to play",
				},
			},
			Required: []string{"file_path"},
		},
	}

	mcpServer.AddTool(playTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return nil, err
		}

		// Play audio asynchronously
		s.playAudioAsync(filePath)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Playing audio file: %s", filePath),
				},
			},
		}, nil
	})

	// Set voice tool
	setVoiceTool := mcp.Tool{
		Name:        "set_voice",
		Description: "Set the voice to use for text-to-speech generation",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"voice_id": map[string]any{
					"type":        "string",
					"description": "ID of the voice to use",
				},
			},
			Required: []string{"voice_id"},
		},
	}

	mcpServer.AddTool(setVoiceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		voiceID, err := request.RequireString("voice_id")
		if err != nil {
			return nil, err
		}

		s.voicesMutex.Lock()
		defer s.voicesMutex.Unlock()

		// Find the voice
		var selectedVoice *types.VoiceResponseModel
		for i, voice := range s.voices {
			if voice.VoiceID == voiceID {
				selectedVoice = &s.voices[i]
				break
			}
		}

		if selectedVoice == nil {
			return nil, fmt.Errorf("voice with ID '%s' not found", voiceID)
		}

		s.currentVoice = selectedVoice

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Voice set to: %s (%s)", selectedVoice.Name, selectedVoice.VoiceID),
				},
			},
		}, nil
	})

	// Get voices tool
	getVoicesTool := mcp.Tool{
		Name:        "get_voices",
		Description: "Get list of available voices and show the currently selected one",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}

	mcpServer.AddTool(getVoicesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Refresh voices from API
		if err := s.refreshVoices(); err != nil {
			return nil, err
		}

		s.voicesMutex.RLock()
		defer s.voicesMutex.RUnlock()

		var voiceList strings.Builder
		voiceList.WriteString("Available voices:\n")

		for _, voice := range s.voices {
			marker := "  "
			if s.currentVoice != nil && voice.VoiceID == s.currentVoice.VoiceID {
				marker = "* "
			}
			voiceList.WriteString(fmt.Sprintf("%s%s (%s) - %s\n",
				marker, voice.Name, voice.VoiceID, voice.Category))
		}

		if s.currentVoice != nil {
			voiceList.WriteString(fmt.Sprintf("\nCurrently selected: %s (%s)",
				s.currentVoice.Name, s.currentVoice.VoiceID))
		} else {
			voiceList.WriteString("\nNo voice currently selected")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: voiceList.String(),
				},
			},
		}, nil
	})

	// History tool
	historyTool := mcp.Tool{
		Name:        "history",
		Description: "List available audio files with text summaries",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}

	mcpServer.AddTool(historyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Read .xi directory
		files, err := os.ReadDir(".xi")
		if err != nil {
			if os.IsNotExist(err) {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: "No audio files found (directory doesn't exist yet)",
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("failed to read .xi directory: %w", err)
		}

		var audioFiles []string
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".mp3") {
				audioFiles = append(audioFiles, file.Name())
			}
		}

		if len(audioFiles) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "No audio files found",
					},
				},
			}, nil
		}

		var historyList strings.Builder
		historyList.WriteString("Available audio files:\n\n")

		for _, audioFile := range audioFiles {
			// Try to read corresponding text file
			textFile := strings.TrimSuffix(audioFile, ".mp3") + ".txt"
			textPath := filepath.Join(".xi", textFile)

			summary := ""
			if content, err := os.ReadFile(textPath); err == nil {
				text := strings.TrimSpace(string(content))
				words := strings.Fields(text)
				if len(words) > 10 {
					summary = strings.Join(words[:10], " ") + "..."
				} else {
					summary = text
				}
			} else {
				summary = "(no text summary available)"
			}

			historyList.WriteString(fmt.Sprintf("• %s\n  %s\n\n", audioFile, summary))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: historyList.String(),
				},
			},
		}, nil
	})
}

func main() {
	// Create ElevenLabs server
	elevenServer, err := NewElevenLabsServer()
	if err != nil {
		log.Fatalf("Failed to create ElevenLabs server: %v", err)
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"ElevenLabs MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Setup tools
	elevenServer.setupTools(mcpServer)

	// Serve via stdio
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Failed to serve MCP server: %v", err)
	}
}
