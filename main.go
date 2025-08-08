package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/taigrr/elevenlabs-mcp/internal/ximcp"
)

func main() {
	mcpServer := server.NewMCPServer(
		"ElevenLabs MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	elevenServer, err := ximcp.NewServer(mcpServer)
	if err != nil {
		log.Fatalf("Failed to create ElevenLabs server: %v", err)
	}

	elevenServer.SetupTools()

	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Failed to serve MCP server: %v", err)
	}
}
