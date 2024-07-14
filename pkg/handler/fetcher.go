package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/wader/goutubedl"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
	io2 "youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

const (
	QualityScaleFrom0To9 string = "6"
)

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

type PlaylistVideoMetaData struct {
	Artist   string
	Album    string
	VideoId  string
	RawUrl   string
	Position int64
}
type PlaylistMetaData struct {
	PlaylistId                 string
	PlaylistVideoMetaDataArray []*PlaylistVideoMetaData
}

func (s PlaylistMetaData) Len() int {
	return len(s.PlaylistVideoMetaDataArray)
}
func (s PlaylistMetaData) Less(i, j int) bool {
	return s.PlaylistVideoMetaDataArray[i].Position > s.PlaylistVideoMetaDataArray[j].Position
}
func (s PlaylistMetaData) Swap(i, j int) {
	s.PlaylistVideoMetaDataArray[i], s.PlaylistVideoMetaDataArray[j] = s.PlaylistVideoMetaDataArray[j], s.PlaylistVideoMetaDataArray[i]
}

func FlushFetchHistory(deliveries []Delivery) {
	var fetchHistory = resource.FetchHistory{Playlists: GenerateFetchHistory(deliveries)}
	log.Debugf("fetchHistory: %v", fetchHistory)

	resource.MarshalPlaylistJson(fetchHistory)
}

func ProcessOneVideo(delivery *Delivery) {
	if !delivery.Done {
		audioFile, err := fetchAudio(delivery)
		if err != nil {
			log.Warnf("Failed to download audio url %s from YouTube, error: %v", delivery.Parcel.Url, err)
			SendWarningMessage(util.FailedToFetchAudioWarningTemplate, err.Error())
			return
		}
		if v, template := IsAudioValid(audioFile); v == false {
			log.Warnf("Downloaded file from YouTube %s is NOT valid: %s", delivery.Parcel.Url, template)
			SendWarningMessage(template, audioFile.FilePath)
			return
		} else {
			delivery.Parcel = audioFile
			err = SendAudio(delivery)

			if err != nil {
				log.Warnf("Failed to send file %s to telegram channel, error: %v", audioFile.FilePath, err)
				audioFile.Caption = audioFile.Caption + fmt.Sprintf("%s", err)
				SendWarningMessage(util.FailedToSendAudioWarningTemplate, audioFile.Caption)
			} else {
				reporter.BriefSummary.SuccessfulFetch++
			}
		}
		if audioFile.FilePath != "" {
			log.Debugf("ready to clean up, audioFile: %v", audioFile)
			io2.Cleanup(audioFile.FilePath)
		}
	} else {
		log.Infof("the state of this delivery %v is DONE, no more process", delivery)
	}
	fmt.Println()
}

func fetchAudio(delivery *Delivery) (Parcel, error) {
	// download a video
	return DownloadYouTubeAudioToPath(delivery)
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

func GetPlaylistMetaDataBy(playlistId string) PlaylistMetaData {
	var playlistMetaData PlaylistMetaData
	svc, err := GetYouTubeService()
	if err != nil {
		log.Errorf("get youtube service error:%v", err)
		return playlistMetaData
	}

	playlistResponse := playlistItemsList(svc, util.YouTubePart, playlistId)

	playlistMetaData.PlaylistId = playlistId
	for _, playlistItem := range playlistResponse.Items {
		publishedAt := playlistItem.Snippet.PublishedAt
		title := playlistItem.Snippet.Title
		localPublishedAt := util.GetLocalDateTime(publishedAt)
		channelTitle := playlistItem.Snippet.ChannelTitle
		channelId := playlistItem.Snippet.ChannelId

		videoId := playlistItem.Snippet.ResourceId.VideoId
		position := playlistItem.Snippet.Position
		log.Debugf("%s(%s) from %s(%s) on position %v artist(%s/%s) was published at %s",
			title, videoId, channelTitle, channelId, position,
			util.GetYouTubePlaylistArtist(playlistId),
			util.GetYouTubePlaylistAlbum(playlistId),
			localPublishedAt)

		videoMetaData := PlaylistVideoMetaData{
			util.GetYouTubePlaylistArtist(playlistId),
			util.GetYouTubePlaylistAlbum(playlistId),
			videoId,
			util.MakeYouTubeRawUrl(videoId),
			position}
		playlistMetaData.PlaylistVideoMetaDataArray = append(playlistMetaData.PlaylistVideoMetaDataArray, &videoMetaData)
	}

	return playlistMetaData
}

func playlistItemsList(service *youtubeapi.Service, part []string, playlistId string) *youtubeapi.PlaylistItemListResponse {
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	call = call.MaxResults(util.GetYouTubePlaylistMaxResultsCount(playlistId))
	response, err := call.Do()
	if err != nil {
		log.Errorf("get playlist items error:%v, playlistId:%s", err, playlistId)
	}
	return response
}

func RetrieveITagOfMinimumSizeAudio(mediaUrl string) ([]int, error) {
	client := youtube.Client{}

	log.Debugf("Ready to get video: %s at %s", mediaUrl, time.Now().Format(util.DateTimeFormat))
	video, err := client.GetVideo(mediaUrl)
	log.Debugf("video duration: %vs at %s", video.Duration.Seconds(), time.Now().Format(util.DateTimeFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to get video, error:%s, mediaUrl:%s", err, mediaUrl)
	}
	var videoMetaDataArray []VideoMetaData
	for _, f := range video.Formats {
		log.Debugf("ItagNo:%v, MimeType:%s, ADM:%s, FPS:%v, QL:%s, AQ:%s, AC:%v, AverBit:%v, Bit:%v, Size:%v",
			f.ItagNo, f.MimeType, f.ApproxDurationMs, f.FPS, f.QualityLabel, f.AudioQuality, f.AudioChannels, f.AverageBitrate, f.Bitrate, f.ContentLength)
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
	if len(videoMetaDataArray) == 0 {
		return nil, fmt.Errorf("proper audio track not found:%s", mediaUrl)
	}

	// Sort videoMetaDataArray by ContentLength in ascending order
	sort.Slice(videoMetaDataArray, func(i, j int) bool {
		return videoMetaDataArray[i].ContentLength < videoMetaDataArray[j].ContentLength
	})

	// Extract iTagNo values from sorted videoMetaDataArray
	var iTagNos []int
	for _, v := range videoMetaDataArray {
		iTagNos = append(iTagNos, v.ITagNo)
	}

	return iTagNos, nil
}

func DownloadYouTubeAudioToPath(delivery *Delivery) (Parcel, error) {
	var parcel Parcel
	log.Debugf("Ready to download media %s(playlistId: %s) at %s", delivery.Parcel.Url, delivery.PlaylistId, time.Now().Format(util.DateTimeFormat))
	result, err := goutubedl.New(context.Background(), delivery.Parcel.Url, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %v, url: %s", err, delivery.Parcel.Url)
	}

	fileExtension := getFileExtension(result.Info.ACodec)
	validMediaFileName := util.FilenamifyMediaTitle(result.Info.Title + fileExtension)
	parcelFilePath := fmt.Sprintf("%s%s", util.GetYouTubeFetchBase().DownloadedFilesPath, validMediaFileName)
	parcel = GenerateParcel(parcelFilePath,
		result.Info.Title,
		util.GetYouTubePlaylistArtist(delivery.PlaylistId),
		util.GetYouTubePlaylistAlbum(delivery.PlaylistId),
		delivery.Parcel.Url)
	log.Debugf("ext: %s, generated parcel: %v", fileExtension, parcel)

	log.Debugf("ready to CREATE media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	parcelFile, err := os.Create(parcel.FilePath)
	log.Debugf("media file %s CREATED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		log.Errorf("creating file error: %v", err)
	}

	// Retrieve the list of iTagNos
	iTagNos, err := RetrieveITagOfMinimumSizeAudio(delivery.Parcel.Url)
	if err != nil {
		log.Errorf("retrieve iTag error, error: %s", err)
		return parcel, fmt.Errorf("retrieve iTag error: %v, url: %s", err, delivery.Parcel.Url)
	}

	var downloadedResult *goutubedl.DownloadResult
	// Polling to download audio until no error occurs
	for _, iTagNo := range iTagNos {
		downloadedResult, err = result.Download(context.Background(), strconv.Itoa(iTagNo))
		if err == nil {
			log.Debugf("iTagNo is %v at %s", iTagNo, time.Now().Format(util.DateTimeFormat))
			break
		}
		log.Errorf("download error with iTagNo %v: %s", iTagNo, err)
	}

	if err != nil {
		log.Errorf("final download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %v, url: %s", err, delivery.Parcel.Url)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)
	log.Debugf("downloading media %s at %s", result.Info.Title, time.Now().Format(util.DateTimeFormat))

	log.Debugf("ready to COPY media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	written, err := io.Copy(parcelFile, downloadedResult)
	log.Infof("media file %s DOWNLOADED & COPIED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		return parcel, fmt.Errorf("copy error: %s, parcel: %v, written: %v", err, parcel, written)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(parcelFile)

	log.Printf("Title: %s, Artist: %s, Album: %s, Url: %s", parcel.Caption, parcel.Artist, parcel.Album, parcel.Url)
	parcel, _ = convertToMp3AndFillMetadata(parcel)

	return parcel, nil
}
func convertToMp3AndFillMetadata(parcel Parcel) (Parcel, error) {
	// 生成新的文件路径，使用.mp3作为扩展名
	newFilePath := strings.TrimSuffix(parcel.FilePath, filepath.Ext(parcel.FilePath)) + ".mp3"
	// 构建ffmpeg命令
	cmd := exec.Command("ffmpeg", "-i", parcel.FilePath,
		"-codec:a", "libmp3lame", "-qscale:a", QualityScaleFrom0To9,
		"-metadata", "artist="+parcel.Artist,
		"-metadata", "title="+util.FilenamifyMediaTitle(parcel.Caption),
		"-metadata", "album="+parcel.Album,
		newFilePath)
	//cmd := exec.Command("ffmpeg", "-i", parcel.FilePath,
	//	"-codec:a", "aac", "-b:a", "64k",
	//	"-metadata", "artist="+parcel.Artist,
	//	"-metadata", "title="+util.FilenamifyMediaTitle(parcel.Caption),
	//	"-metadata", "album="+parcel.Album,
	//	newFilePath)

	fullCommand := strings.Join(cmd.Args, " ")
	log.Printf("Executing command: %s", fullCommand)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("ffmpeg error: %s, stderr: %s", err, stderr.String())
		return parcel, fmt.Errorf("ffmpeg error: %s, command: %s", err, fullCommand)
	}

	ffprobeCmd := exec.Command("ffprobe", "-show_format", newFilePath)
	output, err := ffprobeCmd.CombinedOutput()
	if err != nil {
		log.Errorf("ffprobe error: %s", err)
		return parcel, fmt.Errorf("ffprobe error: %s", err)
	}
	log.Printf("ffprobe output: %s", string(output))
	time.Sleep(5 * time.Second)

	parcel.FilePath = newFilePath
	log.Printf("ffmpeg command executed successfully, new file: %s", parcel.FilePath)

	return parcel, nil
}
func getFileExtension(mimeType string) string {
	// 简单的MimeType到文件扩展名的映射
	switch {
	case strings.Contains(mimeType, "audio/mp4"):
		return ".m4a"
	case strings.Contains(mimeType, "audio/webm"):
		return ".webm"
	case strings.Contains(mimeType, "audio/ogg"):
		return ".ogg"
	default:
		return ".mp4" // 默认情况下使用.mp4，适用于不确定的情况
	}
}

func GenerateFetchHistory(deliveries []Delivery) []resource.HistoryProps {
	channelChatId, _ := env.GetEnvVariable(util.EnvChatIdName)
	subscriberId, _ := strconv.ParseInt(channelChatId, 10, 64)

	var playlistMap map[string][]resource.SubscriberItems
	playlistMap = make(map[string][]resource.SubscriberItems)
	for _, delivery := range deliveries {
		log.Debugf("delivery to be generated: %v", delivery)
		subscribers, ok := playlistMap[delivery.PlaylistId]
		if ok {
			log.Infof("playlist id %s FOUND from playlistMap, length: %v", delivery.PlaylistId, len(playlistMap))
			var newSubscribers []resource.SubscriberItems
			if delivery.Done {
				for _, sub := range subscribers {
					if sub.Id == subscriberId {
						sub.LastFetch = resource.FetchItems{Datetime: sub.LastFetch.Datetime, Timestamp: sub.LastFetch.Timestamp, Urls: append(sub.LastFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			} else {
				log.Debugf("delivery.done FALSE: %v", delivery)
				now := time.Now()
				for _, sub := range subscribers {
					log.Debugf("sub: %v", sub)
					nextFetchTimestamp := delivery.Timestamp
					nextFetchDatetime := delivery.Datetime
					if nextFetchTimestamp == 0 {
						nextFetchTimestamp = now.Unix()
						nextFetchDatetime = now.Format(util.DateTimeFormat)
					}
					if sub.Id == subscriberId {
						sub.NextFetch = resource.FetchItems{Datetime: nextFetchDatetime, Timestamp: nextFetchTimestamp, Urls: append(sub.NextFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			}
			log.Debugf("newSubscribers: %v", newSubscribers)
			playlistMap[delivery.PlaylistId] = newSubscribers
		} else {
			log.Infof("playlist id %s NOT FOUND from playlistMap: %v", delivery.PlaylistId, playlistMap)
			var urls []string
			urls = append(urls, delivery.Parcel.Url)
			var lastFetch, nextFetch resource.FetchItems
			thisFetch := resource.FetchItems{Datetime: delivery.Datetime, Timestamp: delivery.Timestamp, Urls: urls}
			if delivery.Done {
				lastFetch = thisFetch
				log.Debugf("lastFetch: %v", lastFetch)
			} else {
				nextFetch = thisFetch
				log.Debugf("nextFetch: %v", nextFetch)
			}
			subscriberItem := resource.SubscriberItems{Id: subscriberId, LastFetch: lastFetch, NextFetch: nextFetch}
			subscriberItems := []resource.SubscriberItems{subscriberItem}
			playlistMap[delivery.PlaylistId] = subscriberItems
		}
	}
	log.Debugf("playlistMap: %v", playlistMap)
	var historyPropsArray []resource.HistoryProps
	for playlistId := range playlistMap {
		historyProps := resource.HistoryProps{Id: playlistId, Subscribers: playlistMap[playlistId]}
		historyPropsArray = append(historyPropsArray, historyProps)
	}
	return historyPropsArray
}

func MergeHistoryFetchesInto(newDeliveries []Delivery) []Delivery {
	historyProps := resource.MediaHistory
	log.Debugf("newDeliveries count: %v, historyProps count: %v", len(newDeliveries), len(historyProps))
	var mergedDeliveries []Delivery
	for _, newDelivery := range newDeliveries {
		log.Debugf("newDelivery: %v", newDelivery)
		isNewPlayListId := true
		for _, historyProp := range historyProps {
			if newDelivery.PlaylistId == historyProp.Id {
				isNewPlayListId = false
				for _, sub := range historyProp.Subscribers {
					AppendDeliveries(&mergedDeliveries, sub.LastFetch, historyProp.Id, true)
					nextFetchUrls := sub.NextFetch.Urls
					if len(nextFetchUrls) > 0 {
						if util.StringSliceContains(nextFetchUrls, newDelivery.Parcel.Url) {
							log.Infof("newDelivery url %s was FOUND from history NEXT fetch urls: %v", newDelivery.Parcel.Url, nextFetchUrls)
							AppendDeliveries(&mergedDeliveries, sub.NextFetch, historyProp.Id, false)
						} else {
							log.Infof("newDelivery url %s NOT FOUND from history NEXT fetch urls: %v, add it", newDelivery.Parcel.Url, nextFetchUrls)
							mergedDeliveries = append(mergedDeliveries, newDelivery)
						}
					} else {
						log.Infof("next fetch urls EMPTY, subscribers id: %v, playlist id: %v, add it", sub.Id, historyProp.Id)
						mergedDeliveries = append(mergedDeliveries, newDelivery)
					}
				}
				break
			}
		}
		if isNewPlayListId {
			now := time.Now()
			newDelivery.Timestamp = now.Unix()
			newDelivery.Datetime = now.Format(util.DateTimeFormat)
			mergedDeliveries = append(mergedDeliveries, newDelivery)
		}
	}
	mergedDeliveriesWithoutDuplicated := RemoveDuplicatedUrlsByLoop(mergedDeliveries)
	log.Infof("merged deliveries count: %v which removed duplicated items", len(mergedDeliveriesWithoutDuplicated))
	return mergedDeliveriesWithoutDuplicated
}

func GetYouTubeVideosFromPlaylistId(playlistId string) []PlaylistMetaData {
	return []PlaylistMetaData{GetPlaylistMetaDataBy(playlistId)}
}

func GetYouTubeVideosFromPlaylists() []PlaylistMetaData {
	var playlistMetaDataArray []PlaylistMetaData
	for _, param := range util.GetYouTubeFetchBase().Params {
		playlistMetaData := GetPlaylistMetaDataBy(param.Id)
		if param.SortByPosition {
			log.Debugf("SORT the playlist:%s", param.Id)
			sort.Sort(playlistMetaData)
		}
		playlistMetaDataArray = append(playlistMetaDataArray, playlistMetaData)
	}
	return playlistMetaDataArray
}

func AssembleDeliveryFromSingleUrl(url string) Delivery {
	parcel := Parcel{Url: url}
	return Delivery{Parcel: parcel}
}

func AssembleDeliveriesFromPlaylists(playlistMetaDataArray []PlaylistMetaData) []Delivery {
	var deliveries []Delivery
	for _, playlistMetaData := range playlistMetaDataArray {
		delivery := Delivery{}
		delivery.PlaylistId = playlistMetaData.PlaylistId
		for _, playlistVideoMetaData := range playlistMetaData.PlaylistVideoMetaDataArray {
			delivery.Parcel = GenerateParcel("", "", playlistVideoMetaData.Artist, playlistVideoMetaData.Album, playlistVideoMetaData.RawUrl)
			deliveries = append(deliveries, delivery)
		}
	}
	log.Infof("total incoming playlists: %v, total incoming deliveries: %v", len(playlistMetaDataArray), len(deliveries))
	return deliveries
}

func GenerateYouTubeCredentials() (YouTubeCredentials, error) {
	var err error
	var youTubeCredentials YouTubeCredentials

	youtubeKey, err := env.GetEnvVariable(util.EnvYouTubeKeyName)
	if err != nil {
		log.Errorf("%s", err)
		return youTubeCredentials, fmt.Errorf("reading env %s vars error", util.EnvYouTubeKeyName)
	}
	youTubeCredentials.Key = youtubeKey

	return youTubeCredentials, nil
}
