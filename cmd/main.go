package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/handler"
)

func main() {
	fmt.Println("Start fetching, converting, sending...")

	process()

}

func process() {
	videoIds := handler.GetVideoIdsBy(handler.YouTubeChannelId)
	for _, videoId := range videoIds {
		rawUrl := handler.MakeYouTubeRawUrl(videoId)
		audioFile, err := fetchAudio(rawUrl)
		if err != nil {
			log.Fatalf("%s", err)
		}

		err = sendAudio(audioFile)
		if err != nil {
			log.Fatalf("%s", err)
		}

		handler.Cleanup(audioFile)

		log.Infof("\r\n")
	}
}

func fetchAudio(rawUrl string) (handler.Parcel, error) {
	// download a video
	return handler.DownloadYouTubeAudioToPath(rawUrl)
}

func sendAudio(parcel handler.Parcel) error {
	telegramBot, err := handler.GenerateTelegramBot()
	if err != nil {
		log.Fatalf("%s", err)
	}
	// Send an audio file
	return telegramBot.Send(parcel)
}
