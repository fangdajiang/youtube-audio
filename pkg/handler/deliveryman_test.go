package handler

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	LocalFilePath    string = "/tmp/test.txt"
	LocalFileCaption string = "春眠不觉晓"
)

func TestCleanup(t *testing.T) {
}

func TestTelegramBot_Send(t *testing.T) {
	r := require.New(t)

	telegramBot, err := GenerateTelegramBot()
	r.NoError(err)

	parcel := GenerateParcel(LocalFilePath, LocalFileCaption+time.Now().Format("2006-01-02 15:04:05"))

	f, err := os.Create(parcel.filePath)
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
