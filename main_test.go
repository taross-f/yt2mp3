package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のファイル名",
			input:    "test.mp3",
			expected: "test.mp3",
		},
		{
			name:     "無効な文字を含むファイル名",
			input:    "test:file*?.mp3",
			expected: "test_file__.mp3",
		},
		{
			name:     "長すぎるファイル名",
			input:    strings.Repeat("a", 300) + ".mp3",
			expected: strings.Repeat("a", 196) + ".mp3",
		},
		{
			name:     "日本語ファイル名",
			input:    "テスト.mp3",
			expected: "テスト.mp3",
		},
		{
			name:     "空白文字のトリミング",
			input:    " test.mp3 ",
			expected: "test.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRootCmd(t *testing.T) {
	// Save original RunE and restore it after test
	originalRunE := rootCmd.RunE
	defer func() {
		rootCmd.RunE = originalRunE
	}()

	// Mock RunE to avoid actual download
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		return nil
	}

	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "引数なし",
			args:        []string{},
			shouldError: true,
		},
		{
			name:        "正しいURL",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			shouldError: false,
		},
		{
			name:        "複数の引数",
			args:        []string{"url1", "url2"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "test",
				RunE: rootCmd.RunE,
				Args: rootCmd.Args,
			}
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if (err != nil) != tt.shouldError {
				t.Errorf("Execute() error = %v, shouldError %v", err, tt.shouldError)
			}
		})
	}
}

func TestExtractYtDlp(t *testing.T) {
	t.Skip("Skipping test: requires embedded filesystem setup")
}

func TestFixID3Version(t *testing.T) {
	// Create a temporary file with ID3v2.4 tag
	tmpFile := filepath.Join(t.TempDir(), "test.mp3")
	header := []byte("ID3\x04\x00\x00\x00\x00\x00\x00") // ID3v2.4 header
	if err := os.WriteFile(tmpFile, header, 0644); err != nil {
		t.Fatal(err)
	}

	// Test fixID3Version
	if err := fixID3Version(tmpFile); err != nil {
		t.Fatalf("fixID3Version() error = %v", err)
	}

	// Read the file and verify the version was changed to 2.3
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if data[3] != 3 { // Check if version was changed to 2.3
		t.Errorf("ID3 version not changed to 2.3, got %d", data[3])
	}

	// Test with non-existent file
	if err := fixID3Version("nonexistent.mp3"); err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with non-ID3 file
	nonID3File := filepath.Join(t.TempDir(), "non-id3.mp3")
	if err := os.WriteFile(nonID3File, []byte("not an ID3 file"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := fixID3Version(nonID3File); err != nil {
		t.Errorf("fixID3Version() error = %v", err)
	}
}
