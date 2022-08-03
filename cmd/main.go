package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
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
	log.Infof("base: %v", util.MediaBase[0])
	log.Infof("history: %v", util.MediaHistory[0])
}

func process() {
	videoMetaDataArray := handler.GetYouTubePlaylistsAllVideos()
	log.Infof("total videos: %v", len(videoMetaDataArray))

	var wg sync.WaitGroup
	for i, videoMetaData := range videoMetaDataArray {
		size := len(videoMetaDataArray)
		if i < size-1 { //have to?
			wg.Add(1)
			go func(rawUrl string) {
				handler.ProcessOneVideo(rawUrl)
				wg.Done()
			}(videoMetaData.RawUrl)
		} else {
			handler.ProcessOneVideo(videoMetaData.RawUrl)
			wg.Wait()
		}
	}
	log.Infof("ALL %v videos proccessed", len(videoMetaDataArray))
}
