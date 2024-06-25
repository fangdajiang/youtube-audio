package db

import (
	"testing"
	"youtube-audio/pkg/util/log"
)

func TestConnectLocal(t *testing.T) {
	log.Debugf("hello")
	var tests []struct {
		name string
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConnectLocal()
		})
	}
}
func init() {
	log.InitLogging()
}
