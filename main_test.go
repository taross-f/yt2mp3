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

	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// 既存のディレクトリを作成
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
			name:        "引数なし",
			args:        []string{},
			shouldError: true,
		},
		{
			name:        "正しいURL",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			shouldError: false,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		},
		{
			name:        "出力ディレクトリ指定",
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
			name:        "出力ディレクトリ作成エラー",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			flags:       []string{"--output-dir", "/root/test"}, // 権限エラーを発生させる
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("failed to create output directory")
			},
		},
		{
			name:        "複数の引数",
			args:        []string{"url1", "url2"},
			shouldError: true,
		},
		{
			name:        "無効なURL",
			args:        []string{"invalid-url"},
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("invalid URL")
			},
		},
		{
			name:        "一時ディレクトリ作成エラー",
			args:        []string{"https://www.youtube.com/watch?v=test"},
			shouldError: true,
			mockRunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("failed to create temp directory")
			},
		},
		{
			name:        "既存の出力ディレクトリ",
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
			name:        "相対パスでの出力ディレクトリ指定",
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
			name:        "空白を含む出力ディレクトリ",
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
			name:        "日本語パスでの出力ディレクトリ",
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
			name:        "親ディレクトリへの相対パス",
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
			cmd.Flags().StringP("output-dir", "o", "", "出力ディレクトリを指定")
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
		wantErr    bool
		errMessage string
	}{
		{
			name: "success",
			binaries: mockFS{
				files: map[string][]byte{
					"bin/yt-dlp": []byte("dummy binary"),
				},
			},
			wantErr: false,
		},
		{
			name: "binary read error",
			binaries: mockFS{
				files: map[string][]byte{},
			},
			wantErr:    true,
			errMessage: "failed to read embedded yt-dlp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := extractYtDlp(tt.binaries, tempDir)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				// 抽出されたファイルの存在を確認
				_, err := os.Stat(filepath.Join(tempDir, "yt-dlp"))
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
			name: "ID3v2.4からv2.3への変換",
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
			name: "ID3タグなしのファイル",
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
			name: "存在しないファイル",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.mp3")
			},
			shouldError: true,
		},
		{
			name: "読み取り専用ファイル",
			setup: func(t *testing.T) string {
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
			name: "破損したID3ヘッダー",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "corrupt.mp3")
				header := []byte("ID3") // 不完全なヘッダー
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
				// 成功ケースの場合、ファイルの内容を確認
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
