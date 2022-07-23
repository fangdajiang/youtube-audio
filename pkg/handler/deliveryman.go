package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"os"
)

type Parcel struct {
	FilePath string
	Caption  string
}

type TelegramBot struct {
	Token         string //tg bot token, should be an admin in tg channel
	ChannelChatId string //tg channel's username only
	BotChatId     int64  //tg bot chat id
}

func Cleanup(parcel Parcel) {
	err := os.Remove(parcel.FilePath)
	if err != nil {
		log.Fatalf("removing file %s, error: %s", parcel.FilePath, err)
	} else {
		log.Infof("downloaded file cleaned up %s", parcel.FilePath)
	}
	log.Infof("file %s has been removed", parcel.FilePath)
}

func (t *TelegramBot) Send(parcel Parcel) error {
	log.Infof("%s is going to be sent", parcel.FilePath)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Errorf("new bot error, %s", err)
		return fmt.Errorf("building bot error")
	}

	log.Infof("ready to new audio to channel")
	msg := tgbotapi.NewAudioToChannel(t.ChannelChatId, tgbotapi.FilePath(parcel.FilePath))
	msg.Caption = parcel.Caption
	log.Infof("ready to send audio")

	_, err = bot.Send(msg)
	if err != nil {
		log.Errorf("bot send error, %s", err)
		return fmt.Errorf("sending audio error: %s", err)
	}
	log.Infof("audio %s has been sent", parcel.FilePath)

	return nil
}

func (t *TelegramBot) SendWarningMessage(desc string, warningMessage string) {
	log.Warnf("Ready to send warning message about %s to telegram bot", desc)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Fatalf("building msg bot error %s", err)
	}

	msg := tgbotapi.NewMessage(t.BotChatId, fmt.Sprintf(warningMessage, desc))

	_, err = bot.Send(msg)
	if err != nil {
		log.Fatalf("sending warning message error: %s", err)
	}
	log.Infof("Warning message %s has been sent", msg.Text)

}
