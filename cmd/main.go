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
	videoMetaDataArray := handler.GetVideoIdsBy(handler.YouTubeChannelId)
	for _, videoMetaData := range videoMetaDataArray {
		audioFile, err := fetchAudio(videoMetaData.RawUrl)
		if err != nil {
			log.Warnf("Failed to download audio url %s from YouTube, error: %v", videoMetaData.RawUrl, err)
			sendMessage(videoMetaData.RawUrl, handler.FailedToDownloadAudioWarningTemplate)
			continue
		}

		err = sendAudio(audioFile)
		if err != nil {
			log.Warnf("Failed to send file %s to telegram channel, error: %v", audioFile.FilePath, err)
			audioFile.Caption = audioFile.Caption + fmt.Sprintf("%s", err)
			sendMessage(audioFile.Caption, handler.FailedToSendAudioWarningTemplate)
			handler.Cleanup(audioFile)
			continue
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

func sendMessage(desc string, warningMessage string) {
	telegramBot, err := handler.GenerateTelegramBot()
	if err != nil {
		log.Fatalf("%s", err)
	}
	telegramBot.SendWarningMessage(desc, warningMessage)
}
