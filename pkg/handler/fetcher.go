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
	VideoId string
	ITag    string
}

func GetVideoIdsBy(playlistId string) []VideoMetaData {
	youTubeCredentials, err := GenerateYouTubeCredentials()
	if err != nil {
		return []VideoMetaData{}
	}

	ctx := context.Background()
	svc, err := youtubeapi.NewService(ctx, option.WithScopes(youtubeapi.YoutubeReadonlyScope), option.WithAPIKey(youTubeCredentials.Key))
	if err != nil {
		log.Errorf("new service error:%v", err)
		return []VideoMetaData{}
	}

	var videoMetaDataArray []VideoMetaData
	playlistResponse := playlistItemsList(svc, YouTubePart, playlistId, YouTubeMaxResults)

	for _, playlistItem := range playlistResponse.Items {
		title := playlistItem.Snippet.Title
		publishedAt := playlistItem.Snippet.PublishedAt
		position := playlistItem.Snippet.Position

		videoId := playlistItem.Snippet.ResourceId.VideoId
		log.Infof("%v, (%v) on %v at %v\r\n", title, videoId, position, publishedAt)

		videoMetaData := VideoMetaData{videoId, "123"}
		videoMetaDataArray = append(videoMetaDataArray, videoMetaData)
	}

	return videoMetaDataArray
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

func RetrieveFirstAudioITag(mediaUrl string) (int, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(mediaUrl)
	if err != nil {
		return -1, fmt.Errorf("retrieve first audio track error:%s, mediaUrl:%s", err, mediaUrl)
	}
	for i, f := range video.Formats {
		log.Infof("i: %v, ItagNo:%v, FPS:%v, QL:%s, AQ:%s, AC:%v, MimeType:%s",
			i, f.ItagNo, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.MimeType)
		if f.FPS == 0 {
			return f.ItagNo, nil
		}
	}
	return -1, fmt.Errorf("audio track not found:%s", mediaUrl)
}

func DownloadYouTubeAudioToPath(mediaUrl string) (Parcel, error) {
	var parcel Parcel
	log.Infof("Ready to downlod media %s at %s", mediaUrl, time.Now().Format(DateTimeFormat))
	result, err := goutubedl.New(context.Background(), mediaUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %s", mediaUrl)
	}
	iTagNo, err := RetrieveFirstAudioITag(mediaUrl)
	downloadedResult, err := result.Download(context.Background(), strconv.Itoa(iTagNo))
	if err != nil {
		log.Errorf("download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %s, ITagNo: %v", mediaUrl, iTagNo)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)
	log.Infof("media %s downloaded at %s", result.Info.Title, time.Now().Format(DateTimeFormat))

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
	log.Infof("media file %s COPIED at %s", parcel.FilePath, time.Now().Format(DateTimeFormat))
	if err != nil {
		log.Fatalf("copy error: %s, parcel: %s, written: %v", err, parcel, written)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(parcelFile)

	return parcel, nil
}
