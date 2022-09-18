package handler

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/wader/goutubedl"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
	io2 "youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

//ITagNo format id
//  for _, f := range result.Formats() {
//		log.Debugf("format id:%s", f.FormatID)
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

type PlaylistVideoMetaData struct {
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
		audioFile, err := fetchAudio(delivery.Parcel.Url)
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

func GetPlaylistMetaDataBy(playlistId string) PlaylistMetaData {
	var playlistMetaData PlaylistMetaData
	svc, err := GetYouTubeService()
	if err != nil {
		log.Errorf("get youtube service error:%v", err)
		return playlistMetaData
	}

	playlistResponse := playlistItemsList(svc, util.YouTubePart, playlistId, util.GetYouTubePlaylistMaxResultsCount(playlistId))

	playlistMetaData.PlaylistId = playlistId
	for _, playlistItem := range playlistResponse.Items {
		publishedAt := playlistItem.Snippet.PublishedAt
		title := playlistItem.Snippet.Title
		localPublishedAt := util.GetLocalDateTime(publishedAt)
		channelTitle := playlistItem.Snippet.ChannelTitle
		channelId := playlistItem.Snippet.ChannelId

		videoId := playlistItem.Snippet.ResourceId.VideoId
		position := playlistItem.Snippet.Position
		log.Debugf("%s(%s) from %s(%s) on position %v was published at %s", title, videoId, channelTitle, channelId, position, localPublishedAt)

		videoMetaData := PlaylistVideoMetaData{videoId, util.MakeYouTubeRawUrl(videoId), position}
		playlistMetaData.PlaylistVideoMetaDataArray = append(playlistMetaData.PlaylistVideoMetaDataArray, &videoMetaData)
	}

	return playlistMetaData
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
	if maxResults <= 0 || maxResults > util.FetchYouTubeMaxResultsLimit {
		log.Errorf("illegal maxResults error:%v", maxResults)
		maxResults = util.YouTubeDefaultMaxResults
	}
	call = call.MaxResults(maxResults)
	//lastUpdated := queryParamOpt{key: "order", value: "time"}
	response, err := call.Do()
	if err != nil {
		log.Errorf("get playlist items error:%v, playlistId:%s", err, playlistId)
	}
	return response
}

func RetrieveITagOfMinimumSizeAudio(mediaUrl string) (int, error) {
	client := youtube.Client{}

	log.Debugf("Ready to get video: %s at %s", mediaUrl, time.Now().Format(util.DateTimeFormat))
	video, err := client.GetVideo(mediaUrl)
	log.Debugf("video duration: %vs at %s", video.Duration.Seconds(), time.Now().Format(util.DateTimeFormat))
	if err != nil {
		return -1, fmt.Errorf("failed to get video, error:%s, mediaUrl:%s", err, mediaUrl)
	}
	var videoMetaDataArray []VideoMetaData
	for _, f := range video.Formats {
		log.Debugf("ItagNo:%v, ADM:%s, FPS:%v, QL:%s, AQ:%s, AC:%v, AverBit:%v, Bit:%v, Size:%v",
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
		ContentLength:    util.UploadAudioMaxLength,
		ApproxDurationMs: "",
		AudioChannels:    0,
	}
	if len(videoMetaDataArray) == 0 {
		return -1, fmt.Errorf("proper audio track not found:%s", mediaUrl)
	} else if len(videoMetaDataArray) == 1 {
		log.Debugf("Found only 1 proper audio track for %s, iTagNo: %v at %s", mediaUrl, videoMetaDataArray[0].ITagNo, time.Now().Format(util.DateTimeFormat))
		minSizeVideoMetaData = videoMetaDataArray[0]
	} else {
		log.Debugf("Found %v proper audio tracks for %s, at %s", len(videoMetaDataArray), mediaUrl, time.Now().Format(util.DateTimeFormat))
		for _, v := range videoMetaDataArray {
			if v.ContentLength < minSizeVideoMetaData.ContentLength {
				minSizeVideoMetaData = v
			} else {
				maxSizeVideoMetaData = v
			}
		}
		log.Debugf("maxSizeVideoMetaData: %v", maxSizeVideoMetaData)
	}
	log.Debugf("minSizeVideoMetaData: %v", minSizeVideoMetaData)
	if util.UploadAudioMaxLength < minSizeVideoMetaData.ContentLength {
		return -1, fmt.Errorf("the min size %v of audio track EXCEEDS the max %v",
			minSizeVideoMetaData.ContentLength, util.UploadAudioMaxLength)
	}
	return minSizeVideoMetaData.ITagNo, nil
}

func DownloadYouTubeAudioToPath(mediaUrl string) (Parcel, error) {
	var parcel Parcel
	log.Debugf("Ready to download media %s at %s", mediaUrl, time.Now().Format(util.DateTimeFormat))
	result, err := goutubedl.New(context.Background(), mediaUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %v, url: %s", err, mediaUrl)
	}

	validMediaFileName, err := util.FilenamifyMediaTitle(result.Info.Title)
	if err != nil {
		return parcel, err
	}
	parcel = GenerateParcel(fmt.Sprintf("%s%s", util.GetYouTubeFetchBase().DownloadedFilesPath, validMediaFileName), result.Info.Title, mediaUrl)
	log.Debugf("generated parcel: %v", parcel)

	log.Debugf("ready to CREATE media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	parcelFile, err := os.Create(parcel.FilePath)
	log.Debugf("media file %s CREATED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		log.Fatalf("creating file error: %v", err)
	}

	iTagNo, err := RetrieveITagOfMinimumSizeAudio(mediaUrl)
	if err != nil {
		log.Errorf("retrieve iTag error, iTagNo: %v, error: %s", iTagNo, err)
		return parcel, fmt.Errorf("goutubedl download error: %v, url: %s, ITagNo: %v", err, mediaUrl, iTagNo)
	}
	downloadedResult, err := result.Download(context.Background(), strconv.Itoa(iTagNo))
	if err != nil {
		log.Errorf("download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %v, url: %s, ITagNo: %v", err, mediaUrl, iTagNo)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)
	log.Debugf("downloading media %s from %s", result.Info.Title, time.Now().Format(util.DateTimeFormat))

	log.Debugf("ready to COPY media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	written, err := io.Copy(parcelFile, downloadedResult)
	log.Infof("media file %s DOWNLOADED & COPIED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		return parcel, fmt.Errorf("copy error: %s, parcel: %v, written: %v", err, parcel, written)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(parcelFile)

	return parcel, nil
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

func AssembleDeliveriesFromPlaylists() []Delivery {
	playlistMetaDataArray := GetYouTubeVideosFromPlaylists()
	var deliveries []Delivery
	for _, playlistMetaData := range playlistMetaDataArray {
		delivery := Delivery{}
		delivery.PlaylistId = playlistMetaData.PlaylistId
		for _, playlistVideoMetaData := range playlistMetaData.PlaylistVideoMetaDataArray {
			delivery.Parcel = GenerateParcel("", "", playlistVideoMetaData.RawUrl)
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
