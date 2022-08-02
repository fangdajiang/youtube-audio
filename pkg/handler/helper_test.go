package handler

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

func TestGetYouTubePlaylistsAllVideos(t *testing.T) {
	playlistVideosMetaDataArray := GetYouTubePlaylistsAllVideos()
	log.Infof("length: %v", len(playlistVideosMetaDataArray))
	for _, video := range playlistVideosMetaDataArray {
		log.Infof("id:%v, position:%v", video.VideoId, video.Position)
	}

}
