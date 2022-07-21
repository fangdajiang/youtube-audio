package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/handler"
)

func main() {
	fmt.Println("Start fetching, converting, sending...")

	process()

}

func process() {
	audioFile, err := fetchAudio()
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = sendAudio(audioFile)
	if err != nil {
		log.Fatalf("%s", err)
	}

	handler.Cleanup(audioFile)
}

func fetchAudio() (handler.Parcel, error) {
	// Add a flag
	var videoUrl string
	flag.StringVar(&videoUrl, "video-url", "", "This video will be downloaded.")
	flag.Parse()
	// download a video
	return handler.DownloadYouTubeAudioToPath(videoUrl)
}

func sendAudio(parcel handler.Parcel) error {
	telegramBot, err := handler.GenerateTelegramBot()
	if err != nil {
		log.Fatalf("%s", err)
	}
	// Send an audio file
	return telegramBot.Send(parcel)
}
