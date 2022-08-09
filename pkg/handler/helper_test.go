package handler

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
	"youtube-audio/pkg/util"
)

func TestGetLocalDateTime(t *testing.T) {
	r := require.New(t)

	publishedAt := "2022-07-31T14:29:03Z"
	remotePublishedAt, err := time.Parse(YouTubeDateTimeFormat, publishedAt)
	r.NoError(err)
	localPublishedAt := GetLocalDateTime(publishedAt)

	log.Infof("remotePublishedAt:%s, localPublishedAt: %s", remotePublishedAt.Format(DateTimeFormat), localPublishedAt)

}

func TestFilenamifyMediaTitle(t *testing.T) {
	r := require.New(t)

	mediaTitle := "中文abc/标题\\_123!_def`_gh'_done"

	namifiedMediaTitle, err := FilenamifyMediaTitle(mediaTitle)
	r.NoError(err)
	r.Greater(len(namifiedMediaTitle), len(mediaTitle))
	log.Infof("mediaTitle: %v", len(mediaTitle))
	log.Infof("namifiedMediaTitle: %v", len(namifiedMediaTitle))
}

func TestFileExists(t *testing.T) {
	r := require.New(t)

	filePath := "/tmp/test.txt"

	exists, err := FileExists(filePath)
	r.NoError(err)
	r.True(exists, "file NOT exists: %s", filePath)
}

func TestGetYouTubeVideosFromPlaylists(t *testing.T) {
	playlistMetaDataArray := GetYouTubeVideosFromPlaylists()
	log.Infof("playlists count: %v", len(playlistMetaDataArray))
	var videoMetaDataArray []*PlaylistVideoMetaData
	for _, playlistMetaData := range playlistMetaDataArray {
		videoMetaDataArray = append(videoMetaDataArray, playlistMetaData.PlaylistVideoMetaDataArray...)
	}
	for _, video := range videoMetaDataArray {
		log.Infof("id:%v, position:%v", video.VideoId, video.Position)
	}
}

func TestFlushFetchHistory(t *testing.T) {
	r := require.New(t)

	deliveries := AssembleDeliveriesFromPlaylists()
	r.True(len(deliveries) > 0)
	log.Infof("count: %v, deliveries: %v", len(deliveries), deliveries)

	var tamperedDeliveries []Delivery
	for _, delivery := range deliveries {
		log.Infof("original delivery: %v", delivery)
		rand.Seed(time.Now().UnixNano())
		delivery.Done = rand.Float32() < 0.5
		tamperedDeliveries = append(tamperedDeliveries, delivery)
	}
	for _, delivery := range tamperedDeliveries {
		log.Infof("tampered delivery: %v", delivery)
	}

	FlushFetchHistory(deliveries)
}

func TestAssembleDeliveriesFromPlaylists(t *testing.T) {
	deliveries := AssembleDeliveriesFromPlaylists()
	log.Infof("deliveries: %v", deliveries)
}

func TestMergeHistoryFetchesInto(t *testing.T) {
	deliveries := MergeHistoryFetchesInto(AssembleDeliveriesFromPlaylists())
	log.Infof("merged deliveries: %v", deliveries)
}

func TestDeleteSliceElms(t *testing.T) {
	historyProps := util.MediaHistory
	log.Infof("historyProps count: %v, historyProps: %v", len(historyProps), historyProps)
	hp := util.HistoryProps{Id: "PLt-jD7OCbLJ0ZMwvQFZCSuNaUbT3GMqVJ"}
	result := DeleteSliceElms(historyProps, hp)
	log.Infof("result count: %v, result: %v", len(result), result)
	log.Infof("historyProps count2: %v", len(historyProps))

}
