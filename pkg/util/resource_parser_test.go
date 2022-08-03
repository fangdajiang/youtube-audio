package util

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestInitResources(t *testing.T) {
	InitResources()
	log.Infof("base: %v", MediaBase[0])
	log.Infof("history: %v", MediaHistory[0])
}

func TestFetchBase_DecodePlaylistJson(t *testing.T) {
	fetchBase := FetchBase{}
	fetchBase.DecodePlaylistJson(FetchBaseJsonPath)
	MediaBase = fetchBase.Playlists
	log.Infof("base: %v", MediaBase[0])
}

func TestFetchHistory_EncodePlaylistJson(t *testing.T) {
	var urls = []string{"xxx", "yyy"}
	var lastFetch = FetchItems{"1", 1, urls}
	var nextFetch = FetchItems{"3", 3, urls}
	var subscriberItem = SubscriberItems{11, lastFetch, nextFetch}
	var subscriberItems = []SubscriberItems{subscriberItem}
	var historyProps = HistoryProps{"PL", subscriberItems}
	var playlists = []HistoryProps{historyProps}

	var fetchHistory = FetchHistory{playlists}
	log.Infof("fetchHistory: %v", fetchHistory)

	EncodePlaylistJson(TempFetchHistoryJsonPath, fetchHistory)
}
