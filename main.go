package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yt2mp3 [URL]",
	Short: "Download YouTube video and convert to MP3",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "yt2mp3")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Download audio using yt-dlp
		fmt.Println("Downloading audio...")
		ytdlCmd := exec.Command("yt-dlp",
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--output", filepath.Join(tempDir, "%(title)s.%(ext)s"),
			url,
		)
		if output, err := ytdlCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to download audio: %v\nOutput: %s", err, output)
		}

		// Find the downloaded file
		files, err := os.ReadDir(tempDir)
		if err != nil {
			return fmt.Errorf("failed to read temp directory: %v", err)
		}
		if len(files) == 0 {
			return fmt.Errorf("no files downloaded")
		}

		// Move the file to current directory
		downloadedFile := filepath.Join(tempDir, files[0].Name())
		targetFile := files[0].Name()

		// Read file for ID3 tags
		tag, err := id3v2.Open(downloadedFile, id3v2.Options{Parse: true})
		if err != nil {
			return fmt.Errorf("failed to open MP3 file for tagging: %v", err)
		}

		// Set basic tags
		tag.SetTitle(strings.TrimSuffix(files[0].Name(), filepath.Ext(files[0].Name())))
		tag.SetAlbum("YouTube")

		// Add YouTube URL as comment
		tag.AddCommentFrame(id3v2.CommentFrame{
			Language:    "eng",
			Description: "YouTube URL",
			Text:        url,
		})

		// Save the tags
		if err = tag.Save(); err != nil {
			return fmt.Errorf("failed to save ID3 tags: %v", err)
		}
		tag.Close()

		// Move file to current directory
		if err := os.Rename(downloadedFile, targetFile); err != nil {
			return fmt.Errorf("failed to move file: %v", err)
		}

		fmt.Printf("Successfully downloaded and converted to: %s\n", targetFile)
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
