package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"io/fs"

	"github.com/bogem/id3v2"
	"github.com/spf13/cobra"
)

//go:embed bin/*
var binaries embed.FS

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
	// Output directory option
	outputDir string
)

// extractYtDlp extracts the embedded yt-dlp binary to a temporary file
func extractYtDlp(fs fs.FS, dir string) error {
	// Set binary name
	binaryName := "yt-dlp"
	if goos := os.Getenv("GOOS"); goos == "" {
		// If GOOS environment variable is not set, use runtime.GOOS
		if runtime.GOOS == "windows" {
			binaryName = "yt-dlp.exe"
		}
	} else if goos == "windows" {
		binaryName = "yt-dlp.exe"
	}

	// Read the embedded binary
	file, err := fs.Open(filepath.Join("bin", binaryName))
	if err != nil {
		return fmt.Errorf("failed to read embedded yt-dlp: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read embedded yt-dlp: %w", err)
	}

	// Create a temporary file
	tempFile := filepath.Join(dir, binaryName)
	if err := os.WriteFile(tempFile, data, 0755); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	return nil
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
	Use:     "yt2mp3 [URL]",
	Short:   "Download YouTube video and convert to MP3",
	Version: Version,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "yt2mp3")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Extract yt-dlp binary
		if err := extractYtDlp(binaries, tempDir); err != nil {
			return err
		}

		// If output directory is specified, check and create it
		if outputDir != "" {
			// Check access to parent directory
			absOutputDir, err := filepath.Abs(outputDir)
			if err != nil {
				return fmt.Errorf("failed to resolve output directory path: %v", err)
			}
			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %v", err)
			}
			if !strings.HasPrefix(absOutputDir, currentDir) {
				return fmt.Errorf("failed to create output directory: path is outside of current directory")
			}

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
		}

		// Download audio using yt-dlp
		fmt.Println("Downloading audio...")
		ytdlCmd := exec.Command(filepath.Join(tempDir, "yt-dlp"),
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
		if outputDir != "" {
			targetFile = filepath.Join(outputDir, targetFile)
		}

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

		// Fix ID3 tag version to v2.3 for QuickTime compatibility
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

func init() {
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory to specify")
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
