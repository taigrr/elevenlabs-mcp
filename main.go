package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/taigrr/elevenlabs-mcp/internal/ximcp"
)

func main() {
	server, err := ximcp.NewServer()
	if err != nil {
		log.Fatalf("Failed to create ElevenLabs server: %v", err)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Failed to serve MCP server: %v", err)
	}
}
