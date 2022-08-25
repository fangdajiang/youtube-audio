package resource

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"youtube-audio/pkg/util/log"
)

func TestInitResources(t *testing.T) {
	InitResources()
	log.Debugf("base: %v", MediaBase[0])
	log.Debugf("history: %v", MediaHistory[0])
}

func TestFetchBase_DecodePlaylistJson(t *testing.T) {
	MediaBase = getBaseProps()
	log.Debugf("base: %v", MediaBase[0])
}

func TestMarshalUnMarshalJson(t *testing.T) {
	r := require.New(t)

	var fetchHistory = FetchHistory{getPlaylistsExample()}
	bytes1, err := json.Marshal(&fetchHistory)
	r.NoError(err)

	log.Debugf("fetchHistory: %v", string(bytes1))

	var unknownTypeData = []byte(string(bytes1))
	var f interface{}
	err = json.Unmarshal(unknownTypeData, &f)
	r.NoError(err)

	rootMap := f.(map[string]interface{})
	log.Debugf("rootMap: %s", rootMap)
	r.Equal(nil, rootMap["not_exist_field"])
	var playlistsValue = rootMap["playlists"]
	log.Debugf("playlistsValue: %s", playlistsValue)

	playlistsArray := playlistsValue.([]interface{})
	log.Debugf("playlistsArray[0]: %s", playlistsArray[0])

	playlistsMap := playlistsArray[0].(map[string]interface{})
	log.Debugf("playlistsMap: %s", playlistsMap)
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

func TestMarshalPlaylistJson(t *testing.T) {
	fetchHistory := FetchHistory{getPlaylistsExample()}
	MarshalPlaylistJson(fetchHistory)
}
