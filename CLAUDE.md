# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YouTube Audio is a Go CLI application that automatically downloads YouTube videos as audio files and sends them to Telegram channels. It integrates with YouTube Data API for playlist management and Telegram Bot API for content delivery. The application supports scheduled fetching from playlists and on-demand single video processing.

## Architecture Overview

### Core Components

**CLI Layer (`cmd/`)**
- `ya.go`: Root Cobra command with config initialization and JSON logging support
- `run.go`: Main execution logic with three modes: `latest`, `single`, `playlist`
- `version.go`: Version information with JSON output option

**Handler Layer (`pkg/handler/`)**
- `fetcher.go`: YouTube API integration, video downloading via yt-dlp/goutubedl, audio conversion with ffmpeg
- `deliveryman.go`: Telegram Bot API integration, audio file delivery, delivery state management

**Resource Management (`pkg/util/resource/`)**
- Manages external JSON configurations from Alicloud OSS
- `fetch_base.json`: Playlist configurations, download paths, metadata
- `fetch_history.json`: Tracks processed videos to prevent duplicates

**Utility Modules (`pkg/util/`)**
- `env/`: Environment variable handling for credentials
- `oss/`: Alicloud OSS integration for config storage
- `db/`: Database connectivity (optional feature)
- `log/`: Structured logging with JSON format support
- `myio/`: File operations and cleanup utilities

**Reporter (`pkg/reporter/`)**
- Tracks processing statistics (successful/failed fetches)
- Generates summary reports sent to Telegram

### Data Flow

1. **Initialization**: Load configurations from OSS, parse fetch_base.json for playlists
2. **Fetching**: Query YouTube API for latest videos, compare against fetch_history.json
3. **Processing**: Download audio via yt-dlp, convert to MP3 with metadata using ffmpeg
4. **Delivery**: Send audio files to Telegram channel with thumbnail and metadata
5. **Tracking**: Update fetch_history.json with processed videos, send summary report

### External Dependencies

- **yt-dlp**: Video/audio downloading (binary in `bin/dependency/`)
- **ffmpeg/ffprobe**: Audio conversion and metadata handling (binaries in `bin/dependency/`)
- **YouTube Data API**: Playlist and video metadata retrieval
- **Telegram Bot API**: Audio file delivery to channels
- **Alicloud OSS**: Configuration storage and logging

## Development Commands

### Build
```bash
# Standard build
go build -o bin/ya main.go

# Cross-compile for Linux (required for ARM-based development machines)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ya main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./pkg/handler/
go test ./pkg/util/
```

### Running
```bash
# Set required environment variables first:
# BOT_TOKEN, BOT_CHAT_ID, CHAT_ID, YOUTUBE_KEY, ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY

# Process latest videos from configured playlists
go run main.go run -m latest

# Process single YouTube video
go run main.go run -m single https://www.youtube.com/watch?v=xxx

# Process entire playlist
go run main.go run -m playlist https://www.youtube.com/playlist?list=xxx

# Enable JSON logging
go run main.go run -m latest --log-as-json
```

### Docker
```bash
# Build Docker image
docker build -t youtube-audio:latest -f ./Dockerfile .

# Run with environment variables
docker run -d \
  -e BOT_TOKEN=xxx \
  -e BOT_CHAT_ID=xxx \
  -e CHAT_ID=xxx \
  -e YOUTUBE_KEY=xxx \
  -e ALICLOUD_ACCESS_KEY=xxx \
  -e ALICLOUD_SECRET_KEY=xxx \
  youtube-audio:latest
```

## Configuration Management

The application relies on two external JSON files stored in Alicloud OSS:

- `fetch_base.json`: Defines YouTube playlists to monitor, download paths, and metadata
- `fetch_history.json`: Tracks processing history to prevent duplicate downloads

These files are loaded at runtime via OSS API. Local development requires valid Alicloud credentials and access to the configured OSS bucket.

## Environment Variables

Required for all operations:
- `BOT_TOKEN`: Telegram bot token
- `BOT_CHAT_ID`: Telegram bot chat ID for alerts
- `CHAT_ID`: Telegram channel ID for audio delivery
- `YOUTUBE_KEY`: YouTube Data API key
- `ALICLOUD_ACCESS_KEY`: OSS access credentials
- `ALICLOUD_SECRET_KEY`: OSS secret credentials

## Key Implementation Notes

- The application uses goroutines for concurrent video processing
- ffmpeg conversion includes metadata tagging (artist, title, album)
- Telegram file size limit (50MB) is enforced during processing
- Audio quality is configurable via QualityScaleFrom0To9 constant (currently set to "6")
- History tracking prevents reprocessing videos within 48-hour windows
- Failed operations trigger warning messages to the Telegram bot chat