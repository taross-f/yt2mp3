# yt2mp3

A command-line tool to convert YouTube videos to MP3 format. Built as a single binary with no external dependencies.

## Features

- Single binary with no external dependencies
- Simple command-line interface
- High-quality audio extraction (320kbps MP3)
- Automatic metadata preservation
- Cross-platform support (Windows, macOS, Linux)

## Requirements

- Go 1.23 or higher (for building from source)

## Installation

### Option 1: Download Binary

Download the latest release from the [releases page](https://github.com/yourusername/yt2mp3/releases).

### Option 2: Build from Source

```bash
go install github.com/yourusername/yt2mp3@latest
```

## Usage

```bash
yt2mp3 "https://www.youtube.com/watch?v=VIDEO_ID"
```

The converted MP3 file will be saved in the current directory.
The filename is automatically set based on the video title.

## Technical Details

- Pure Go implementation
- Uses youtube/v2 for video downloading
- Built-in MP3 encoding using reisen
- High-quality audio output (320kbps)

## Notes

- This tool is intended for personal use only
- Please comply with copyright laws
- Follow YouTube's Terms of Service when using this tool

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 