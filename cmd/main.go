package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"youtube-audio/pkg/handler"
)

func main() {
	fmt.Println("Start fetching, converting, sending...")

	// Add a flag
	var audioFilePath string
	flag.StringVar(&audioFilePath, "audio-file", "", "This audio file will be sent.")
	flag.Parse()
	// Send an audio file
	err := handler.SendLocalAudioFile(audioFilePath)

	if err != nil {
		log.Fatalf("%s", err)
	}
}
