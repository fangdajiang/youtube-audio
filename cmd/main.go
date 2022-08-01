package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"youtube-audio/pkg/handler"
)

func main() {
	fmt.Printf("Start fetching, converting, sending... from %s\n", time.Now().Format(handler.DateTimeFormat))

	process()

}

func processOneVideo(videoUrl string) {
	audioFile, err := fetchAudio(videoUrl)
	if err != nil {
		log.Warnf("Failed to download audio url %s from YouTube, error: %v", videoUrl, err)
		sendMessage(videoUrl, handler.FailedToDownloadAudioWarningTemplate)
		return
	}
	if !handler.IsAudioValid(audioFile) {
		log.Warnf("Downloaded audio url %s from YouTube is NOT valid, error: %v", videoUrl, err)
		sendMessage(audioFile.FilePath, handler.InvalidDownloadedAudioWarningTemplate)
		return
	}

	err = sendAudio(audioFile)

	if err != nil {
		log.Warnf("Failed to send file %s to telegram channel, error: %v", audioFile.FilePath, err)
		audioFile.Caption = audioFile.Caption + fmt.Sprintf("%s", err)
		sendMessage(audioFile.Caption, handler.FailedToSendAudioWarningTemplate)
		handler.Cleanup(audioFile)
		return
	}

	handler.Cleanup(audioFile)

	log.Infof("\r\n")
}

func process() {
	videoMetaDataArray := handler.GetVideoIdsBy(handler.YouTubeChannelId)

	for i, videoMetaData := range videoMetaDataArray {
		size := len(videoMetaDataArray)

		if i < size-1 { //have to?
			go processOneVideo(videoMetaData.RawUrl)
		} else {
			processOneVideo(videoMetaData.RawUrl)
		}
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
