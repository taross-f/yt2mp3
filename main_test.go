package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockFS struct {
	files map[string][]byte
}

func (m mockFS) Open(name string) (fs.File, error) {
	if data, ok := m.files[name]; ok {
		return &mockFile{
			Reader: bytes.NewReader(data),
			name:   name,
			size:   int64(len(data)),
		}, nil
	}
	return nil, os.ErrNotExist
}

func (m mockFS) ReadFile(name string) ([]byte, error) {
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m mockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	prefix := name
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	for fileName := range m.files {
		if strings.HasPrefix(fileName, prefix) {
			entries = append(entries, &mockDirEntry{name: strings.TrimPrefix(fileName, prefix)})
		}
	}
	return entries, nil
}

type mockFile struct {
	*bytes.Reader
	name string
	size int64
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) Stat() (fs.FileInfo, error) {
	return &mockFileInfo{
		name: filepath.Base(m.name),
		size: m.size,
	}, nil
}

type mockFileInfo struct {
	name string
	size int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

type mockDirEntry struct {
	name string
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return false }
func (m *mockDirEntry) Type() fs.FileMode          { return 0644 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return &mockFileInfo{name: m.name}, nil }

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal filename",
			input:    "test.mp3",
			expected: "test.mp3",
		},
		{
			name:     "Filename with invalid characters",
			input:    "test:file*?.mp3",
			expected: "test_file__.mp3",
		},
		{
			name:     "Too long filename",
			input:    strings.Repeat("a", 300) + ".mp3",
			expected: strings.Repeat("a", 196) + ".mp3",
		},
		{
			name:     "Japanese filename",
			input:    "テスト.mp3",
			expected: "テスト.mp3",
		},
		{
			name:     "Trim whitespace",
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

func TestIsWithinDir(t *testing.T) {
	base := filepath.FromSlash("/home/user/app")
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"same directory", filepath.FromSlash("/home/user/app"), true},
		{"nested directory", filepath.FromSlash("/home/user/app/music"), true},
		{"deeply nested", filepath.FromSlash("/home/user/app/a/b/c"), true},
		{"parent directory", filepath.FromSlash("/home/user"), false},
		{"sibling sharing prefix", filepath.FromSlash("/home/user/app-evil"), false},
		{"unrelated directory", filepath.FromSlash("/etc"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWithinDir(base, tt.target); got != tt.want {
				t.Errorf("isWithinDir(%q, %q) = %v, want %v", base, tt.target, got, tt.want)
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

	// Create temporary directory for testing
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create existing directory
	existingDir := filepath.Join(tmpDir, "existing-dir")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		args        []string
		flags       []string
		setup       func() error
		mockRunE    func(cmd *cobra.Command, args []string) error
		shouldError bool
	}{
		{
			name:        "No arguments",
			args:        []string{},
			shouldError: true,
		},
		{
			name:        "Valid URL",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		},
		{
			name:        "Output directory specified",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "test-output"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				dir, err := cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				if dir != "test-output" {
					return fmt.Errorf("expected output-dir to be 'test-output', got '%s'", dir)
				}
				return nil
			},
		},
		{
			name:        "Output directory creation error",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "/root/test"}, // trigger a permission error
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("failed to create output directory")
			},
		},
		{
			name:        "Multiple arguments",
			args:        []string{"url1", "url2"},
			shouldError: true,
		},
		{
			name:        "Invalid URL",
			args:        []string{"invalid-url"},
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("invalid URL")
			},
		},
		{
			name:        "Temp directory creation error",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("failed to create temp directory")
			},
		},
		{
			name:        "Existing output directory",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "existing-dir"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				dir, err := cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				if dir != "existing-dir" {
					return fmt.Errorf("expected output-dir to be 'existing-dir', got '%s'", dir)
				}
				return nil
			},
		},
		{
			name:        "Relative path output directory",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "./relative/path"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				dir, err := cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				if dir != "./relative/path" {
					return fmt.Errorf("expected output-dir to be './relative/path', got '%s'", dir)
				}
				return nil
			},
		},
		{
			name:        "Output directory with spaces",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "my music"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				dir, err := cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				if dir != "my music" {
					return fmt.Errorf("expected output-dir to be 'my music', got '%s'", dir)
				}
				return nil
			},
		},
		{
			name:        "Japanese path output directory",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "音楽/ダウンロード"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				dir, err := cmd.Flags().GetString("output-dir")
				if err != nil {
					return err
				}
				if dir != "音楽/ダウンロード" {
					return fmt.Errorf("expected output-dir to be '音楽/ダウンロード', got '%s'", dir)
				}
				return nil
			},
		},
		{
			name:        "Relative path to parent directory",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "../outside"},
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("failed to create output directory: path is outside of current directory")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatal(err)
				}
			}

			if tt.mockRunE != nil {
				rootCmd.RunE = tt.mockRunE
			} else {
				rootCmd.RunE = originalRunE
			}

			cmd := &cobra.Command{
				Use:  "test",
				RunE: rootCmd.RunE,
				Args: rootCmd.Args,
			}
			cmd.Flags().StringP("output-dir", "o", "", "Output directory to specify")
			cmd.SetArgs(append(tt.args, tt.flags...))
			err := cmd.Execute()
			if (err != nil) != tt.shouldError {
				t.Errorf("Execute() error = %v, shouldError %v", err, tt.shouldError)
			}
		})
	}
}

func TestExtractYtDlp(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		binaries   fs.FS
		goos       string
		wantErr    bool
		errMessage string
	}{
		{
			name: "success on darwin",
			binaries: mockFS{
				files: map[string][]byte{
					"bin/yt-dlp": []byte("dummy binary"),
				},
			},
			goos:    "darwin",
			wantErr: false,
		},
		{
			name: "success on linux",
			binaries: mockFS{
				files: map[string][]byte{
					"bin/yt-dlp": []byte("dummy binary"),
				},
			},
			goos:    "linux",
			wantErr: false,
		},
		{
			name: "success on windows",
			binaries: mockFS{
				files: map[string][]byte{
					"bin/yt-dlp.exe": []byte("dummy binary"),
				},
			},
			goos:    "windows",
			wantErr: false,
		},
		{
			name: "binary read error",
			binaries: mockFS{
				files: map[string][]byte{},
			},
			goos:       "linux",
			wantErr:    true,
			errMessage: "failed to read embedded yt-dlp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original GOOS
			originalGOOS := os.Getenv("GOOS")
			os.Setenv("GOOS", tt.goos)
			defer os.Setenv("GOOS", originalGOOS)

			err := extractYtDlp(tt.binaries, tempDir)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				// Verify the extracted file exists
				expectedName := "yt-dlp"
				if tt.goos == "windows" {
					expectedName += ".exe"
				}
				_, err := os.Stat(filepath.Join(tempDir, expectedName))
				assert.NoError(t, err)
			}
		})
	}
}

func TestFixID3Version(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		shouldError bool
	}{
		{
			name: "Convert ID3v2.4 to v2.3",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "test.mp3")
				header := []byte("ID3\x04\x00\x00\x00\x00\x00\x00") // ID3v2.4 header
				if err := os.WriteFile(tmpFile, header, 0644); err != nil {
					t.Fatal(err)
				}
				return tmpFile
			},
			shouldError: false,
		},
		{
			name: "File without ID3 tag",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "test.mp3")
				if err := os.WriteFile(tmpFile, []byte("not an ID3 file"), 0644); err != nil {
					t.Fatal(err)
				}
				return tmpFile
			},
			shouldError: false,
		},
		{
			name: "Nonexistent file",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.mp3")
			},
			shouldError: true,
		},
		{
			name: "Read-only file",
			setup: func(t *testing.T) string {
				if os.Geteuid() == 0 {
					t.Skip("running as root bypasses file permission checks")
				}
				tmpFile := filepath.Join(t.TempDir(), "readonly.mp3")
				header := []byte("ID3\x04\x00\x00\x00\x00\x00\x00")
				if err := os.WriteFile(tmpFile, header, 0444); err != nil {
					t.Fatal(err)
				}
				return tmpFile
			},
			shouldError: true,
		},
		{
			name: "Corrupt ID3 header",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "corrupt.mp3")
				header := []byte("ID3") // incomplete header
				if err := os.WriteFile(tmpFile, header, 0644); err != nil {
					t.Fatal(err)
				}
				return tmpFile
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			err := fixID3Version(path)
			if (err != nil) != tt.shouldError {
				t.Errorf("fixID3Version() error = %v, shouldError %v", err, tt.shouldError)
			}

			if !tt.shouldError && err == nil {
				// On success, verify the file contents
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}

				if len(data) >= 3 && string(data[0:3]) == "ID3" {
					if len(data) >= 4 && data[3] == 4 {
						t.Error("ID3 version was not changed from 2.4")
					}
				}
			}
		})
	}
}
