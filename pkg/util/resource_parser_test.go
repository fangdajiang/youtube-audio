package util

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestInitResources(t *testing.T) {
	InitResources()
	log.Infof("channel: %v", MediaChannels[0])
	log.Infof("history: %v", MediaHistory[0])
}
