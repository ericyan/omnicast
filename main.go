package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ericyan/iputil"

	"github.com/ericyan/omnicast/upnp"
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
	h := flag.Bool("h", false, "show help")
	flag.Parse()

	if *h {
		flag.Usage()
		os.Exit(0)
	}

	dev := upnp.NewDevice("Basic Device", "Basic", 1)
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
