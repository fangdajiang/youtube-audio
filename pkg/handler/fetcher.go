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
	VideoId string
	RawUrl  string
}

func GetVideoIdsBy(playlistId string) []PlaylistVideosMetaData {
	youTubeCredentials, err := GenerateYouTubeCredentials()
	if err != nil {
		return []PlaylistVideosMetaData{}
	}

	ctx := context.Background()
	svc, err := youtubeapi.NewService(ctx, option.WithScopes(youtubeapi.YoutubeReadonlyScope), option.WithAPIKey(youTubeCredentials.Key))
	if err != nil {
		log.Errorf("new service error:%v", err)
		return []PlaylistVideosMetaData{}
	}

	var playlistVideosMetaDataArray []PlaylistVideosMetaData
	playlistResponse := playlistItemsList(svc, YouTubePart, playlistId, YouTubeMaxResults)

	for _, playlistItem := range playlistResponse.Items {
		title := playlistItem.Snippet.Title
		publishedAt := playlistItem.Snippet.PublishedAt
		position := playlistItem.Snippet.Position

		videoId := playlistItem.Snippet.ResourceId.VideoId
		log.Infof("%v, (%v) on %v at %v\r\n", title, videoId, position, publishedAt)

		videoMetaData := PlaylistVideosMetaData{videoId, MakeYouTubeRawUrl(videoId)}
		playlistVideosMetaDataArray = append(playlistVideosMetaDataArray, videoMetaData)
	}

	return playlistVideosMetaDataArray
}

func playlistItemsList(service *youtubeapi.Service, part []string, playlistId string, maxResults int64) *youtubeapi.PlaylistItemListResponse {
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	if maxResults > 0 && maxResults <= 50 {
		call = call.MaxResults(maxResults)
	} else {
		log.Warnf("illegal maxResults error:%v", maxResults)
	}
	response, err := call.Do()
	if err != nil {
		log.Fatalf("get playlist items error:%v, playlistId:%s", err, playlistId)
	}
	return response
}

func RetrieveITagOfMinimumAudioSize(mediaUrl string) (int, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(mediaUrl)
	log.Infof("video duration: %vs", video.Duration.Seconds())
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
		return -1, fmt.Errorf("audio track not found:%s", mediaUrl)
	} else if len(videoMetaDataArray) == 1 {
		log.Infof("Found only 1 audio track for %s, iTagNo: %v at %s", mediaUrl, videoMetaDataArray[0].ITagNo, time.Now().Format(DateTimeFormat))
		minSizeVideoMetaData = videoMetaDataArray[0]
	} else {
		log.Infof("Found %v audio tracks for %s, at %s", len(videoMetaDataArray), mediaUrl, time.Now().Format(DateTimeFormat))
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
	iTagNo, err := RetrieveITagOfMinimumAudioSize(mediaUrl)
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

	validMediaFileName, err := FilenamifyMediaTitle(result.Info.Title)
	if err != nil {
		return parcel, err
	}
	parcel = GenerateParcel(fmt.Sprintf("%s%s", ResourceStorePath, validMediaFileName), result.Info.Title)
	log.Debugf("parcel: %v", parcel)

	log.Infof("ready to CREATE media file %s at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	parcelFile, err := os.Create(parcel.FilePath)
	log.Infof("media file %s CREATED at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	if err != nil {
		log.Fatal(err)
	}
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
