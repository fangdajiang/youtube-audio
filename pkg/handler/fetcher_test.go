package handler

import (
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"youtube-audio/pkg/util"
)

func init() {
	util.InitResources()
}

func TestFindChannelId(t *testing.T) {
	r := require.New(t)
	svc, err := GetYouTubeService()
	r.NoError(err)

	channelAlias := "德国自干五"
	queryType := "channel"

	call := svc.Search.List(YouTubePart)
	call = call.Q(channelAlias).Type(queryType)
	response, err := call.Do()
	r.NoError(err)
	log.Infof("len: %v", len(response.Items))
	for _, item := range response.Items {
		log.Infof("channelId: %s, channelTitle: %s, description: %s, ", item.Snippet.ChannelId, item.Snippet.ChannelTitle, item.Snippet.Description)
	}

}

func TestGetVideoMetaDataArrayBy(t *testing.T) {
	r := require.New(t)

	playlistId := "PL_gom9iTTcZrCXj4niVYgAdkTbybJpQQR"

	playlistMetaData := GetPlaylistMetaDataBy(playlistId)
	maxResultsCount := GetYouTubePlaylistMaxResultsCount(playlistId)

	r.Len(playlistMetaData.PlaylistVideoMetaDataArray, int(maxResultsCount))

	for _, videoMetaData := range playlistMetaData.PlaylistVideoMetaDataArray {
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

	iTagNo, err := RetrieveITagOfMinimumSizeAudio("https://www.youtube.com/watch?v=Y-EX1u34E2M")
	r.NoError(err)
	//r.Equal(249, iTagNo)
	log.Infof("iTagNo:%v", iTagNo)
}
