package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	FetchBaseJsonPath        string = "/Users/fangdajiang/IdeaProjects/youtube-audio/resource/fetch_base.json"
	FetchHistoryJsonPath     string = "/Users/fangdajiang/IdeaProjects/youtube-audio/resource/fetch_history.json"
	TempFetchHistoryJsonPath string = "/Users/fangdajiang/IdeaProjects/youtube-audio/resource/tmp_fetch_history.json"
)

var MediaBase []BaseProps
var MediaHistory []HistoryProps

type FetchItems struct {
	Datetime  string   `json:"datetime"`
	Timestamp int64    `json:"timestamp"`
	Urls      []string `json:"urls"`
}

type SubscriberItems struct {
	Id        int64      `json:"id"`
	LastFetch FetchItems `json:"last_fetch"`
	NextFetch FetchItems `json:"next_fetch"`
}

type HistoryProps struct {
	Id          string            `json:"id"`
	Subscribers []SubscriberItems `json:"subscribers"`
}

type ParamItems struct {
	Id              string `json:"id"`
	MaxResultsCount int64  `json:"max_results_count"`
	SortByPosition  bool   `json:"sort_by_position"`
}

type BaseProps struct {
	Owner               string `json:"owner"`
	PrefixUrl           string `json:"prefix_url"`
	MediaExtension      string `json:"media_extension"`
	DownloadedFilesPath string `json:"downloaded_files_path"`
	Params              []ParamItems
}

type FetchBase struct {
	Playlists []BaseProps `json:"playlists"`
}

type FetchHistory struct {
	Playlists []HistoryProps `json:"playlists"`
}

func CollectValidNextFetchUrls() []string {
	_, nextFetches := CollectFetches()
	var nextFetchUrls []string
	for _, nextFetch := range nextFetches {
		nextFetchTime := time.Unix(nextFetch.Timestamp, 0)
		durationTillNow := time.Since(nextFetchTime)
		log.Infof("next_fetch till now hours: %v", durationTillNow.Hours())
		if durationTillNow.Hours() > 8 {
			log.Warnf("next_fetch time has expired: %s, urls: %v", nextFetch.Datetime, nextFetch.Urls)
		} else {
			nextFetchUrls = append(nextFetchUrls, nextFetch.Urls...)
		}
	}
	return nextFetchUrls
}

func CollectFetches() ([]FetchItems, []FetchItems) {
	fetchHistory := FetchHistory{}
	fetchHistory.DecodePlaylistJson(FetchHistoryJsonPath)
	var lastFetchItems []FetchItems
	var nextFetchItems []FetchItems
	for _, historyProp := range fetchHistory.Playlists {
		for _, subscriberItems := range historyProp.Subscribers {
			if subscriberItems.LastFetch.Timestamp != 0 {
				lastFetchItems = append(lastFetchItems, subscriberItems.LastFetch)
			} else {
				log.Infof("empty last_fetch: %v", subscriberItems.LastFetch)
			}
			if subscriberItems.NextFetch.Timestamp != 0 {
				nextFetchItems = append(nextFetchItems, subscriberItems.NextFetch)
			} else {
				log.Infof("empty next_fetch: %v", subscriberItems.NextFetch)
			}
		}
	}
	return lastFetchItems, nextFetchItems
}

func EncodePlaylistJson(jsonPath string, fetchHistory FetchHistory) {
	resourceJson, _ := os.Create(jsonPath)
	defer func(file *os.File) {
		_ = file.Close()
	}(resourceJson)
	resourceEncoder := json.NewEncoder(resourceJson)
	resourceEncoder.SetIndent("", "  ")
	err := resourceEncoder.Encode(fetchHistory)
	if err != nil {
		log.Fatalf("encoding json error:%v, json:%s, fetchHistory:%v", err, jsonPath, fetchHistory)
	}
}

func (fh *FetchHistory) DecodePlaylistJson(jsonPath string) {
	resourceJson, _ := os.Open(jsonPath)
	defer func(file *os.File) {
		_ = file.Close()
	}(resourceJson)
	resourceDecoder := json.NewDecoder(resourceJson)
	err := resourceDecoder.Decode(&fh)
	if err != nil {
		log.Fatalf("decoding json error:%v, json path:%s", err, jsonPath)
	}
}

func (fb *FetchBase) DecodePlaylistJson(jsonPath string) {
	resourceJson, _ := os.Open(jsonPath)
	defer func(file *os.File) {
		_ = file.Close()
	}(resourceJson)
	resourceDecoder := json.NewDecoder(resourceJson)
	err := resourceDecoder.Decode(&fb)
	if err != nil {
		log.Fatalf("decoding json error:%v, json path:%s", err, jsonPath)
	}
}

func InitResources() {
	fetchBase := FetchBase{}
	fetchBase.DecodePlaylistJson(FetchBaseJsonPath)
	MediaBase = fetchBase.Playlists

	fetchHistory := FetchHistory{}
	fetchHistory.DecodePlaylistJson(FetchHistoryJsonPath)
	MediaHistory = fetchHistory.Playlists
}
