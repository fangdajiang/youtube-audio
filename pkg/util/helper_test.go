package util

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

func TestGetLocalDateTime(t *testing.T) {
	r := require.New(t)

	publishedAt := "2022-07-31T14:29:03Z"
	remotePublishedAt, err := time.Parse(YouTubeDateTimeFormat, publishedAt)
	r.NoError(err)
	localPublishedAt := GetLocalDateTime(publishedAt)

	log.Debugf("remotePublishedAt:%s, localPublishedAt: %s", remotePublishedAt.Format(DateTimeFormat), localPublishedAt)

}

func TestFilenamifyMediaTitle(t *testing.T) {
	r := require.New(t)

	mediaTitle := "中文abc/标题\\_123!_def`_gh'_done #shorts"

	namifiedMediaTitle, err := FilenamifyMediaTitle(mediaTitle)
	r.NoError(err)
	r.Greater(len(namifiedMediaTitle), len(mediaTitle))
	log.Debugf("mediaTitle: %v", len(mediaTitle))
	log.Debugf("namifiedMediaTitle: %v", len(namifiedMediaTitle))
}

func TestFileExists(t *testing.T) {
	r := require.New(t)

	filePath := "/tmp/test.txt"

	exists, err := myio.FileExists(filePath)
	r.NoError(err)
	r.True(exists, "file NOT exists: %s", filePath)
}

func TestDeleteSliceElms(t *testing.T) {
	historyProps := resource.MediaHistory
	log.Debugf("historyProps count: %v, historyProps: %v", len(historyProps), historyProps)
	hp := resource.HistoryProps{Id: "PLt-jD7OCbLJ0ZMwvQFZCSuNaUbT3GMqVJ"}
	result := DeleteSliceElms(historyProps, hp)
	log.Debugf("result count: %v, result: %v", len(result), result)
	log.Debugf("historyProps count2: %v", len(historyProps))

}

func TestContainString(t *testing.T) {
	r := require.New(t)

	// test single youtube url
	sl := []string{"https://www.youtube.com/watch?v=jQZ36-zERtM"}
	log.Debugf("sl: %v", sl)

	s1 := StringSlice2Interface(sl)

	set := mapset.NewSetFromSlice(s1)

	r.True(set.Contains("https://www.youtube.com/watch?v=jQZ36-zERtM"), "contains https://www.youtube.com/watch?v=jQZ36-zERtM")
	r.False(set.Contains("https://www.youtube.com/watch?v=123"), "NOT contains https://www.youtube.com/watch?v=123")

	// test playlist url
	sl2 := []string{"https://www.youtube.com/playlist?list=PLstzraCE5l2j_Sih-L9CoFq-r71NflIfi"}
	log.Debugf("sl2: %v", sl2)

	s2 := StringSlice2Interface(sl2)

	set2 := mapset.NewSetFromSlice(s2)

	r.True(set2.Contains("https://www.youtube.com/playlist?list=PLstzraCE5l2j_Sih-L9CoFq-r71NflIfi"), "contains https://www.youtube.com/playlist?list=")
	r.False(set2.Contains("https://www.youtube.com/playlist?list2="), "NOT contains https://www.youtube.com/playlist?list2=")

}

func init() {
	log.InitLogging()
}
