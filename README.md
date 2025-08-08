# elevenlabs-mcp

[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://godoc.org/github.com/taigrr/elevenlabs-mcp?status.svg)](https://godoc.org/github.com/taigrr/elevenlabs-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/taigrr/elevenlabs-mcp)](https://goreportcard.com/report/github.com/taigrr/elevenlabs-mcp)

An MCP (Model Context Protocol) server that bridges AI assistants with ElevenLabs text-to-speech capabilities.

This server is not affiliated with, nor associated with ElevenLabs in any way.

## Purpose

This MCP server enables AI assistants to generate speech from text using the ElevenLabs API.
It provides a seamless interface for converting text to high-quality speech, managing voice selection, and handling audio playback within MCP-compatible environments.

As a prerequisite, you must already have an account with [ElevenLabs](https://elevenlabs.io).
After creating your account, you can get your API key [from here](https://help.elevenlabs.io/hc/en-us/articles/14599447207697-How-to-authorize-yourself-using-your-xi-api-key-).
Note, your API key will read access to your voices and to Text-to-Speech generation as a minimum to function properly.

## Installation

```bash
go install .
```

## Configuration

Set your ElevenLabs API key:

```bash
export XI_API_KEY=your_api_key_here
```

## Usage

The server communicates via stdio using the MCP protocol.

You'll need a compatible MCP client to interact with this server.

Generated audio files are automatically saved to `.xi/<timestamp>-<hex5>.mp3` with corresponding `.txt` files containing the original text for reference.

## MCP Tools

The server provides the following tools to MCP clients:

- **say** - Convert text to speech and save as MP3
- **read** - Read a text file and convert it to speech
- **play** - Play audio files using system audio
- **set_voice** - Change the voice used for generation
- **get_voices** - List available voices and show current selection
- **history** - List previously generated audio files with (truncated) text summaries

## Dependencies

- [ElevenLabs API](https://elevenlabs.io) for text-to-speech generation
- [Beep library](https://github.com/gopxl/beep) for audio playback
- [MCP-Go](https://github.com/mark3labs/mcp-go) for MCP server functionality

## License

This project is licensed under the 0BSD License, written by [Rob Landley](https://github.com/landley).
As such, you may use this library without restriction or attribution, but please don't pass it off as your own.
Attribution, though not required, is appreciated.

By contributing, you agree all code submitted also falls under the License.

