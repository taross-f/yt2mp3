# yt2mp3

A command-line tool to convert YouTube videos to MP3 format.

## Requirements

- Go 1.23 or higher
- yt-dlp (must be installed on your system)
- ffmpeg (must be installed on your system)

## Installation

### 1. Install Required Tools

For macOS:
```bash
brew install yt-dlp ffmpeg
```

### 2. Install yt2mp3

```bash
go install github.com/yourusername/yt2mp3@latest
```

## Usage

```bash
yt2mp3 "https://www.youtube.com/watch?v=VIDEO_ID"
```

The converted MP3 file will be saved in the current directory.
The filename is automatically set based on the video title.

## Features

- Simple command-line interface
- High-quality audio extraction
- Automatic metadata preservation
- Fast conversion using yt-dlp and ffmpeg

## Notes

- This tool is intended for personal use only
- Please comply with copyright laws
- Follow YouTube's Terms of Service when using this tool

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 