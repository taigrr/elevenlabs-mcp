package ximcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/taigrr/elevenlabs/client/types"
)

func (s *Server) SetupTools() {
	s.mcpServer.AddTool(s.say())
	s.mcpServer.AddTool(s.read())
	s.mcpServer.AddTool(s.play())
	s.mcpServer.AddTool(s.setVoice())
	s.mcpServer.AddTool(s.getVoices())
	s.mcpServer.AddTool(s.history())
}

func (s *Server) say() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := request.RequireString("text")
		if err != nil {
			return nil, err
		}

		filepath, err := s.GenerateAudio(text)
		if err != nil {
			return nil, err
		}

		s.PlayAudioAsync(filepath)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Audio generated, saved to %s, and playing", filepath),
				},
			},
		}, nil
	}
	return tool, handler
}

func (s *Server) read() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return nil, err
		}

		audioPath, err := s.ReadFileToAudio(filePath)
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
	}
	return tool, handler
}

func (s *Server) play() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return nil, err
		}

		s.PlayAudioAsync(filePath)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Playing audio file: %s", filePath),
				},
			},
		}, nil
	}
	return tool, handler
}

func (s *Server) setVoice() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
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

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		voiceID, err := request.RequireString("voice_id")
		if err != nil {
			return nil, err
		}

		selectedVoice, err := s.SetVoice(voiceID)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Voice set to: %s (%s)", selectedVoice.Name, selectedVoice.VoiceID),
				},
			},
		}, nil
	}
	return tool, handler
}

func (s *Server) getVoices() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
		Name:        "get_voices",
		Description: "Get list of available voices and show the currently selected one",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		voices, currentVoice, err := s.GetVoices()
		if err != nil {
			return nil, err
		}

		voiceList := s.formatVoiceList(voices, currentVoice)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: voiceList,
				},
			},
		}, nil
	}
	return tool, handler
}

func (s *Server) formatVoiceList(voices []types.VoiceResponseModel, currentVoice *types.VoiceResponseModel) string {
	var voiceList strings.Builder
	voiceList.WriteString("Available voices:\n")

	for _, voice := range voices {
		marker := "  "
		if currentVoice != nil && voice.VoiceID == currentVoice.VoiceID {
			marker = "* "
		}
		voiceList.WriteString(fmt.Sprintf("%s%s (%s) - %s\n",
			marker, voice.Name, voice.VoiceID, voice.Category))
	}

	if currentVoice != nil {
		voiceList.WriteString(fmt.Sprintf("\nCurrently selected: %s (%s)",
			currentVoice.Name, currentVoice.VoiceID))
	} else {
		voiceList.WriteString("\nNo voice currently selected")
	}

	return voiceList.String()
}

func (s *Server) history() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.Tool{
		Name:        "history",
		Description: "List available audio files with text summaries",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		audioFiles, err := s.GetAudioHistory()
		if err != nil {
			return nil, err
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

		historyList := s.formatHistoryList(audioFiles)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: historyList,
				},
			},
		}, nil
	}
	return tool, handler
}

func (s *Server) formatHistoryList(audioFiles []AudioFile) string {
	var historyList strings.Builder
	historyList.WriteString("Available audio files:\n\n")

	for _, audioFile := range audioFiles {
		historyList.WriteString(fmt.Sprintf("• %s\n  %s\n\n", audioFile.Name, audioFile.Summary))
	}

	return historyList.String()
}
