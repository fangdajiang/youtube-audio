package handler

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
)

//ITag format id
//  for _, f := range result.Formats() {
//		log.Infof("format id:%s", f.FormatID)
//	}

func DownloadYouTubeAudio(mediaUrl string) (Parcel, error) {
	var parcel Parcel
	log.Infof("Ready to downlod media %s", mediaUrl)
	result, err := goutubedl.New(context.Background(), mediaUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %s", mediaUrl)
	}
	downloadedResult, err := result.Download(context.Background(), ITag)
	if err != nil {
		log.Errorf("download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %s, ITag: %s", mediaUrl, ITag)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)

	validMediaFileName, err := FilenamifyMediaTitle(result.Info.Title)
	if err != nil {
		return parcel, err
	}
	audioFilePath := fmt.Sprintf("%s%s", ResourceStorePath, validMediaFileName)
	parcel.filePath = audioFilePath
	parcel.caption = result.Info.Title
	log.Debugf("parcel: %v", parcel)

	f, err := os.Create(audioFilePath)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(f, downloadedResult)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	return parcel, nil
}
