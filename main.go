package main

import (
	"youtube-audio/cmd"
)

var (
	version = "0.2.0"
)

func main() {
	cmd.Execute(version)
}
