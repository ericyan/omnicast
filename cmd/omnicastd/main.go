package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ericyan/iputil"

	"github.com/ericyan/omnicast/gcast"
	"github.com/ericyan/omnicast/upnp"
	"github.com/ericyan/omnicast/upnp/av"
)

var defaultHost = ""

func init() {
	if addr, _ := iputil.DefaultIPv4(); addr != nil {
		defaultHost = addr.IP.String()
	}
}

func main() {
	host := flag.String("host", defaultHost, "host")
	port := flag.Int("p", 2278, "port")
	cast := flag.String("cast", "", "address to cast device")
	h := flag.Bool("h", false, "show help")
	flag.Parse()

	if *h {
		flag.Usage()
		os.Exit(0)
	}

	r, err := gcast.NewReceiver(*cast)
	if err != nil {
		log.Fatalln(err)
	}

	player, err := gcast.NewSender("sender-omnicast", r)
	if err != nil {
		log.Fatalln(err)
	}
	dev, err := av.NewMediaRenderer(r.Name+" (DLNA)", player)
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
