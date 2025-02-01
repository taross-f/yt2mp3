package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のファイル名",
			input:    "normal file name",
			expected: "normal file name",
		},
		{
			name:     "無効な文字を含むファイル名",
			input:    "file/with:invalid*chars?",
			expected: "file_with_invalid_chars_",
		},
		{
			name:     "長すぎるファイル名",
			input:    string(make([]rune, 250)) + ".mp3",
			expected: string(make([]rune, 196)) + ".mp3",
		},
		{
			name:     "日本語ファイル名",
			input:    "テスト動画 - 音楽.mp3",
			expected: "テスト動画 - 音楽.mp3",
		},
		{
			name:     "空白文字のトリミング",
			input:    "  spaces  ",
			expected: "spaces",
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
	// テスト後にクリーンアップする関数
	cleanup := func() {
		// カレントディレクトリのMP3ファイルを削除
		files, err := filepath.Glob("*.mp3")
		if err != nil {
			t.Logf("クリーンアップ中にエラーが発生: %v", err)
			return
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				t.Logf("ファイル %s の削除中にエラーが発生: %v", f, err)
			}
		}
	}
	// テスト終了後にクリーンアップを実行
	defer cleanup()

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
			args:        []string{"https://www.youtube.com/watch?v=0ywCx6NtlrI"}, // Creative Commons動画
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
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()
			if (err != nil) != tt.shouldError {
				t.Errorf("rootCmd.Execute() error = %v, shouldError %v", err, tt.shouldError)
			}
			// 各テストケース後にもクリーンアップを実行
			cleanup()
		})
	}
}
