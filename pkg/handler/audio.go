package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/util"
)

const (
	Token  string = "TOKEN"
	ChatId string = "CHAT_ID"
)

var (
	botToken  string //tg bot's token, should be an admin in tg channel
	botChatId string //tg channel's username only
)

func SendLocalAudioFile(localFilePath string) error {
	log.Infof("%s will be sent", localFilePath)
	var err error

	// Get the TOKEN and the CHAT_ID
	botToken, err = util.GetEnvVariable(Token)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("reading env %s vars error", Token)
	}
	botChatId, err = util.GetEnvVariable(ChatId)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("reading env %s vars error", ChatId)
	}
	log.Infof("chat id string: %s", botChatId)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("building bot error")
	}

	msg := tgbotapi.NewAudioToChannel(botChatId, tgbotapi.FilePath(localFilePath))

	message, err := bot.Send(msg)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("sending audio error")
	}
	log.Debugf("audio sent response: %#v", message)

	return nil
}
