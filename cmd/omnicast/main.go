package main

import (
	"log"
	"net/url"
	"os"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/gcast"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("missing command/arugments")
	}

	if os.Args[1] != "load" {
		log.Fatalf("Unsupported command: %s\n", os.Args[1])
	}

	mediaURL := os.Args[2]

	var player omnicast.MediaPlayer
	player, err := gcast.Find()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(player.Name(), mediaURL)

	uri, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		log.Fatal(err)
	}

	err = player.Load(uri, nil)
	if err != nil {
		log.Fatal(err)
	}
}
