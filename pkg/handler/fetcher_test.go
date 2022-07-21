package handler

import (
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestYoutube_GetVideosByPlaylistId(t *testing.T) {
	r := require.New(t)

	videoIds := GetVideoIdsBy("UU8UCbiPrm2zN9nZHKdTevZA")
	r.Len(videoIds, 5)

	for _, videoId := range videoIds {
		log.Infof("videoId: %s", videoId)
	}
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

func TestYoutube_GetItagInfo(t *testing.T) {
	r := require.New(t)
	client := youtube.Client{}

	url := "https://www.youtube.com/watch?v=rFejpH_tAHM"
	video, err := client.GetVideo(url)
	r.NoError(err)
	r.Len(video.Formats, 24)

	for _, f := range video.Formats {
		log.Infof("ItagNo:%v, FPS:%v, VQ:%s, AQ:%s, Size:%v, MimeType:%s",
			f.ItagNo, f.FPS, f.QualityLabel, f.AudioQuality, f.ContentLength, f.MimeType)
	}
}
