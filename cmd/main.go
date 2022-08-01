package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"youtube-audio/pkg/handler"
	"youtube-audio/pkg/util"
)

func main() {
	fmt.Printf("Start fetching, converting, sending... from %s\n", time.Now().Format(handler.DateTimeFormat))

	process()

}

func init() {
	util.InitResources()
	log.Infof("channel: %v", util.MediaChannels[0])
	log.Infof("history: %v", util.MediaHistory[0])
}

func process() {
	videoMetaDataArray := handler.GetVideoIdsBy(handler.YouTubeChannelId)

	for i, videoMetaData := range videoMetaDataArray {
		size := len(videoMetaDataArray)

		if i < size-1 { //have to?
			go handler.ProcessOneVideo(videoMetaData.RawUrl)
		} else {
			handler.ProcessOneVideo(videoMetaData.RawUrl)
		}
	}
}
