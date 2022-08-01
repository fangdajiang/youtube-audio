package handler

import (
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetVideoMetaDataArrayBy(t *testing.T) {
	r := require.New(t)

	playlistId := "PLKFNuj0yup6ng8YSmsM5aUFrtIkDQfjIM"
	//playlistId := "PLD_nomDtqAAftttx00BRUDCzDlqyAPRoG"

	playlistVideosMetaDataArray := GetVideoMetaDataArrayBy(playlistId)
	maxResultsCount := GetYouTubeChannelMaxResultsCount(playlistId)

	r.Len(playlistVideosMetaDataArray, int(maxResultsCount))

	for _, videoMetaData := range playlistVideosMetaDataArray {
		log.Infof("videoMetaData position: %v", videoMetaData.Position)
	}
}

func TestYouTube_GetItagInfo(t *testing.T) {
	r := require.New(t)
	client := youtube.Client{}

	url := "https://www.youtube.com/watch?v=Y-EX1u34E2M"
	video, err := client.GetVideo(url)
	r.NoError(err)
	//r.Len(video.Formats, 24)

	for i, f := range video.Formats {
		log.Infof("i: %v, ItagNo:%v, ADM:%s, FPS:%v, QL:%s, AQ:%s, AC:%v, AverBit:%v, Bit:%v, Size:%v",
			i, f.ItagNo, f.ApproxDurationMs, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.AverageBitrate, f.Bitrate, f.ContentLength)
	}
}

func TestDownloadYouTubeAudioToPath(t *testing.T) {
	r := require.New(t)

	parcel, err := DownloadYouTubeAudioToPath("https://www.youtube.com/watch?v=gqVnTo7aLEE")
	r.NoError(err)
	r.Equal("摸着石头过河，后来呢？/王剑每日观察/短视频", parcel.Caption)

	Cleanup(parcel)
}

func TestRetrieveITagOfMinimumAudioSize(t *testing.T) {
	r := require.New(t)

	iTagNo, err := RetrieveITagOfMinimumAudioSize("https://www.youtube.com/watch?v=Y-EX1u34E2M")
	r.NoError(err)
	//r.Equal(249, iTagNo)
	log.Infof("iTagNo:%v", iTagNo)
}
