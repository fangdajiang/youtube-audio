package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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
	var fetchHistory = FetchHistory{getPlaylistsExample()}
	log.Infof("fetchHistory: %v", fetchHistory)

	EncodePlaylistJson(TempFetchHistoryJsonPath, fetchHistory)
}

func TestMarshalUnMarshalJson(t *testing.T) {
	r := require.New(t)

	var fetchHistory = FetchHistory{getPlaylistsExample()}
	bytes1, err := json.Marshal(&fetchHistory)
	r.NoError(err)

	log.Infof("fetchHistory: %v", string(bytes1))

	var unknownTypeData = []byte(string(bytes1))
	var f interface{}
	err = json.Unmarshal(unknownTypeData, &f)
	r.NoError(err)

	rootMap := f.(map[string]interface{})
	log.Infof("rootMap: %s", rootMap)
	r.Equal(nil, rootMap["not_exist_field"])
	var playlistsValue = rootMap["playlists"]
	log.Infof("playlistsValue: %s", playlistsValue)

	playlistsArray := playlistsValue.([]interface{})
	log.Infof("playlistsArray[0]: %s", playlistsArray[0])

	playlistsMap := playlistsArray[0].(map[string]interface{})
	log.Infof("playlistsMap: %s", playlistsMap)
	r.Equal("PL", playlistsMap["id"])

}

func getPlaylistsExample() []HistoryProps {
	var urls = []string{"xxx", "yyy"}
	var lastFetch = FetchItems{"1", 1, urls}
	var nextFetch = FetchItems{"3", 3, urls}
	var subscriberItem = SubscriberItems{11, lastFetch, nextFetch}
	var subscriberItems = []SubscriberItems{subscriberItem}
	var historyProps = HistoryProps{"PL", subscriberItems}
	return []HistoryProps{historyProps}
}
