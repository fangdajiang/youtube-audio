package handler

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilenamifyMediaTitle(t *testing.T) {
	r := require.New(t)

	mediaTitle := "中文abc/标题\\_123!_def`_gh'_done"

	namifiedMediaTitle, err := FilenamifyMediaTitle(mediaTitle)
	r.NoError(err)
	r.Greater(len(namifiedMediaTitle), len(mediaTitle))
	log.Infof("mediaTitle: %v", len(mediaTitle))
	log.Infof("namifiedMediaTitle: %v", len(namifiedMediaTitle))
}

func TestFileExists(t *testing.T) {
	r := require.New(t)

	filePath := "/tmp/test.txt"

	exists, err := FileExists(filePath)
	r.NoError(err)
	r.True(exists, "file NOT exists: %s", filePath)
}
