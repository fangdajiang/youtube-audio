package handler

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/wader/goutubedl"
	"io"
	"os"
)

//ITagNo format id
//  for _, f := range result.Formats() {
//		log.Infof("format id:%s", f.FormatID)
//	}

func DownloadYouTubeAudioToPath(mediaUrl string) (Parcel, error) {
	var parcel Parcel
	log.Infof("Ready to downlod media %s", mediaUrl)
	result, err := goutubedl.New(context.Background(), mediaUrl, goutubedl.Options{})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %s", mediaUrl)
	}
	downloadedResult, err := result.Download(context.Background(), ITagNo)
	if err != nil {
		log.Errorf("download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %s, ITagNo: %s", mediaUrl, ITagNo)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)

	validMediaFileName, err := FilenamifyMediaTitle(result.Info.Title)
	if err != nil {
		return parcel, err
	}
	parcel = GenerateParcel(fmt.Sprintf("%s%s", ResourceStorePath, validMediaFileName), result.Info.Title)
	log.Debugf("parcel: %v", parcel)

	parcelFile, err := os.Create(parcel.filePath)
	if err != nil {
		log.Fatal(err)
	}
	written, err := io.Copy(parcelFile, downloadedResult)
	if err != nil {
		log.Fatalf("copy error: %s, parcel: %s, written: %v", err, parcel, written)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(parcelFile)

	return parcel, nil
}
