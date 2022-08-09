package handler

import (
	"errors"
	"fmt"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
	"strconv"
	"time"
	"youtube-audio/pkg/util"
)

const (
	DateTimeFormat                        string = "2006-01-02 15:04:05"
	YouTubeDateTimeFormat                 string = "2006-01-02T15:04:05Z"
	EnvTokenName                          string = "BOT_TOKEN"
	EnvChatIdName                         string = "CHAT_ID"
	EnvBotChatIdName                      string = "BOT_CHAT_ID"
	EnvYouTubeKeyName                     string = "YOUTUBE_KEY"
	IllegalCharacterReplacementInFilename string = "_"
	FilenameMaxLength                     int    = 512
	UploadAudioMaxLength                  int64  = 52428800
	YouTubeDefaultMaxResults              int64  = 1
	FetchYouTubeMaxResultsLimit           int64  = 50
	FetchMaxUrlsLimit                     int    = 10
	FailedToSendAudioWarningTemplate      string = "FAILED TO SEND AUDIO %s TO THE CHANNEL"
	FailedToDownloadAudioWarningTemplate  string = "FAILED TO DOWNLOAD AUDIO %s"
	InvalidDownloadedAudioWarningTemplate string = "INVALID DOWNLOADED FILE %s"
)

var YouTubePart = []string{"snippet"}

// RemoveDuplicatedUrlsByLoop 通过两重循环过滤重复元素 ref: https://blog.csdn.net/qq_27068845/article/details/77407358
func RemoveDuplicatedUrlsByLoop(slc []Delivery) []Delivery {
	var result []Delivery
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i].Parcel.Url == result[j].Parcel.Url {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

func appendDeliveries(deliveries *[]Delivery, fetchItems util.FetchItems, playlistId string, done bool) {
	// remain time from FetchItems
	fetchTimestamp := fetchItems.Timestamp
	fetchDatetime := fetchItems.Datetime
	if done {
		// apply current time to last_fetch block
		now := time.Now()
		fetchTimestamp = now.Unix()
		fetchDatetime = now.Format(DateTimeFormat)
	} else {
		if len(fetchItems.Urls) > 0 {
			fetchTime := time.Unix(fetchTimestamp, 0)
			durationTillNow := time.Since(fetchTime)
			log.Infof("fetch time till now hours: %v", durationTillNow.Hours())
			if durationTillNow.Hours() > 48 {
				log.Warnf("fetch block time has EXPIRED: %s, playlistId: %s, urls: %v, drop it", fetchDatetime, playlistId, fetchItems.Urls)
				return
			}
		} else {
			log.Infof("EMPTY fetch items urls, playlistId: %s, urls: %v, ignore it", playlistId, fetchItems.Urls)
			return
		}
	}
	// always keep the fetch block, but under maximum count of urls
	if len(fetchItems.Urls) > FetchMaxUrlsLimit {
		removeUrlsCount := len(fetchItems.Urls) - FetchMaxUrlsLimit
		fetchItems.Urls = fetchItems.Urls[removeUrlsCount:]
	}
	for _, fetchUrl := range fetchItems.Urls {
		historyFetch := Delivery{
			Parcel:     GenerateParcel("", "", fetchUrl),
			PlaylistId: playlistId,
			Done:       done,
			Timestamp:  fetchTimestamp,
			Datetime:   fetchDatetime,
		}
		*deliveries = append(*deliveries, historyFetch)
	}
}

func DeleteSliceElms(sl []util.HistoryProps, elms ...util.HistoryProps) []util.HistoryProps {
	if len(sl) == 0 || len(elms) == 0 {
		return sl
	}
	// 先将元素转为 set
	m := make(map[string]util.HistoryProps)
	for _, v := range elms {
		m[v.Id] = util.HistoryProps{}
	}
	// 过滤掉指定元素
	res := make([]util.HistoryProps, 0, len(sl))
	for _, v := range sl {
		if _, ok := m[v.Id]; !ok {
			res = append(res, v)
		}
	}
	return res
}

func MergeHistoryFetchesInto(newDeliveries []Delivery) []Delivery {
	historyProps := util.MediaHistory
	log.Infof("newDeliveries count: %v, historyProps count: %v", len(newDeliveries), len(historyProps))
	var mergedDeliveries []Delivery
	for _, newDelivery := range newDeliveries {
		log.Infof("newDelivery: %v", newDelivery)
		isNewPlayListId := true
		for _, historyProp := range historyProps {
			if newDelivery.PlaylistId == historyProp.Id {
				isNewPlayListId = false
				for _, sub := range historyProp.Subscribers {
					searchNextFetchUrls := false
					lastFetchUrls := sub.LastFetch.Urls
					if len(lastFetchUrls) > 0 {
						sort.Strings(lastFetchUrls)
						lastFetchUrlsIndex := sort.SearchStrings(lastFetchUrls, newDelivery.Parcel.Url)
						log.Infof("lastFetchUrlsIndex: %v, last fetch urls count: %v", lastFetchUrlsIndex, len(lastFetchUrls))
						if lastFetchUrlsIndex < len(lastFetchUrls) && lastFetchUrls[lastFetchUrlsIndex] == newDelivery.Parcel.Url {
							// will cause append same next fetch urls repeatedly
							log.Infof("newDelivery url %s was FOUND from history LAST fetch urls, drop it and add all next fetch urls", newDelivery.Parcel.Url)
							appendDeliveries(&mergedDeliveries, sub.NextFetch, historyProp.Id, false)
						} else {
							log.Infof("newDelivery url %s NOT FOUND from history LAST fetch urls", newDelivery.Parcel.Url)
							searchNextFetchUrls = true
						}
						appendDeliveries(&mergedDeliveries, sub.LastFetch, historyProp.Id, true)
					} else {
						log.Infof("last fetch urls EMPTY, subscribers id: %v, playlist id: %v", sub.Id, historyProp.Id)
						searchNextFetchUrls = true
					}

					if searchNextFetchUrls {
						nextFetchUrls := sub.NextFetch.Urls
						if len(nextFetchUrls) > 0 {
							sort.Strings(nextFetchUrls)
							nextFetchUrlsIndex := sort.SearchStrings(nextFetchUrls, newDelivery.Parcel.Url)
							log.Infof("nextFetchUrlsIndex: %v, next fetch urls count: %v", nextFetchUrlsIndex, len(nextFetchUrls))
							if nextFetchUrlsIndex < len(nextFetchUrls) && nextFetchUrls[nextFetchUrlsIndex] == newDelivery.Parcel.Url {
								log.Infof("newDelivery url %s was FOUND from history NEXT fetch urls", newDelivery.Parcel.Url)
								appendDeliveries(&mergedDeliveries, sub.NextFetch, historyProp.Id, false)
							} else {
								log.Infof("newDelivery url %s NOT FOUND from history NEXT fetch urls, add it", newDelivery.Parcel.Url)
								mergedDeliveries = append(mergedDeliveries, newDelivery)
							}
						} else {
							log.Infof("next fetch urls EMPTY, subscribers id: %v, playlist id: %v, add it", sub.Id, historyProp.Id)
							mergedDeliveries = append(mergedDeliveries, newDelivery)
						}
					}
				}
				break
			}
		}
		if isNewPlayListId {
			now := time.Now()
			newDelivery.Timestamp = now.Unix()
			newDelivery.Datetime = now.Format(DateTimeFormat)
			mergedDeliveries = append(mergedDeliveries, newDelivery)
		}
	}
	mergedDeliveriesWithoutDuplicated := RemoveDuplicatedUrlsByLoop(mergedDeliveries)
	log.Infof("merged deliveries count: %v which removed duplicated items", len(mergedDeliveriesWithoutDuplicated))
	return mergedDeliveriesWithoutDuplicated
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

func GenerateFetchHistory(deliveries []Delivery) []util.HistoryProps {
	channelChatId, _ := util.GetEnvVariable(EnvChatIdName)
	subscriberId, _ := strconv.ParseInt(channelChatId, 10, 64)

	var playlistMap map[string][]util.SubscriberItems
	playlistMap = make(map[string][]util.SubscriberItems)
	for _, delivery := range deliveries {
		subscribers, ok := playlistMap[delivery.PlaylistId]
		if ok {
			log.Infof("playlist id %s FOUND from playlistMap: %v", delivery.PlaylistId, playlistMap)
			var newSubscribers []util.SubscriberItems
			if delivery.Done {
				for _, sub := range subscribers {
					if sub.Id == subscriberId {
						sub.LastFetch = util.FetchItems{Datetime: sub.LastFetch.Datetime, Timestamp: sub.LastFetch.Timestamp, Urls: append(sub.LastFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			} else {
				now := time.Now()
				for _, sub := range subscribers {
					nextFetchTimestamp := sub.NextFetch.Timestamp
					nextFetchDatetime := sub.NextFetch.Datetime
					if nextFetchTimestamp == 0 {
						nextFetchTimestamp = now.Unix()
						nextFetchDatetime = now.Format(DateTimeFormat)
					}
					if sub.Id == subscriberId {
						sub.NextFetch = util.FetchItems{Datetime: nextFetchDatetime, Timestamp: nextFetchTimestamp, Urls: append(sub.NextFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			}
			log.Infof("newSubscribers: %v", newSubscribers)
			playlistMap[delivery.PlaylistId] = newSubscribers
		} else {
			log.Infof("playlist id %s NOT FOUND from playlistMap: %v", delivery.PlaylistId, playlistMap)
			var urls []string
			urls = append(urls, delivery.Parcel.Url)
			var lastFetch, nextFetch util.FetchItems
			thisFetch := util.FetchItems{Datetime: delivery.Datetime, Timestamp: delivery.Timestamp, Urls: urls}
			if delivery.Done {
				lastFetch = thisFetch
			} else {
				nextFetch = thisFetch
			}
			subscriberItem := util.SubscriberItems{Id: subscriberId, LastFetch: lastFetch, NextFetch: nextFetch}
			subscriberItems := []util.SubscriberItems{subscriberItem}
			playlistMap[delivery.PlaylistId] = subscriberItems
		}
	}
	log.Infof("playlistMap: %v", playlistMap)
	var historyPropsArray []util.HistoryProps
	for playlistId := range playlistMap {
		historyProps := util.HistoryProps{Id: playlistId, Subscribers: playlistMap[playlistId]}
		historyPropsArray = append(historyPropsArray, historyProps)
	}
	return historyPropsArray
}

func GetYouTubeVideosFromPlaylists() []PlaylistMetaData {
	var playlistMetaDataArray []PlaylistMetaData
	for _, param := range GetYouTubeFetchBase().Params {
		playlistMetaData := GetPlaylistMetaDataBy(param.Id)
		if param.SortByPosition {
			log.Infof("SORT the playlist:%s", param.Id)
			sort.Sort(playlistMetaData)
		}
		playlistMetaDataArray = append(playlistMetaDataArray, playlistMetaData)
	}
	return playlistMetaDataArray
}

func GetYouTubeFetchBase() util.BaseProps {
	youtube := util.BaseProps{}
	for _, cp := range util.MediaBase {
		if cp.Owner == "YouTube" {
			youtube = cp
			break
		}
		continue
	}
	if youtube.Owner == "" {
		log.Fatalf("getting youtube playlists from json error, util.MediaBase:%v", util.MediaBase)
	}
	return youtube
}

func GetYouTubePlaylistMaxResultsCount(playlistId string) int64 {
	youtube := GetYouTubeFetchBase()
	for _, scope := range youtube.Params {
		if scope.Id == playlistId {
			return scope.MaxResultsCount
		}
	}
	if youtube.Owner == "" {
		log.Fatalf("getting youtube playlist max results count from json error, scopes:%v", youtube.Params)
	}
	return 1
}

func GetLocalDateTime(formattedDateTime string) string {
	youtubeTime, err := time.Parse(YouTubeDateTimeFormat, formattedDateTime)
	if err != nil {
		log.Errorf("parse youtube datetime error:%s, formattedDateTime: %s with format: %s", err, formattedDateTime, YouTubeDateTimeFormat)
		return formattedDateTime
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return youtubeTime.In(loc).Format(DateTimeFormat)
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func MakeYouTubeRawUrl(videoId string) string {
	return GetYouTubeFetchBase().PrefixUrl + videoId
}

func FilenamifyMediaTitle(title string) (string, error) {
	rawMediaTitle := fmt.Sprintf("%s%s", title, GetYouTubeFetchBase().MediaExtension)
	log.Infof("rawMediaTitle %s", rawMediaTitle)
	validMediaFileName, err := filenamify.Filenamify(rawMediaTitle, filenamify.Options{
		Replacement: IllegalCharacterReplacementInFilename,
		MaxLength:   FilenameMaxLength,
	})
	if err != nil {
		log.Errorf("convert raw media title to a valid file name error:%s", err)
		return "", fmt.Errorf("filenamify error: %s", rawMediaTitle)
	}
	log.Infof("validMediaFileName %s", validMediaFileName)

	return validMediaFileName, nil
}

func GenerateParcel(filePath string, caption string, url string) Parcel {
	parcel := Parcel{
		FilePath: filePath,
		Caption:  caption,
		Url:      url,
	}
	return parcel
}

func GenerateYouTubeCredentials() (YouTubeCredentials, error) {
	var err error
	var youTubeCredentials YouTubeCredentials

	youtubeKey, err := util.GetEnvVariable(EnvYouTubeKeyName)
	if err != nil {
		log.Errorf("%s", err)
		return youTubeCredentials, fmt.Errorf("reading env %s vars error", EnvYouTubeKeyName)
	}
	youTubeCredentials.Key = youtubeKey

	return youTubeCredentials, nil
}

func GenerateTelegramBot() (TelegramBot, error) {
	var err error
	var telegramBot TelegramBot

	// Get the TOKEN and the CHAT_ID
	botToken, err := util.GetEnvVariable(EnvTokenName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvTokenName)
	}
	telegramBot.Token = botToken

	channelChatId, err := util.GetEnvVariable(EnvChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvChatIdName)
	}
	telegramBot.ChannelChatId = channelChatId

	botChatId, err := util.GetEnvVariable(EnvBotChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvBotChatIdName)
	}
	intBotChatId, _ := strconv.ParseInt(botChatId, 10, 64)
	telegramBot.BotChatId = intBotChatId

	return telegramBot, nil
}
