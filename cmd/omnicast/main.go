package main

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/gcast"
)

func gcastPlayer() (*gcast.Sender, error) {
	ctx, stopDiscovery := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	ch, err := gcast.Discover(ctx)
	if err != nil {
		return nil, err
	}

	for dev := range ch {
		if dev.CapableOf(gcast.VideoOut, gcast.AudioOut) {
			log.Printf("Found Google Cast device: %s (%s)\n", dev.Name, dev.UUID)

			sender, err := gcast.NewSender("sender-omnicast", dev)
			if err != nil {
				return nil, err
			}

			stopDiscovery()
			return sender, nil
		}
	}

	return nil, errors.New("no Google Cast device found")
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("missing command/arugments")
	}

	if os.Args[1] != "load" {
		log.Fatalf("Unsupported command: %s\n", os.Args[1])
	}

	mediaURL := os.Args[2]

	var player omnicast.MediaPlayer
	player, err := gcastPlayer()
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
