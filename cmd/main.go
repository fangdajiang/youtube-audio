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
	playlistMetaDataArray := handler.GetYouTubeVideosFromPlaylists()
	var videoMetaDataArray []*handler.PlaylistVideoMetaData
	for _, playlistMetaData := range playlistMetaDataArray {
		videoMetaDataArray = append(videoMetaDataArray, playlistMetaData.PlaylistVideoMetaDataArray...)
	}
	videosCount := len(videoMetaDataArray)
	log.Infof("total playlists: %v, videos: %v", len(playlistMetaDataArray), videosCount)

	var wg sync.WaitGroup
	for i, videoMetaData := range videoMetaDataArray {
		log.Infof("%v, raw url: %s", i, videoMetaData.RawUrl)
		if i < videosCount-1 { //have to?
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
	handler.FlushFetchHistory(playlistMetaDataArray)
}
