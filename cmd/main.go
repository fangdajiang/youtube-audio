package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/handler"
)

func main() {
	fmt.Println("Start fetching, converting, sending...")

	err := fetchVideo()
	//err := sendAudio()

	if err != nil {
		log.Fatalf("%s", err)
	}
}

func fetchVideo() error {
	// Add a flag
	var videoUrl string
	flag.StringVar(&videoUrl, "video-url", "", "This video will be downloaded.")
	flag.Parse()
	// download a video
	return handler.DownloadYouTubeAudio(videoUrl)
}
func sendAudio() error {
	// Add a flag
	var audioFilePath string
	flag.StringVar(&audioFilePath, "audio-file", "", "This audio file will be sent.")
	flag.Parse()
	// Send an audio file
	return handler.SendLocalAudioFile(audioFilePath)
}
