package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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

func findPlayer(gcastHint, mprisHint string) (omnicast.MediaPlayer, error) {
	if mprisHint != "" {
		return mpris.NewPlayer(mprisHint)
	}

	if gcastHint != "" {
		return gcast.Find(gcastHint)
	}

	return gcast.Find()
}

func main() {
	host := flag.String("host", defaultHost, "host")
	port := flag.Int("p", 2278, "port")
	gcastHint := flag.String("gcast", "", "Google Cast device name or UUID")
	mprisHint := flag.String("mpris", "", "MPRIS destination")
	h := flag.Bool("h", false, "show help")
	flag.Parse()

	if *h {
		flag.Usage()
		os.Exit(0)
	}

	log.Printf("Listening on %s:%d...", *host, *port)

	player, err := findPlayer(*gcastHint, *mprisHint)
	if err != nil {
		log.Fatalln(err)
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
