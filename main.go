package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/spf13/cobra"
	"github.com/wader/goutubedl"
)

var rootCmd = &cobra.Command{
	Use:   "yt2mp3 [URL]",
	Short: "Download YouTube video and convert to MP3",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		ctx := context.Background()
		result, err := goutubedl.New(ctx, url, goutubedl.Options{})
		if err != nil {
			return fmt.Errorf("failed to get video info: %v", err)
		}

		filename := sanitizeFilename(result.Info.Title)
		tempFile := filename + ".temp"
		mp3File := filename + ".mp3"
		fmt.Printf("Downloading: %s\n", result.Info.Title)

		// Download the best audio format
		downloadResult, err := result.Download(ctx, "bestaudio")
		if err != nil {
			return fmt.Errorf("failed to download: %v", err)
		}
		defer downloadResult.Close()

		outFile, err := os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %v", err)
		}
		defer func() {
			outFile.Close()
			os.Remove(tempFile)
		}()

		_, err = io.Copy(outFile, downloadResult)
		if err != nil {
			return fmt.Errorf("failed to save file: %v", err)
		}
		outFile.Close()

		// Convert to MP3 using ffmpeg
		fmt.Println("Converting to MP3...")
		ffmpegCmd := exec.Command("ffmpeg", "-i", tempFile, "-codec:a", "libmp3lame", "-q:a", "0", mp3File)
		if output, err := ffmpegCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to convert to MP3: %v\nOutput: %s", err, output)
		}

		// Add ID3 tags
		if err := addID3Tags(mp3File, result.Info); err != nil {
			return fmt.Errorf("failed to add ID3 tags: %v", err)
		}

		fmt.Printf("Successfully downloaded and converted to: %s\n", mp3File)
		return nil
	},
}

func sanitizeFilename(filename string) string {
	// Remove invalid characters
	filename = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, filename)

	// Trim spaces
	filename = strings.TrimSpace(filename)

	// Ensure filename is not too long
	if len(filename) > 200 {
		ext := filepath.Ext(filename)
		filename = filename[:200-len(ext)] + ext
	}

	return filename
}

func addID3Tags(filename string, info goutubedl.Info) error {
	tag, err := id3v2.Open(filename, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open MP3 file for tagging: %v", err)
	}
	defer tag.Close()

	// Set basic tags
	tag.SetTitle(info.Title)
	tag.SetArtist(info.Channel)
	tag.SetAlbum("YouTube")

	// Add YouTube URL as comment
	tag.AddCommentFrame(id3v2.CommentFrame{
		Language:    "eng",
		Description: "YouTube URL",
		Text:        info.WebpageURL,
	})

	// Save the tags
	if err = tag.Save(); err != nil {
		return fmt.Errorf("failed to save ID3 tags: %v", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
