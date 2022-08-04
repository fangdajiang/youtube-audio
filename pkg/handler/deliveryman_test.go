package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	LocalFilePath    string = "/tmp/test.txt"
	LocalFileCaption string = "春眠不觉晓"
	UselessUrl       string = "https://www.youtube.com/watch?v=Xy8BOay7hDc"
)

var (
	telegramBot TelegramBot
	parcel      Parcel
)

func init() {
	log.Infof("initing deliveryman test")

	telegramBot, _ = GenerateTelegramBot()
	parcel = GenerateParcel(LocalFilePath, LocalFileCaption+time.Now().Format(DateTimeFormat), UselessUrl)

}

func TestTelegramBot_RetrieveUpdates(t *testing.T) {
	r := require.New(t)

	bot, err := tgbotapi.NewBotAPI(telegramBot.Token)
	r.NoError(err)
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

}

func TestTelegramBot_Send(t *testing.T) {
	r := require.New(t)

	f, err := os.Create(parcel.FilePath)
	r.NoError(err)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	_, err = f.WriteString("Hello Test")
	r.NoError(err)

	err = telegramBot.Send(parcel)
	r.NoError(err)

	Cleanup(parcel)

}

func TestCleanup(t *testing.T) {
}

func TestTelegramBot_SendWarningMessage(t *testing.T) {
	r := require.New(t)

	f, err := os.Create(parcel.FilePath)
	r.NoError(err)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	_, err = f.WriteString("Hello Test")
	r.NoError(err)

	telegramBot.SendWarningMessage(parcel.Caption, FailedToSendAudioWarningTemplate)
	Cleanup(parcel)
}

func TestIsAudioValid(t *testing.T) {
	r := require.New(t)

	filePath := "/tmp/test.txt"

	valid := IsAudioValid(parcel)
	r.True(valid, "invalid file: %s", filePath)
}
