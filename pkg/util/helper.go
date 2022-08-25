package util

import (
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/flytam/filenamify"
	"os"
	"strings"
	"time"
	"youtube-audio/pkg/util/log"
)

const (
	DateTimeFormat                        string = "2006-01-02 15:04:05"
	YouTubeDateTimeFormat                 string = "2006-01-02T15:04:05Z"
	EnvTokenName                          string = "BOT_TOKEN"
	EnvChatIdName                         string = "CHAT_ID"
	EnvBotChatIdName                      string = "BOT_CHAT_ID"
	EnvYouTubeKeyName                     string = "YOUTUBE_KEY"
	IllegalCharacterReplacementInFilename string = "_"
	FilenameMaxLength                     int    = 512
	UploadAudioMaxLength                  int64  = 52428800
	YouTubeDefaultMaxResults              int64  = 1
	FetchYouTubeMaxResultsLimit           int64  = 50
	FetchMaxUrlsLimit                     int    = 10
	FailedToSendAudioWarningTemplate      string = "FAILED TO SEND AUDIO %s TO THE CHANNEL"
	FailedToFetchAudioWarningTemplate     string = "FAILED TO FETCH AUDIO %v"
	FileNotExistWarningTemplate           string = "FILE NOT EXIST: *%s*"
	EmptyFilePathWarningTemplate          string = "EMPTY FILE PATH: *%s*"
	InvalidFileSizeWarningTemplate        string = "INVALID FILE SIZE: *%s*"
)

var YouTubePart = []string{"snippet"}

func StringSliceContains(sl []string, s string) bool {
	sl2 := StringSlice2Interface(sl)
	set := mapset.NewSetFromSlice(sl2)
	return set.Contains(s)
}

func StringSlice2Interface(sl []string) []interface{} {
	sl2 := make([]interface{}, 0)
	for _, str := range sl {
		sl2 = append(sl2, str)
	}
	return sl2
}

func DeleteSliceElms(sl []HistoryProps, elms ...HistoryProps) []HistoryProps {
	if len(sl) == 0 || len(elms) == 0 {
		return sl
	}
	// 先将元素转为 set
	m := make(map[string]HistoryProps)
	for _, v := range elms {
		m[v.Id] = HistoryProps{}
	}
	// 过滤掉指定元素
	res := make([]HistoryProps, 0, len(sl))
	for _, v := range sl {
		if _, ok := m[v.Id]; !ok {
			res = append(res, v)
		}
	}
	return res
}

func GetYouTubeFetchBase() BaseProps {
	youtube := BaseProps{}
	for _, cp := range MediaBase {
		if cp.Owner == "YouTube" {
			youtube = cp
			break
		}
		continue
	}
	if youtube.Owner == "" {
		log.Fatalf("getting youtube playlists from json error, util.MediaBase:%v", MediaBase)
	}
	return youtube
}

func GetYouTubePlaylistMaxResultsCount(playlistId string) int64 {
	youtube := GetYouTubeFetchBase()
	for _, scope := range youtube.Params {
		if scope.Id == playlistId {
			return scope.MaxResultsCount
		}
	}
	if youtube.Owner == "" {
		log.Errorf("getting youtube playlist max results count from json error, scopes:%v", youtube.Params)
	}
	return 1
}

func GetLocalDateTime(formattedDateTime string) string {
	youtubeTime, err := time.Parse(YouTubeDateTimeFormat, formattedDateTime)
	if err != nil {
		log.Errorf("parse youtube datetime error:%s, formattedDateTime: %s with format: %s", err, formattedDateTime, YouTubeDateTimeFormat)
		return formattedDateTime
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return youtubeTime.In(loc).Format(DateTimeFormat)
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func MakeYouTubeRawUrl(videoId string) string {
	return GetYouTubeFetchBase().PrefixUrl + videoId
}

func FilenamifyMediaTitle(title string) (string, error) {
	rawMediaTitle := fmt.Sprintf("%s%s", title, GetYouTubeFetchBase().MediaExtension)
	log.Debugf("rawMediaTitle %s", rawMediaTitle)
	validMediaFileName, err := filenamify.Filenamify(rawMediaTitle, filenamify.Options{
		Replacement: IllegalCharacterReplacementInFilename,
		MaxLength:   FilenameMaxLength,
	})
	if err != nil {
		log.Errorf("convert raw media title to a valid file name error:%s", err)
		return "", fmt.Errorf("filenamify error: %s", rawMediaTitle)
	}
	validMediaFileName = strings.ReplaceAll(validMediaFileName, "#", IllegalCharacterReplacementInFilename)
	validMediaFileName = strings.ReplaceAll(validMediaFileName, " ", "")
	log.Debugf("validMediaFileName %s", validMediaFileName)

	return validMediaFileName, nil
}
