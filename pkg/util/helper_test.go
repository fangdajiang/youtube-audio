package util

import (
	mapset "github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetLocalDateTime(t *testing.T) {
	r := require.New(t)

	publishedAt := "2022-07-31T14:29:03Z"
	remotePublishedAt, err := time.Parse(YouTubeDateTimeFormat, publishedAt)
	r.NoError(err)
	localPublishedAt := GetLocalDateTime(publishedAt)

	log.Infof("remotePublishedAt:%s, localPublishedAt: %s", remotePublishedAt.Format(DateTimeFormat), localPublishedAt)

}

func TestFilenamifyMediaTitle(t *testing.T) {
	r := require.New(t)

	mediaTitle := "中文abc/标题\\_123!_def`_gh'_done #shorts"

	namifiedMediaTitle, err := FilenamifyMediaTitle(mediaTitle)
	r.NoError(err)
	r.Greater(len(namifiedMediaTitle), len(mediaTitle))
	log.Infof("mediaTitle: %v", len(mediaTitle))
	log.Infof("namifiedMediaTitle: %v", len(namifiedMediaTitle))
}

func TestFileExists(t *testing.T) {
	r := require.New(t)

	filePath := "/tmp/test.txt"

	exists, err := FileExists(filePath)
	r.NoError(err)
	r.True(exists, "file NOT exists: %s", filePath)
}

func TestDeleteSliceElms(t *testing.T) {
	historyProps := MediaHistory
	log.Infof("historyProps count: %v, historyProps: %v", len(historyProps), historyProps)
	hp := HistoryProps{Id: "PLt-jD7OCbLJ0ZMwvQFZCSuNaUbT3GMqVJ"}
	result := DeleteSliceElms(historyProps, hp)
	log.Infof("result count: %v, result: %v", len(result), result)
	log.Infof("historyProps count2: %v", len(historyProps))

}

func TestContainString(t *testing.T) {
	sl := []string{"https://www.youtube.com/watch?v=jQZ36-zERtM"}
	log.Infof("sl: %v", sl)

	sl2 := StringSlice2Interface(sl)

	set := mapset.NewSetFromSlice(sl2)
	log.Infof("contains https://www.youtube.com/watch?v=jQZ36-zERtM: %v", set.Contains("https://www.youtube.com/watch?v=jQZ36-zERtM"))
	log.Infof("contains https://www.youtube.com/watch?v=JOj-k2UE2Bs: %v", set.Contains("https://www.youtube.com/watch?v=JOj-k2UE2Bs"))

}
