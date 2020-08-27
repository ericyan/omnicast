package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ericyan/iputil"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/gcast"
	"github.com/ericyan/omnicast/mpris"
	"github.com/ericyan/omnicast/upnp"
	"github.com/ericyan/omnicast/upnp/av"
)

var defaultHost = ""

func init() {
	if addr, _ := iputil.DefaultIPv4(); addr != nil {
		defaultHost = addr.IP.String()
	}
}

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

func mprisPlayer() (*mpris.Player, error) {
	dests, err := mpris.Discover()
	if err != nil {
		return nil, err
	}

	for _, dest := range dests {
		return mpris.NewPlayer(dest)
	}

	return nil, errors.New("no MPRIS player found")
}

func main() {
	host := flag.String("host", defaultHost, "host")
	port := flag.Int("p", 2278, "port")
	h := flag.Bool("h", false, "show help")
	flag.Parse()

	if *h {
		flag.Usage()
		os.Exit(0)
	}

	log.Printf("Listening on %s:%d...", *host, *port)

	var player omnicast.MediaPlayer

	player, err := gcastPlayer()
	if err != nil {
		log.Println(err)

		// Fallback to MPRIS
		player, err = mprisPlayer()
		if err != nil {
			log.Fatalln(err)
		}
	}

	dev, err := av.NewMediaRenderer(player.Name()+" (DLNA)", player)
	if err != nil {
		log.Fatalln(err)
	}

	addr := *host + ":" + strconv.Itoa(*port)

	srv, err := upnp.NewServer(dev, addr)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal %s received, stopping server...\n", s)
	srv.Close()
}
