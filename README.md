# ElevenLabs MCP Server

An MCP (Model Context Protocol) server that provides text-to-speech capabilities using ElevenLabs API.

## Features

- **say**: Convert text to speech and save as MP3 file
- **read**: Read a text file and convert it to speech
- **play**: Play audio files using the beep library
- **set_voice**: Change the voice used for generation (memory only)
- **get_voices**: List available voices and show currently selected one
- **history**: List available audio files with text summaries

## Setup

1. Get your ElevenLabs API key from [ElevenLabs](https://elevenlabs.io)
2. Set the environment variable:
   ```bash
   export XI_API_KEY=your_api_key_here
   ```

## Build

```bash
go build -o elevenlabs-mcp
```

## Usage

The server runs via stdio and communicates using the MCP protocol:

```bash
export XI_API_KEY=your_api_key_here
./elevenlabs-mcp
```

## Tools

### say
Convert text to speech and save as MP3.
- **text** (string, required): Text to convert to speech

### read  
Read a text file and convert it to speech.
- **file_path** (string, required): Path to the text file

### play
Play an audio file.
- **file_path** (string, required): Path to the audio file

### set_voice
Change the voice used for generation.
- **voice_id** (string, required): ID of the voice to use

### get_voices
List all available voices and show the currently selected one.
- No parameters required

### history
List available audio files with text summaries.
- No parameters required

## Audio Files

Generated audio files are saved to `.xi/<timestamp>-<hex5>.mp3` with corresponding `.txt` files containing the original text. Files can be replayed using the `play` tool.

## Dependencies

- ElevenLabs API for text-to-speech generation
- Beep library for audio playback
- Mark3labs MCP-Go for MCP server functionality