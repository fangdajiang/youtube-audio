package handler

import (
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetVideoIdsBy(t *testing.T) {
	r := require.New(t)

	videoIds := GetVideoIdsBy("UU8UCbiPrm2zN9nZHKdTevZA")
	r.Len(videoIds, 5)

	for _, videoId := range videoIds {
		log.Infof("videoId: %s", videoId)
	}
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

func TestDownloadYouTubeAudioToPath(t *testing.T) {
	r := require.New(t)

	parcel, err := DownloadYouTubeAudioToPath("https://www.youtube.com/watch?v=gqVnTo7aLEE")
	r.NoError(err)
	r.Equal("摸着石头过河，后来呢？/王剑每日观察/短视频", parcel.caption)

	Cleanup(parcel)
}
