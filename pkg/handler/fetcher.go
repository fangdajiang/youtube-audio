package handler

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"
	"io"
	"os"
	"strconv"
	"time"
)

//ITagNo format id
//  for _, f := range result.Formats() {
//		log.Infof("format id:%s", f.FormatID)
//	}

type YouTubeCredentials struct {
	Key string
}

type VideoMetaData struct {
	FPS              int
	ITagNo           int
	Bitrate          int
	AverageBitrate   int
	ContentLength    int64
	ApproxDurationMs string
	AudioChannels    int
}

type PlaylistVideosMetaData struct {
	VideoId  string
	RawUrl   string
	Position int64
}
type PlaylistVideosMetaDataArray []*PlaylistVideosMetaData

func (s PlaylistVideosMetaDataArray) Len() int {
	return len(s)
}
func (s PlaylistVideosMetaDataArray) Less(i, j int) bool {
	return s[i].Position > s[j].Position
}
func (s PlaylistVideosMetaDataArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func ProcessOneVideo(videoUrl string) {
	audioFile, err := fetchAudio(videoUrl)
	if err != nil {
		log.Warnf("Failed to download audio url %s from YouTube, error: %v", videoUrl, err)
		SendMessage(videoUrl, FailedToDownloadAudioWarningTemplate)
	}
	if !IsAudioValid(audioFile) {
		log.Warnf("Downloaded audio url %s from YouTube is NOT valid", videoUrl)
		SendMessage(audioFile.FilePath, InvalidDownloadedAudioWarningTemplate)
	}

	err = SendAudio(audioFile)

	if err != nil {
		log.Warnf("Failed to send file %s to telegram channel, error: %v", audioFile.FilePath, err)
		audioFile.Caption = audioFile.Caption + fmt.Sprintf("%s", err)
		SendMessage(audioFile.Caption, FailedToSendAudioWarningTemplate)
	}

	Cleanup(audioFile)
	log.Infof("\r\n")
}

func fetchAudio(rawUrl string) (Parcel, error) {
	// download a video
	return DownloadYouTubeAudioToPath(rawUrl)
}

func GetYouTubeService() (*youtubeapi.Service, error) {
	youTubeCredentials, err := GenerateYouTubeCredentials()
	if err != nil {
		log.Errorf("generate youtube credentials error:%v", err)
		return nil, err
	}

	ctx := context.Background()
	svc, err := youtubeapi.NewService(ctx, option.WithScopes(youtubeapi.YoutubeReadonlyScope), option.WithAPIKey(youTubeCredentials.Key))
	if err != nil {
		log.Errorf("new service error:%v", err)
		return nil, err
	}

	return svc, nil
}

func GetVideoMetaDataArrayBy(playlistId string) PlaylistVideosMetaDataArray {
	var playlistVideosMetaDataArray PlaylistVideosMetaDataArray
	svc, err := GetYouTubeService()
	if err != nil {
		log.Errorf("get youtube service error:%v", err)
		return playlistVideosMetaDataArray
	}

	playlistResponse := playlistItemsList(svc, YouTubePart, playlistId, GetYouTubePlaylistMaxResultsCount(playlistId))

	for _, playlistItem := range playlistResponse.Items {
		publishedAt := playlistItem.Snippet.PublishedAt
		title := playlistItem.Snippet.Title
		localPublishedAt := GetLocalDateTime(publishedAt)
		channelTitle := playlistItem.Snippet.ChannelTitle
		channelId := playlistItem.Snippet.ChannelId

		videoId := playlistItem.Snippet.ResourceId.VideoId
		position := playlistItem.Snippet.Position
		log.Infof("%s(%s) from %s(%s) on position %v was published at %s\r\n", title, videoId, channelTitle, channelId, position, localPublishedAt)

		videoMetaData := PlaylistVideosMetaData{videoId, MakeYouTubeRawUrl(videoId), position}
		playlistVideosMetaDataArray = append(playlistVideosMetaDataArray, &videoMetaData)
	}

	return playlistVideosMetaDataArray
}

//type queryParamOpt struct {
//	key, value string
//}

//func (qp queryParamOpt) Get() (string, string) {
//	return qp.key, qp.value
//}

func playlistItemsList(service *youtubeapi.Service, part []string, playlistId string, maxResults int64) *youtubeapi.PlaylistItemListResponse {
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	if maxResults <= 0 || maxResults > FetchYouTubeMaxResultsLimit {
		log.Errorf("illegal maxResults error:%v", maxResults)
		maxResults = YouTubeDefaultMaxResults
	}
	call = call.MaxResults(maxResults)
	//lastUpdated := queryParamOpt{key: "order", value: "time"}
	response, err := call.Do()
	if err != nil {
		log.Fatalf("get playlist items error:%v, playlistId:%s", err, playlistId)
	}
	return response
}

func RetrieveITagOfMinimumSizeAudio(mediaUrl string) (int, error) {
	client := youtube.Client{}

	log.Infof("Ready to get video: %s at %s", mediaUrl, time.Now().Format(DateTimeFormat))
	video, err := client.GetVideo(mediaUrl)
	log.Infof("video duration: %vs at %s", video.Duration.Seconds(), time.Now().Format(DateTimeFormat))
	if err != nil {
		return -1, fmt.Errorf("failed to get video, error:%s, mediaUrl:%s", err, mediaUrl)
	}
	var videoMetaDataArray []VideoMetaData
	for _, f := range video.Formats {
		log.Infof("ItagNo:%v, ADM:%s, FPS:%v, QL:%s, AQ:%s, AC:%v, AverBit:%v, Bit:%v, Size:%v",
			f.ItagNo, f.ApproxDurationMs, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.AverageBitrate, f.Bitrate, f.ContentLength)
		if f.FPS == 0 {
			videoMetaData := VideoMetaData{f.FPS, f.ItagNo, f.Bitrate, f.AverageBitrate, f.ContentLength,
				f.ApproxDurationMs, f.AudioChannels}
			bitrate := videoMetaData.AverageBitrate
			if bitrate == 0 {
				// Some formats don't have the average bitrate
				bitrate = videoMetaData.Bitrate
			}
			size := videoMetaData.ContentLength
			if size == 0 {
				// Some formats don't have this information
				size = int64(float64(bitrate) * video.Duration.Seconds() / 8)
				videoMetaData.ContentLength = size
			}
			videoMetaDataArray = append(videoMetaDataArray, videoMetaData)
		}
	}
	maxSizeVideoMetaData := VideoMetaData{}
	minSizeVideoMetaData := VideoMetaData{
		FPS:              0,
		ITagNo:           0,
		Bitrate:          0,
		AverageBitrate:   0,
		ContentLength:    UploadAudioMaxLength,
		ApproxDurationMs: "",
		AudioChannels:    0,
	}
	if len(videoMetaDataArray) == 0 {
		return -1, fmt.Errorf("proper audio track not found:%s", mediaUrl)
	} else if len(videoMetaDataArray) == 1 {
		log.Infof("Found only 1 proper audio track for %s, iTagNo: %v at %s", mediaUrl, videoMetaDataArray[0].ITagNo, time.Now().Format(DateTimeFormat))
		minSizeVideoMetaData = videoMetaDataArray[0]
	} else {
		log.Infof("Found %v proper audio tracks for %s, at %s", len(videoMetaDataArray), mediaUrl, time.Now().Format(DateTimeFormat))
		for _, v := range videoMetaDataArray {
			if v.ContentLength < minSizeVideoMetaData.ContentLength {
				minSizeVideoMetaData = v
			} else {
				maxSizeVideoMetaData = v
			}
		}
		log.Infof("maxSizeVideoMetaData: %v", maxSizeVideoMetaData)
	}
	log.Infof("minSizeVideoMetaData: %v", minSizeVideoMetaData)
	if UploadAudioMaxLength < minSizeVideoMetaData.ContentLength {
		return -1, fmt.Errorf("the min size %v of audio track EXCEEDS the max %v, url:%s",
			minSizeVideoMetaData.ContentLength, UploadAudioMaxLength, mediaUrl)
	}
	return minSizeVideoMetaData.ITagNo, nil
}

func DownloadYouTubeAudioToPath(mediaUrl string) (Parcel, error) {
	var parcel Parcel
	log.Infof("Ready to download media %s at %s", mediaUrl, time.Now().Format(DateTimeFormat))
	result, err := goutubedl.New(context.Background(), mediaUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %s", mediaUrl)
	}

	validMediaFileName, err := FilenamifyMediaTitle(result.Info.Title)
	if err != nil {
		return parcel, err
	}
	parcel = GenerateParcel(fmt.Sprintf("%s%s", GetYouTubePlaylists().DownloadedFilesPath, validMediaFileName), result.Info.Title)
	log.Infof("generated parcel: %v", parcel)

	log.Infof("ready to CREATE media file %s at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	parcelFile, err := os.Create(parcel.FilePath)
	log.Infof("media file %s CREATED at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	if err != nil {
		log.Fatal(err)
	}

	iTagNo, err := RetrieveITagOfMinimumSizeAudio(mediaUrl)
	if err != nil {
		log.Errorf("retrieve iTag error, iTagNo: %v, error: %s", iTagNo, err)
		return parcel, fmt.Errorf("goutubedl download error: %s, ITagNo: %v", mediaUrl, iTagNo)
	}
	downloadedResult, err := result.Download(context.Background(), strconv.Itoa(iTagNo))
	if err != nil {
		log.Errorf("download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %s, ITagNo: %v", mediaUrl, iTagNo)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)
	log.Infof("downloading media %s from %s", result.Info.Title, time.Now().Format(DateTimeFormat))

	log.Infof("ready to COPY media file %s at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	written, err := io.Copy(parcelFile, downloadedResult)
	log.Infof("media file %s DOWNLOADED & COPIED till %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	if err != nil {
		log.Fatalf("copy error: %s, parcel: %s, written: %v", err, parcel, written)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(parcelFile)

	return parcel, nil
}
