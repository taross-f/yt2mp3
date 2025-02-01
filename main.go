package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kkdai/youtube/v2"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "yt2mp3 [YouTube URL]",
		Short: "Convert YouTube videos to MP3 format",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	videoURL := args[0]

	// Initialize YouTube client
	client := youtube.Client{}

	// Get video info
	video, err := client.GetVideo(videoURL)
	if err != nil {
		return fmt.Errorf("failed to get video info: %v", err)
	}

	// Select best audio format
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return fmt.Errorf("no suitable audio format found")
	}

	// Sort formats by audio quality
	var bestFormat *youtube.Format
	bestBitrate := 0
	for i, format := range formats {
		if format.AudioQuality == "AUDIO_QUALITY_MEDIUM" || format.AudioQuality == "AUDIO_QUALITY_HIGH" {
			if format.Bitrate > bestBitrate {
				bestBitrate = format.Bitrate
				bestFormat = &formats[i]
			}
		}
	}

	if bestFormat == nil {
		bestFormat = &formats[0]
	}

	// Download stream
	stream, _, err := client.GetStream(video, bestFormat)
	if err != nil {
		return fmt.Errorf("failed to get stream: %v", err)
	}
	defer stream.Close()

	// Prepare output filename
	outputFile := sanitizeFilename(video.Title) + ".mp3"

	// Create output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer out.Close()

	// Copy audio data directly (since we're getting MP3 format directly)
	_, err = io.Copy(out, stream)
	if err != nil {
		return fmt.Errorf("failed to save audio: %v", err)
	}

	fmt.Printf("Successfully downloaded: %s\n", outputFile)
	return nil
}

func sanitizeFilename(filename string) string {
	// Remove invalid characters
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := filename

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces
	result = strings.TrimSpace(result)

	return result
}
