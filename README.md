# yt2mp3

YouTubeの動画をMP3形式に変換するコマンドラインツールです。

## 必要条件

- Go 1.23以上
- yt-dlp（システムにインストールされている必要があります）
- ffmpeg（システムにインストールされている必要があります）

## インストール

### 1. 必要なツールのインストール

macOSの場合：
```bash
brew install yt-dlp ffmpeg
```

### 2. yt2mp3のインストール

```bash
go install github.com/yourusername/yt2mp3@latest
```

## 使い方

```bash
yt2mp3 "https://www.youtube.com/watch?v=VIDEO_ID"
```

変換されたMP3ファイルは、現在のディレクトリに保存されます。
ファイル名は動画のタイトルに基づいて自動的に設定されます。

## 注意事項

- このツールは個人使用目的のみを想定しています
- 著作権法を遵守してご利用ください
- YouTubeの利用規約にしたがってご利用ください 