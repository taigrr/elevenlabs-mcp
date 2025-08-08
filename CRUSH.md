# ElevenLabs MCP Server

## Build/Test Commands
- Build: `go build -o elevenlabs-mcp`
- Run: `./elevenlabs-mcp` (requires XI_API_KEY env var)
- Test: `go test ./...`
- Lint: `golangci-lint run` (if available) or `go vet ./...`
- Format: `gofmt -w .` or `goimports -w .`
- Dependencies: `go mod tidy && go mod download`

## Environment Setup
- Required: `export XI_API_KEY=your_api_key_here`
- Audio files saved to: `.xi/<millis>-<hex5>.mp3`

## Code Style
- Use `goimports` for formatting
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- No single-letter variables except loop counters
- Use meaningful error messages with context
- Prefer explicit error handling over panics
- Use sync.RWMutex for concurrent access to shared data
- Constants for magic strings/numbers, defined at package level

## MCP Tools Provided
- `say`: Convert text to speech, save as MP3
- `read`: Read text file and convert to speech  
- `play`: Play audio file using beep library
- `set_voice`: Change TTS voice (memory only)
- `get_voices`: List available voices, show current selection
- `history`: List available audio files with text summaries

## Dependencies
- `github.com/mark3labs/mcp-go` - MCP server framework
- `github.com/taigrr/elevenlabs` - ElevenLabs API client
- `github.com/gopxl/beep` - Audio playback
- `github.com/google/uuid` - UUID generation