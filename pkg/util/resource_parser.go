package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	EnvYouTubeAudioHomeName  string = "APP_HOME"
	FetchBaseJsonPath        string = "/resource/fetch_base.json"
	FetchHistoryJsonPath     string = "/resource/fetch_history.json"
	TempFetchHistoryJsonPath string = "/resource/tmp_fetch_history.json"
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

func GetFetchJsonPath(jsonPath string) string {
	homePath, err := GetEnvVariable(EnvYouTubeAudioHomeName)
	if err != nil {
		log.Errorf("reading env %s vars error", EnvYouTubeAudioHomeName)
		return "/tmp"
	}
	return homePath + jsonPath
}

func getBaseProps() []BaseProps {
	fetchBase := FetchBase{}
	fetchBase.DecodePlaylistJson(GetFetchJsonPath(FetchBaseJsonPath))
	return fetchBase.Playlists
}

func getHistoryProps() []HistoryProps {
	fetchHistory := FetchHistory{}
	fetchHistory.DecodePlaylistJson(GetFetchJsonPath(FetchHistoryJsonPath))
	return fetchHistory.Playlists
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
	MediaBase = getBaseProps()

	MediaHistory = getHistoryProps()
}
