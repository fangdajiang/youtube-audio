package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	DownloadBaseJsonPath    string = "/Users/fangdajiang/IdeaProjects/youtube-audio/resource/download_base.json"
	DownloadHistoryJsonPath string = "/Users/fangdajiang/IdeaProjects/youtube-audio/resource/download_history.json"
)

var MediaBase []BaseProps
var MediaHistory []HistoryProps

type HistoryProps struct {
	Id                 string
	LastDownloadedTime string
	Urls               []string
}

type ParamItems struct {
	Id              string
	MaxResultsCount int64
	SortByPosition  bool
}

type BaseProps struct {
	Owner               string
	PrefixUrl           string
	MediaExtension      string
	DownloadedFilesPath string
	Params              []ParamItems
}

type DownloadBase struct {
	Playlists []BaseProps
}

type DownloadHistory struct {
	Playlists []HistoryProps
}

func (dh *DownloadHistory) DecodePlaylistJson(jsonPath string) {
	resourceJson, _ := os.Open(jsonPath)
	defer func(file *os.File) {
		_ = file.Close()
	}(resourceJson)
	resourceDecoder := json.NewDecoder(resourceJson)
	err := resourceDecoder.Decode(&dh)
	if err != nil {
		log.Fatalf("decoding json error:%v, json:%s", err, jsonPath)
	}
}

func (db *DownloadBase) DecodePlaylistJson(jsonPath string) {
	resourceJson, _ := os.Open(jsonPath)
	defer func(file *os.File) {
		_ = file.Close()
	}(resourceJson)
	resourceDecoder := json.NewDecoder(resourceJson)
	err := resourceDecoder.Decode(&db)
	if err != nil {
		log.Fatalf("decoding json error:%v, json:%s", err, jsonPath)
	}
}

func InitResources() {
	downloadBase := DownloadBase{}
	downloadBase.DecodePlaylistJson(DownloadBaseJsonPath)
	MediaBase = downloadBase.Playlists

	downloadHistory := DownloadHistory{}
	downloadHistory.DecodePlaylistJson(DownloadHistoryJsonPath)
	MediaHistory = downloadHistory.Playlists
}
