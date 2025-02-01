# yt2mp3

[![Test](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/test.yml)
[![Release](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml/badge.svg)](https://github.com/taross-f/yt2mp3/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/taross-f/yt2mp3)](https://goreportcard.com/report/github.com/taross-f/yt2mp3)

Simple YouTube to MP3 converter written in Go. This tool downloads YouTube videos and converts them to MP3 format without requiring external dependencies like yt-dlp or ffmpeg.

## Features

- Download YouTube videos and convert to MP3
- No external dependencies required
- Cross-platform support (Windows, macOS, Linux)
- Simple command-line interface

## Installation

### From Source

```bash
go install github.com/taross-f/yt2mp3@latest
```

### Binary Releases

Download the latest binary for your platform from the [releases page](https://github.com/taross-f/yt2mp3/releases).

## Usage

```bash
# Basic usage
yt2mp3 "https://www.youtube.com/watch?v=VIDEO_ID"
```

The MP3 file will be saved in the current directory with the video title as the filename.

## License

MIT License 