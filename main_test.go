package main

import (
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
	// オリジナルのrunEを保存
	originalRunE := rootCmd.RunE
	defer func() {
		// テスト終了後に元に戻す
		rootCmd.RunE = originalRunE
	}()

	// テスト用のモックコマンド
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		// 実際のダウンロードは行わず、成功したものとする
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
		})
	}
}
