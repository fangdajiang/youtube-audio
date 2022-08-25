package util

import (
	"encoding/json"
	"strings"
	"youtube-audio/pkg/util/log"
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

func (fb *FetchBase) DecodePlaylistJson(jsonFileName string) {
	fetchJson, err := GetResourceJson(jsonFileName)
	if err != nil {
		log.Errorf("get resource json error, jsonFileName: %s, error: %s", jsonFileName, err)
		return
	}
	resourceDecoder := json.NewDecoder(strings.NewReader(fetchJson))
	err = resourceDecoder.Decode(&fb)
	if err != nil {
		log.Fatalf("decoding json error:%v, jsonFileName:%s", err, jsonFileName)
	}
}

func (fh *FetchHistory) DecodePlaylistJson(jsonFileName string) {
	fetchJson, err := GetResourceJson(jsonFileName)
	if err != nil {
		log.Errorf("get resource json error, jsonFileName: %s, error: %s", jsonFileName, err)
		return
	}
	resourceDecoder := json.NewDecoder(strings.NewReader(fetchJson))
	err = resourceDecoder.Decode(&fh)
	if err != nil {
		log.Fatalf("decoding json error:%v, jsonFileName:%s", err, jsonFileName)
	}
}

func getBaseProps() []BaseProps {
	fetchBase := FetchBase{}
	fetchBase.DecodePlaylistJson(FetchBaseFileName)
	return fetchBase.Playlists
}

func getHistoryProps() []HistoryProps {
	fetchHistory := FetchHistory{}
	fetchHistory.DecodePlaylistJson(FetchHistoryFileName)
	return fetchHistory.Playlists
}

func MarshalPlaylistJson(fetchHistory FetchHistory) {
	jsonBytes, err := json.MarshalIndent(&fetchHistory, "", "  ")
	if err != nil {
		log.Fatalf("marshal indent error:%v, ossFileName:%s, fetchHistory:%v", err, FetchHistoryFileName, fetchHistory)
	}
	log.Debugf("fetchHistory json: %v", string(jsonBytes))
	UpdateResourceJson(FetchHistoryFileName, string(jsonBytes))
}

func InitResources() {
	MediaBase = getBaseProps()

	MediaHistory = getHistoryProps()
}
