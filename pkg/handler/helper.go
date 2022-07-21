package handler

import (
	"fmt"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/util"
)

const (
	EnvTokenName                          string = "TOKEN"
	EnvChatIdName                         string = "CHAT_ID"
	IllegalCharacterReplacementInFilename string = "_"
	FilenameMaxLength                     int    = 512
	ITagNo                                string = "249"
	AudioFileExtensionName                string = ".ogg"
	ResourceStorePath                     string = "/tmp/"
)

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
		filePath: filePath,
		caption:  caption,
	}
	return parcel
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

	botChatId, err := util.GetEnvVariable(EnvChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", EnvChatIdName)
	}
	telegramBot.ChatId = botChatId

	return telegramBot, nil
}
