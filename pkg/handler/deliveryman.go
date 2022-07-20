package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"os"
)

type Parcel struct {
	filePath string
	caption  string
}

type TelegramBot struct {
	Token  string //tg bot token, should be an admin in tg channel
	ChatId string //tg channel's username only
}

func Cleanup(parcel Parcel) {
	err := os.Remove(parcel.filePath)
	if err != nil {
		log.Errorf("removing file %s, error: %s", parcel.filePath, err)
	} else {
		log.Infof("downloaded file cleaned up %s", parcel.filePath)
	}
}

func (t *TelegramBot) Send(parcel Parcel) error {
	log.Infof("%s will be sent", parcel.filePath)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("building bot error")
	}

	msg := tgbotapi.NewAudioToChannel(t.ChatId, tgbotapi.FilePath(parcel.filePath))
	msg.Caption = parcel.caption

	message, err := bot.Send(msg)
	if err != nil {
		log.Errorf("%s", err)
		return fmt.Errorf("sending audio error")
	}
	log.Debugf("audio sent response: %#v", message)

	return nil
}
