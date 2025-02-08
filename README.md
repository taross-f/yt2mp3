# yt2mp3

[![Test](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml)
[![Release](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taross-f/yt2mp3)](https://goreportcard.com/report/github.com/taross-f/yt2mp3)
[![codecov](https://codecov.io/gh/taross-f/yt2mp3/branch/main/graph/badge.svg)](https://codecov.io/gh/taross-f/yt2mp3)

A CLI tool to download YouTube videos as MP3 files.

## Installation

### macOS (Apple Silicon)

1. Download `yt2mp3-darwin-arm64` from [Releases](https://github.com/taross-f/yt2mp3/releases)

2. Make the binary executable
```bash
chmod +x yt2mp3-darwin-arm64
```

3. Handle security warning (first run only)
```bash
# Remove quarantine attribute to allow execution
xattr -d com.apple.quarantine yt2mp3-darwin-arm64
```

### Windows

1. Download `yt2mp3-windows-amd64.exe` from [Releases](https://github.com/taross-f/yt2mp3/releases)

2. Handle security warning
   - When you first run the executable, Windows SmartScreen might show a warning
   - Click "More info" and then "Run anyway" to proceed
   - This warning appears because the executable is not signed with a certificate

3. (Optional) Add to PATH
   - Move the executable to a permanent location (e.g., `C:\Program Files\yt2mp3\`)
   - Add that location to your PATH environment variable
   - This allows you to run the tool from any directory

## Usage

### macOS
```bash
# Check version
./yt2mp3-darwin-arm64 --version

# Download video as MP3
./yt2mp3-darwin-arm64 "https://www.youtube.com/watch?v=..."
```

### Windows
```cmd
# Check version
yt2mp3-windows-amd64.exe --version

# Download video as MP3
yt2mp3-windows-amd64.exe "https://www.youtube.com/watch?v=..."
```

## Features

- Extract MP3 from YouTube videos
- Automatic ID3 tag setting (title, album, URL)
- QuickTime compatible tag format
- Automatic filename sanitization
- No external dependencies (yt-dlp included)

## License

MIT License 