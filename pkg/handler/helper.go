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
	AudioFileExtensionName                string = ".ogg"
	IllegalCharacterReplacementInFilename string = "_"
	FilenameMaxLength                     int    = 512
	ITag                                  string = "249"
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
	log.Infof("chat id string: %s", botChatId)
	telegramBot.ChatId = botChatId

	return telegramBot, nil
}
