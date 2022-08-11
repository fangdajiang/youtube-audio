package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Delivery struct {
	Parcel     Parcel
	PlaylistId string
	Done       bool
	Timestamp  int64
	Datetime   string
}

type Parcel struct {
	FilePath string
	Caption  string
	Url      string
}

type TelegramBot struct {
	//sync.Mutex
	Token         string //tg bot token, should be an admin in tg channel
	ChannelChatId string //tg channel's username only
	BotChatId     int64  //tg bot chat id
}

func SendAudio(delivery *Delivery) error {
	telegramBot, err := GenerateTelegramBot()
	if err != nil {
		return err
	}
	// Send an audio file
	err = telegramBot.Send(delivery.Parcel)
	if err == nil {
		markDelivered(delivery)
	}
	return err
}

func markDelivered(delivery *Delivery) {
	//only for testing
	//rand.Seed(time.Now().UnixNano())
	//delivery.Done = rand.Float32() < 0.5

	delivery.Done = true
	if delivery.Timestamp == 0 {
		now := time.Now()
		delivery.Timestamp = now.Unix()
		delivery.Datetime = now.Format(DateTimeFormat)
	}
}

func SendMessage(desc string, warningMessage string) {
	telegramBot, err := GenerateTelegramBot()
	if err != nil {
		log.Errorf("%s", err)
	}
	telegramBot.SendWarningMessage(desc, warningMessage)
}

func IsAudioValid(parcel Parcel) bool {
	if parcel.FilePath == "" {
		log.Warnf("file path EMPTY: %v", parcel)
		return false
	}
	// exists?
	audioExists, err := FileExists(parcel.FilePath)
	if !audioExists {
		log.Warnf("downloaded file does NOT exist: %s, %v", parcel.FilePath, err)
		return false
	}
	// empty?
	audioFileInfo, _ := os.Stat(parcel.FilePath)
	log.Infof("audioFileInfo size: %v", audioFileInfo.Size())
	if audioFileInfo.Size() < 1024 {
		log.Warnf("downloaded file size(%v) is not BIG enough(>= 1024B): %s", audioFileInfo.Size(), parcel.FilePath)
		return false
	}
	return true
}

func Cleanup(parcel Parcel) {
	parcelExists, err := FileExists(parcel.FilePath)
	if !parcelExists {
		log.Warnf("parcel file does NOT exist: %s, %v", parcel.FilePath, err)
		return
	}
	err = os.Remove(parcel.FilePath)
	if err != nil {
		log.Errorf("removing file %s, error: %s", parcel.FilePath, err)
	} else {
		log.Infof("downloaded file cleaned up %s", parcel.FilePath)
	}
	log.Infof("file %s has been removed", parcel.FilePath)
}

func (t *TelegramBot) Send(parcel Parcel) error {
	//t.Lock()
	//defer t.Unlock()

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
	//t.Lock()
	//defer t.Unlock()

	log.Warnf("Ready to send warning message about %s to telegram bot", desc)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Errorf("building msg bot error %s", err)
		return
	}

	msg := tgbotapi.NewMessage(t.BotChatId, fmt.Sprintf(warningMessage, desc))

	_, err = bot.Send(msg)
	if err != nil {
		log.Errorf("sending warning message error: %s", err)
		return
	}
	log.Infof("Warning message %s has been sent", msg.Text)

}
