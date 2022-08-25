package myio

import (
	"errors"
	"os"
	"youtube-audio/pkg/util/log"
)

func Cleanup(filePath string) {
	parcelExists, err := FileExists(filePath)
	if !parcelExists {
		log.Warnf("filePath file does NOT exist: %s, %v", filePath, err)
		return
	}
	err = os.Remove(filePath)
	if err != nil {
		log.Errorf("removing file %s, error: %s", filePath, err)
	} else {
		log.Debugf("file cleaned up %s", filePath)
	}
	log.Infof("file %s has been removed", filePath)
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
