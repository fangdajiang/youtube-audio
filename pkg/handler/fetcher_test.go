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

	url := "https://www.youtube.com/watch?v=I-4CCOLvE1g"
	video, err := client.GetVideo(url)
	r.NoError(err)
	//r.Len(video.Formats, 13)

	for i, f := range video.Formats {
		log.Infof("i: %v, ItagNo:%v, FPS:%v, QL:%s, AQ:%s, AC:%v, MimeType:%s",
			i, f.ItagNo, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.MimeType)
	}
}

func TestDownloadYouTubeAudioToPath(t *testing.T) {
	r := require.New(t)

	parcel, err := DownloadYouTubeAudioToPath("https://www.youtube.com/watch?v=gqVnTo7aLEE")
	r.NoError(err)
	r.Equal("摸着石头过河，后来呢？/王剑每日观察/短视频", parcel.Caption)

	Cleanup(parcel)
}

func TestRetrieveFirstAudioITag(t *testing.T) {
	r := require.New(t)

	iTagNo, err := RetrieveFirstAudioITag("https://www.youtube.com/watch?v=BTDt-OSzFZc")
	r.NoError(err)
	r.Equal(140, iTagNo)
}
