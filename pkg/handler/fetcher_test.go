package handler

import (
	"github.com/kkdai/youtube/v2"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

func init() {
	resource.InitResources()
	log.InitLogging()
}

func TestGetYouTubeVideosFromPlaylist(t *testing.T) {
	playlistMetaDataArray := GetYouTubeVideosFromPlaylistId("PLD_nomDtqAAc-v9CctRQEOSPXQPIy_JD2")
	var videoMetaDataArray []*PlaylistVideoMetaData
	for _, playlistMetaData := range playlistMetaDataArray {
		videoMetaDataArray = append(videoMetaDataArray, playlistMetaData.PlaylistVideoMetaDataArray...)
	}
	for _, video := range videoMetaDataArray {
		log.Debugf("id:%v, position:%v", video.VideoId, video.Position)
	}
}

func TestGetYouTubeVideosFromPlaylists(t *testing.T) {
	playlistMetaDataArray := GetYouTubeVideosFromPlaylists()
	log.Debugf("playlists count: %v", len(playlistMetaDataArray))
	var videoMetaDataArray []*PlaylistVideoMetaData
	for _, playlistMetaData := range playlistMetaDataArray {
		videoMetaDataArray = append(videoMetaDataArray, playlistMetaData.PlaylistVideoMetaDataArray...)
	}
	for _, video := range videoMetaDataArray {
		log.Debugf("id:%v, position:%v", video.VideoId, video.Position)
	}
}

func TestFlushFetchHistory(t *testing.T) {
	r := require.New(t)

	deliveries := AssembleDeliveriesFromPlaylists(GetYouTubeVideosFromPlaylists())
	r.True(len(deliveries) > 0)
	log.Debugf("count: %v, deliveries: %v", len(deliveries), deliveries)

	var tamperedDeliveries []Delivery
	for _, delivery := range deliveries {
		log.Debugf("original delivery: %v", delivery)
		rand.Seed(time.Now().UnixNano())
		delivery.Done = rand.Float32() < 0.5
		tamperedDeliveries = append(tamperedDeliveries, delivery)
	}
	for _, delivery := range tamperedDeliveries {
		log.Debugf("tampered delivery: %v", delivery)
	}

	FlushFetchHistory(deliveries)
}

func TestAssembleDeliveriesFromPlaylists(t *testing.T) {
	deliveries := AssembleDeliveriesFromPlaylists(GetYouTubeVideosFromPlaylistId("PLstzraCE5l2j_Sih-L9CoFq-r71NflIfi"))
	log.Debugf("deliveries: %v", deliveries)
}

func TestFindChannelId(t *testing.T) {
	r := require.New(t)
	svc, err := GetYouTubeService()
	r.NoError(err)

	channelAlias := "德国自干五"
	queryType := "channel"

	call := svc.Search.List(util.YouTubePart)
	call = call.Q(channelAlias).Type(queryType)
	response, err := call.Do()
	r.NoError(err)
	log.Debugf("len: %v", len(response.Items))
	for _, item := range response.Items {
		log.Debugf("channelId: %s, channelTitle: %s, description: %s, ", item.Snippet.ChannelId, item.Snippet.ChannelTitle, item.Snippet.Description)
	}

}

func TestGetVideoMetaDataArrayBy(t *testing.T) {
	r := require.New(t)

	playlistId := "PL_gom9iTTcZrCXj4niVYgAdkTbybJpQQR"

	playlistMetaData := GetPlaylistMetaDataBy(playlistId)
	maxResultsCount := util.GetYouTubePlaylistMaxResultsCount(playlistId)

	r.Len(playlistMetaData.PlaylistVideoMetaDataArray, int(maxResultsCount))

	for _, videoMetaData := range playlistMetaData.PlaylistVideoMetaDataArray {
		log.Debugf("videoMetaData position: %v", videoMetaData.Position)
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
		log.Debugf("i: %v, ItagNo:%v, ADM:%s, FPS:%v, QL:%s, AQ:%s, AC:%v, AverBit:%v, Bit:%v, Size:%v",
			i, f.ItagNo, f.ApproxDurationMs, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.AverageBitrate, f.Bitrate, f.ContentLength)
	}
}

func TestDownloadYouTubeAudioToPath(t *testing.T) {
	r := require.New(t)

	de := AssembleDeliveryFromSingleUrl("https://www.youtube.com/watch?v=gqVnTo7aLEE")
	parcel, err := DownloadYouTubeAudioToPath(&de)
	r.NoError(err)
	r.Equal("摸着石头过河，后来呢？/王剑每日观察/短视频", parcel.Caption)

	myio.Cleanup(parcel.FilePath)
}

func TestRetrieveITagOfMinimumAudioSize(t *testing.T) {
	r := require.New(t)

	iTagNo, err := RetrieveITagOfMinimumSizeAudio("https://www.youtube.com/watch?v=Y-EX1u34E2M")
	r.NoError(err)
	//r.Equal(249, iTagNo)
	log.Debugf("iTagNo:%v", iTagNo)
}

func TestMergeHistoryFetchesInto(t *testing.T) {
	deliveries := MergeHistoryFetchesInto(AssembleDeliveriesFromPlaylists(GetYouTubeVideosFromPlaylists()))
	log.Debugf("merged deliveries: %v", deliveries)
}

func Test_setAudioMetadata(t *testing.T) {
	// 创建一个临时文件路径
	filePath := "/Users/fangdajiang/Desktop/test.mp4"
	caption := "摸着石头过河"
	artist := "FDJ"
	album := "千钧一发"
	parcel = GenerateParcel(filePath, caption, artist, album, "mediaUrl")

	// 测试ffmpeg命令成功执行的情况
	parcel, _ = convertToMp3AndFillMetadata(parcel)
	log.Debugf("parcel: %v", parcel)

}
