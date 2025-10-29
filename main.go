package main

import (
	"log"

	"github.com/subosito/gotenv"

	"youtube-audio/cmd"
)

var (
	version = "0.2.0"
)

func main() {
	if err := gotenv.Load(); err != nil {
		log.Printf("dotenv load warning: %v", err)
	}
	cmd.Execute(version)
}
