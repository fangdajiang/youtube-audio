package util

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/flytam/filenamify"
	"strings"
	"time"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/oss"
	"youtube-audio/pkg/util/resource"
)

const (
	DateTimeFormat                        string = "2006-01-02 15:04:05"
	YouTubeDateTimeFormat                 string = "2006-01-02T15:04:05Z"
	EnvTokenName                          string = "BOT_TOKEN"
	EnvChatIdName                         string = "CHAT_ID"
	EnvBotChatIdName                      string = "BOT_CHAT_ID"
	EnvYouTubeKeyName                     string = "YOUTUBE_KEY"
	IllegalCharacterReplacementInFilename string = "_"
	FilenameMaxLength                     int    = 250 // 255 - 5, 5 means length of .webm
	UploadAudioMaxLength                  int64  = 52428800
	YouTubeDefaultMaxResults              int64  = 3
	FetchYouTubeMaxResultsLimit           int64  = 15
	FetchBlockMaxUrlsLimit                int    = 10
	FailedToSendAudioWarningTemplate      string = "FAILED TO SEND AUDIO %s TO THE CHANNEL"
	FailedToFetchAudioWarningTemplate     string = "FAILED TO FETCH AUDIO %v"
	FileNotExistWarningTemplate           string = "FILE NOT EXIST: *%s*"
	EmptyFilePathWarningTemplate          string = "EMPTY FILE PATH: *%s*"
	InvalidFileSizeWarningTemplate        string = "INVALID FILE SIZE: *%s*"
)

var YouTubePart = []string{"snippet"}

func UploadLog2Oss() {
	oss.Upload2Oss(log.LoggingFilePath)
}

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

func DeleteSliceElms(sl []resource.HistoryProps, elms ...resource.HistoryProps) []resource.HistoryProps {
	if len(sl) == 0 || len(elms) == 0 {
		return sl
	}
	// 先将元素转为 set
	m := make(map[string]resource.HistoryProps)
	for _, v := range elms {
		m[v.Id] = resource.HistoryProps{}
	}
	// 过滤掉指定元素
	res := make([]resource.HistoryProps, 0, len(sl))
	for _, v := range sl {
		if _, ok := m[v.Id]; !ok {
			res = append(res, v)
		}
	}
	return res
}

func GetYouTubeFetchBase() resource.BaseProps {
	youtube := resource.BaseProps{}
	for _, cp := range resource.MediaBase {
		if cp.Owner == "YouTube" {
			youtube = cp
			break
		}
		continue
	}
	if youtube.Owner == "" {
		log.Fatalf("getting youtube playlists from json error, util.MediaBase:%v", resource.MediaBase)
	}
	return youtube
}

func GetYouTubePlaylistMaxResultsCount(playlistId string) int64 {
	youtube := GetYouTubeFetchBase()
	for _, scope := range youtube.Params {
		if scope.Id == playlistId {
			maxResults := scope.MaxResultsCount
			if maxResults <= 0 || maxResults > FetchYouTubeMaxResultsLimit {
				log.Errorf("illegal maxResults error:%v", maxResults)
				maxResults = YouTubeDefaultMaxResults
			}
			return maxResults
		}
	}
	if youtube.Owner == "" {
		log.Errorf("getting youtube playlist max results count from json error, scopes:%v", youtube.Params)
	}
	return YouTubeDefaultMaxResults
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

func MakeYouTubeRawUrl(videoId string) string {
	return GetYouTubeFetchBase().PrefixUrl + videoId
}

func FilenamifyMediaTitle(title string) (string, error) {
	log.Debugf("rawMediaTitle %s, length: %v", title, len(title))
	validMediaFileName, err := filenamify.Filenamify(title, filenamify.Options{
		Replacement: IllegalCharacterReplacementInFilename,
		MaxLength:   FilenameMaxLength,
	})
	if err != nil {
		log.Errorf("convert raw media title to a valid file name error:%s", err)
		return "", fmt.Errorf("filenamify error: %s", title)
	}
	validMediaFileName = strings.ReplaceAll(validMediaFileName, "#", IllegalCharacterReplacementInFilename)
	validMediaFileName = strings.ReplaceAll(validMediaFileName, " ", "")
	validMediaFileName = fmt.Sprintf("%s%s", validMediaFileName, GetYouTubeFetchBase().MediaExtension)

	return validMediaFileName, nil
}
