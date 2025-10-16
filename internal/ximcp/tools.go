package ximcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/taigrr/elevenlabs/client/types"
)

type SayArgs struct {
	Text string `json:"text" jsonschema:"Text to convert to speech"`
}

type ReadArgs struct {
	FilePath string `json:"file_path" jsonschema:"Path to the text file to read and convert to speech"`
}

type PlayArgs struct {
	FilePath string `json:"file_path" jsonschema:"Path to the audio file to play"`
}

type SetVoiceArgs struct {
	VoiceID string `json:"voice_id" jsonschema:"ID of the voice to use"`
}

func (s *Server) setupTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "say",
		Description: "Convert text to speech, save as MP3 file, and play the audio",
	}, s.say)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "read",
		Description: "Read a text file and convert it to speech, saving as MP3",
	}, s.read)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "play",
		Description: "Play an audio file",
	}, s.play)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "set_voice",
		Description: "Set the voice to use for text-to-speech generation",
	}, s.setVoice)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_voices",
		Description: "Get list of available voices and show the currently selected one",
	}, s.getVoices)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history",
		Description: "List available audio files with text summaries",
	}, s.history)
}

func (s *Server) say(ctx context.Context, req *mcp.CallToolRequest, args SayArgs) (*mcp.CallToolResult, any, error) {
	filepath, err := s.GenerateAudio(args.Text)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	s.PlayAudioAsync(filepath)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Audio generated, saved to %s, and playing", filepath)},
		},
	}, nil, nil
}

func (s *Server) read(ctx context.Context, req *mcp.CallToolRequest, args ReadArgs) (*mcp.CallToolResult, any, error) {
	audioPath, err := s.ReadFileToAudio(args.FilePath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("File '%s' converted to speech and saved to: %s", args.FilePath, audioPath)},
		},
	}, nil, nil
}

func (s *Server) play(ctx context.Context, req *mcp.CallToolRequest, args PlayArgs) (*mcp.CallToolResult, any, error) {
	s.PlayAudioAsync(args.FilePath)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Playing audio file: %s", args.FilePath)},
		},
	}, nil, nil
}

func (s *Server) setVoice(ctx context.Context, req *mcp.CallToolRequest, args SetVoiceArgs) (*mcp.CallToolResult, any, error) {
	selectedVoice, err := s.SetVoice(args.VoiceID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Voice set to: %s (%s)", selectedVoice.Name, selectedVoice.VoiceID)},
		},
	}, nil, nil
}

func (s *Server) getVoices(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	voices, currentVoice, err := s.GetVoices()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	voiceList := s.formatVoiceList(voices, currentVoice)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: voiceList},
		},
	}, nil, nil
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

func (s *Server) history(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	audioFiles, err := s.GetAudioHistory()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	if len(audioFiles) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No audio files found"},
			},
		}, nil, nil
	}

	historyList := s.formatHistoryList(audioFiles)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: historyList},
		},
	}, nil, nil
}

func (s *Server) formatHistoryList(audioFiles []AudioFile) string {
	var historyList strings.Builder
	historyList.WriteString("Available audio files:\n\n")

	for _, audioFile := range audioFiles {
		historyList.WriteString(fmt.Sprintf("• %s\n  %s\n\n", audioFile.Name, audioFile.Summary))
	}

	return historyList.String()
}
