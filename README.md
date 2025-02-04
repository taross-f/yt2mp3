# yt2mp3

[![Test](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml)
[![Release](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taross-f/yt2mp3)](https://goreportcard.com/report/github.com/taross-f/yt2mp3)

A CLI tool to download YouTube videos as MP3 files.

## Installation

### macOS

1. Download the latest release from [Releases](https://github.com/taross-f/yt2mp3/releases)
   - Intel Mac: `yt2mp3-darwin-amd64`
   - Apple Silicon Mac: `yt2mp3-darwin-arm64`

2. Make the binary executable
```bash
chmod +x yt2mp3-darwin-*
```

3. Handle security warning (first run only)
```bash
# Remove quarantine attribute to allow execution
xattr -d com.apple.quarantine yt2mp3-darwin-*
```

### Windows

1. Download `yt2mp3-windows-amd64.exe` from [Releases](https://github.com/taross-f/yt2mp3/releases)
2. Run the downloaded file

### Linux

1. Download `yt2mp3-linux-amd64` from [Releases](https://github.com/taross-f/yt2mp3/releases)
2. Make the binary executable
```bash
chmod +x yt2mp3-linux-amd64
```

## Usage

```bash
# Check version
./yt2mp3 --version

# Download video as MP3
./yt2mp3 "https://www.youtube.com/watch?v=..."
```

## Features

- Extract MP3 from YouTube videos
- Automatic ID3 tag setting (title, album, URL)
- QuickTime compatible tag format
- Automatic filename sanitization
- No external dependencies (yt-dlp included)

## License

MIT License 