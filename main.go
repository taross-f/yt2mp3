package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "yt2mp3 [YouTube URL]",
		Short: "YouTube動画をMP3に変換するツール",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "エラーが発生しました: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	videoURL := args[0]
	outputDir := "."

	// yt-dlpを使用して直接MP3としてダウンロード
	ytdlpCmd := exec.Command("yt-dlp",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--output", filepath.Join(outputDir, "%(title)s.%(ext)s"),
		videoURL)

	ytdlpCmd.Stdout = os.Stdout
	ytdlpCmd.Stderr = os.Stderr

	if err := ytdlpCmd.Run(); err != nil {
		return fmt.Errorf("動画のダウンロードと変換に失敗しました: %v", err)
	}

	fmt.Println("変換が完了しました")
	return nil
}
