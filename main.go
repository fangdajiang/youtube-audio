package main

import (
	"log"

	"github.com/subosito/gotenv"
	"github.com/wader/goutubedl"

	"youtube-audio/cmd"
)

var (
	version = "0.2.0"
)

func main() {
	if err := gotenv.Load(); err != nil {
		log.Printf("dotenv load warning: %v", err)
	}
	goutubedl.Path = "yt-dlp"
	cmd.Execute(version)
}
