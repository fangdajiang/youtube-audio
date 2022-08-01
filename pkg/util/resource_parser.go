package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	DownloadBaseJsonPath    string = "./resource/download_base.json"
	DownloadHistoryJsonPath string = "./resource/download_history.json"
)

var MediaChannels []ChannelProps
var MediaHistory []HistoryProps

type HistoryProps struct {
	Id                 string
	LastDownloadedTime string
	Urls               []string
}

type ScopeProps struct {
	Id              string
	MaxResultsCount int64
	SortByPosition  bool
}

type ChannelProps struct {
	Owner               string
	PrefixUrl           string
	MediaExtension      string
	DownloadedFilesPath string
	Scopes              []ScopeProps
}

type DownloadBase struct {
	Channels []ChannelProps
}

type DownloadHistory struct {
	Channels []HistoryProps
}

func (dh *DownloadHistory) DecodeChannelJson(jsonPath string) {
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

func (db *DownloadBase) DecodeChannelJson(jsonPath string) {
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
	downloadBase.DecodeChannelJson(DownloadBaseJsonPath)
	MediaChannels = downloadBase.Channels

	downloadHistory := DownloadHistory{}
	downloadHistory.DecodeChannelJson(DownloadHistoryJsonPath)
	MediaHistory = downloadHistory.Channels
}
