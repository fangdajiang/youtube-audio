package handler

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
	"strings"
)

//ITag format id
//  for _, f := range result.Formats() {
//		log.Infof("format id:%s", f.FormatID)
//	}
const ITag = "249"
const AudioFileExtensionName = ".ogg"

func DownloadYouTubeAudio(videoUrl string) error {
	log.Infof("Ready to downlod video %s", videoUrl)
	result, err := goutubedl.New(context.Background(), videoUrl, goutubedl.Options{})
	if err != nil {
		log.Fatal(err)
	}
	downloadResult, err := result.Download(context.Background(), ITag)
	if err != nil {
		log.Fatal(err)
	}
	defer downloadResult.Close()
	videoRawTitle := fmt.Sprintf("%s%s", result.Info.Title, AudioFileExtensionName)
	videoFileName := strings.Replace(videoRawTitle, "/", "-", -1)

	f, err := os.Create(videoFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	io.Copy(f, downloadResult)
	return nil
}
