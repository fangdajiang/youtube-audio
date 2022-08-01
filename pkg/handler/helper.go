package handler

import (
	"errors"
	"fmt"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
	"youtube-audio/pkg/util"
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
	AudioFileExtensionName                string = ".webm"
	ResourceStorePath                     string = "/tmp/"
	YouTubeMaxResults                     int64  = 1
	YouTubePrefixUrl                      string = "https://www.youtube.com/watch?v="
	YouTubeChannelId                      string = "UU8UCbiPrm2zN9nZHKdTevZA"
	FailedToSendAudioWarningTemplate      string = "FAILED TO SEND AUDIO %s TO THE CHANNEL"
	FailedToDownloadAudioWarningTemplate  string = "FAILED TO DOWNLOAD AUDIO %s"
	InvalidDownloadedAudioWarningTemplate string = "INVALID DOWNLOADED FILE %s"
)

var YouTubePart = []string{"snippet"}

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
	return YouTubePrefixUrl + videoId
}

func FilenamifyMediaTitle(title string) (string, error) {
	rawMediaTitle := fmt.Sprintf("%s%s", title, AudioFileExtensionName)
	log.Infof("rawMediaTitle %s", rawMediaTitle)
	validMediaFileName, err := filenamify.Filenamify(rawMediaTitle, filenamify.Options{
		Replacement: IllegalCharacterReplacementInFilename,
		MaxLength:   FilenameMaxLength,
	})
	if err != nil {
		log.Errorf("convert raw media title to a valid file name error:%s", err)
		return "", fmt.Errorf("filenamify error: %s", rawMediaTitle)
	}
	log.Infof("validMediaFileName %s", validMediaFileName)

	return validMediaFileName, nil
}

func GenerateParcel(filePath string, caption string) Parcel {
	parcel := Parcel{
		FilePath: filePath,
		Caption:  caption,
	}
	return parcel
}

func GenerateYouTubeCredentials() (YouTubeCredentials, error) {
	var err error
	var youTubeCredentials YouTubeCredentials

	youtubeKey, err := util.GetEnvVariable(EnvYouTubeKeyName)
	if err != nil {
		log.Errorf("%s", err)
		return youTubeCredentials, fmt.Errorf("reading env %s vars error", EnvYouTubeKeyName)
	}
	youTubeCredentials.Key = youtubeKey

	return youTubeCredentials, nil
}

func GenerateTelegramBot() (TelegramBot, error) {
	var err error
	var telegramBot TelegramBot

	// Get the TOKEN and the CHAT_ID
	botToken, err := util.GetEnvVariable(EnvTokenName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvTokenName)
	}
	telegramBot.Token = botToken

	channelChatId, err := util.GetEnvVariable(EnvChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvChatIdName)
	}
	telegramBot.ChannelChatId = channelChatId

	botChatId, err := util.GetEnvVariable(EnvBotChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvBotChatIdName)
	}
	intBotChatId, _ := strconv.ParseInt(botChatId, 10, 64)
	telegramBot.BotChatId = intBotChatId

	return telegramBot, nil
}
