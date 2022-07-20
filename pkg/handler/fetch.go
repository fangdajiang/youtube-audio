package handler

import (
	"context"
	"fmt"
	"github.com/flytam/filenamify"
	log "github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
)

//ITag format id
//  for _, f := range result.Formats() {
//		log.Infof("format id:%s", f.FormatID)
//	}
const ITag = "249"
const AudioFileExtensionName = ".ogg"
const ResourceStorePath = "/tmp/"

func DownloadYouTubeAudio(videoUrl string) (string, string, error) {
	log.Infof("Ready to downlod video %s", videoUrl)
	result, err := goutubedl.New(context.Background(), videoUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return "", "", fmt.Errorf("goutubedl new error: %s", videoUrl)
	}
	downloadResult, err := result.Download(context.Background(), ITag)
	if err != nil {
		log.Errorf("download error:%s", err)
		return "", "", fmt.Errorf("goutubedl download error: %s, ITag: %s", videoUrl, ITag)
	}
	defer downloadResult.Close()
	videoRawTitle := fmt.Sprintf("%s%s", result.Info.Title, AudioFileExtensionName)
	log.Infof("videoRawTitle %s", videoRawTitle)
	videoValidFileName, err := filenamify.Filenamify(videoRawTitle, filenamify.Options{
		Replacement: "_",
		MaxLength:   512,
	})
	if err != nil {
		log.Errorf("convert to valid file name error:%s", err)
		return "", "", fmt.Errorf("filenamify conversion error: %s, AudioFileExtensionName: %s", videoRawTitle, AudioFileExtensionName)
	}
	log.Infof("videoValidFileName %s", videoValidFileName)
	//videoFileName := strings.Replace(videoRawTitle, "/", "_", -1)
	audioFilePath := fmt.Sprintf("%s%s", ResourceStorePath, videoValidFileName)
	log.Debugf("audioFilePath: %s", audioFilePath)

	f, err := os.Create(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	io.Copy(f, downloadResult)

	return audioFilePath, result.Info.Title, nil
}
