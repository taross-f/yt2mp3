package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/spf13/cobra"
)

//go:embed bin/*
var binaries embed.FS

// extractYtDlp extracts the embedded yt-dlp binary to a temporary file
func extractYtDlp() (string, error) {
	// Determine the binary name based on the OS
	binaryName := "yt-dlp"
	if runtime.GOOS == "windows" {
		binaryName = "yt-dlp.exe"
	}

	// Read the embedded binary
	data, err := binaries.ReadFile(filepath.Join("bin", binaryName))
	if err != nil {
		return "", fmt.Errorf("failed to read embedded yt-dlp: %v", err)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "yt-dlp-*"+filepath.Ext(binaryName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Make the file executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return "", fmt.Errorf("failed to make yt-dlp executable: %v", err)
		}
	}

	// Write the binary data
	if _, err := tmpFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write yt-dlp: %v", err)
	}

	return tmpFile.Name(), nil
}

// sanitizeFilename removes or replaces invalid characters from the filename
// and ensures the filename length is within acceptable limits
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscore
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces from start and end
	result = strings.TrimSpace(result)

	// Ensure filename is not too long (max 200 chars including extension)
	if len(result) > 200 {
		ext := filepath.Ext(result)
		result = result[:200-len(ext)] + ext
	}

	return result
}

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

		// Extract yt-dlp binary
		ytdlpPath, err := extractYtDlp()
		if err != nil {
			return err
		}
		defer os.Remove(ytdlpPath)

		// Download audio using yt-dlp
		fmt.Println("Downloading audio...")
		ytdlCmd := exec.Command(ytdlpPath,
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

		// Use the downloaded file
		downloadedFile := filepath.Join(tempDir, files[0].Name())
		targetFile := sanitizeFilename(files[0].Name())

		// Open file for ID3 tag editing
		tag, err := id3v2.Open(downloadedFile, id3v2.Options{Parse: true})
		if err != nil {
			return fmt.Errorf("failed to open MP3 file for tagging: %v", err)
		}

		// Set basic tags
		tag.SetTitle(strings.TrimSuffix(targetFile, filepath.Ext(targetFile)))
		tag.SetAlbum("YouTube")
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

		// QuickTime互換のため、ID3タグバージョンをv2.3に固定する
		if err = fixID3Version(downloadedFile); err != nil {
			return fmt.Errorf("failed to fix ID3 version: %v", err)
		}

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

// fixID3Version opens the MP3 file and changes the tag version from ID3v2.4 to ID3v2.3 if needed.
func fixID3Version(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 10)
	n, err := f.Read(header)
	if err != nil || n != 10 {
		return fmt.Errorf("failed to read header")
	}

	if string(header[0:3]) != "ID3" {
		// No ID3 tag present, nothing to fix
		return nil
	}

	// If version is 2.4 (0x04), change it to 2.3 (0x03)
	if header[3] == 4 {
		header[3] = 3
		_, err = f.Seek(0, 0)
		if err != nil {
			return err
		}
		_, err = f.Write(header)
		if err != nil {
			return err
		}
	}

	return nil
}
