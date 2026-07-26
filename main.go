package main

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/taigrr/elevenlabs-mcp/internal/ximcp"
)

// version is overridable via -ldflags and otherwise resolved from build info.
var version = "devel"

func init() {
	if version != "devel" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			version = v
		}
	}
}

func main() {
	log.Printf("elevenlabs-mcp %s", version)

	server, err := ximcp.NewServer()
	if err != nil {
		log.Fatalf("Failed to create ElevenLabs server: %v", err)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Failed to serve MCP server: %v", err)
	}
}
