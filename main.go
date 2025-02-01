package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

		filename := sanitizeFilename(result.Info.Title) + ".mp3"
		fmt.Printf("Downloading: %s\n", result.Info.Title)

		// Download the best audio format
		downloadResult, err := result.Download(ctx, "bestaudio")
		if err != nil {
			return fmt.Errorf("failed to download: %v", err)
		}
		defer downloadResult.Close()

		outFile, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, downloadResult)
		if err != nil {
			return fmt.Errorf("failed to save file: %v", err)
		}

		fmt.Printf("Successfully downloaded and converted to: %s\n", filename)
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
