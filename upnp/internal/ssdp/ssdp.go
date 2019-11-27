package ssdp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"
)

const (
	MulticastIPv4Addr     = "239.255.255.250:1900"
	MTU                   = 8192
	AliveInterval         = 15 * time.Minute
	CacheControlDirective = "max-age=1800"
	ServerName            = runtime.GOARCH + "/" + runtime.GOARCH + " UPnP/2.0 omnicast/0.1"
)

type Device interface {
	// Returns the Unique Device Name, which will be the prefix of the USN
	// header field in all discovery messages.
	UDN() string
	// Returns the URN of the device.
	URN() string
	// Returns the URNs of all services provided by the device.
	ServiceURNs() []string
}

type Server struct {
	dev   Device
	loc   *url.URL
	addr  *net.UDPAddr
	conn  *net.UDPConn
	alive *time.Ticker
	done  chan bool
}

// NewServer returns a SSDP server for the given device that announces
// the URL to its UPnP description.
func NewServer(dev Device, loc *url.URL) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", MulticastIPv4Addr)
	if err != nil {
		return nil, err
	}

	return &Server{dev: dev, loc: loc, addr: addr}, nil
}

func (srv *Server) ListenAndServe() error {
	conn, err := net.ListenMulticastUDP("udp", nil, srv.addr)
	if err != nil {
		return err
	}
	conn.SetReadBuffer(MTU)
	srv.conn = conn

	log.Printf("SSDP server listening on: %s", srv.conn.LocalAddr())

	srv.done = make(chan bool)

	srv.sendNotification("ssdp:alive")
	if AliveInterval > 0 {
		srv.alive = time.NewTicker(AliveInterval)
		go func() {
			for range srv.alive.C {
				srv.sendNotification("ssdp:alive")
			}
		}()
	}

	buf := make([]byte, MTU)
	for {
		select {
		case <-srv.done:
			close(srv.done)
			return nil
		default:
			n, raddr, err := srv.conn.ReadFromUDP(buf)
			if err != nil {
				if err, ok := err.(net.Error); ok && (err.Timeout() || err.Temporary()) {
					continue
				}

				return err
			}

			req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf[:n])))
			if err != nil {
				log.Println("failed to parse request:", err)
				continue
			}

			err = srv.handleRequest(req, raddr)
			if err != nil {
				log.Println("failed to handle request:", err)
			}
		}
	}
}

func (srv *Server) Close() error {
	if srv.conn == nil {
		return nil
	}

	if srv.alive != nil {
		srv.alive.Stop()
	}

	srv.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	srv.done <- true
	<-srv.done

	srv.sendNotification("ssdp:byebye")

	return srv.conn.Close()
}

func (srv *Server) capabilities() map[string]string {
	caps := map[string]string{srv.dev.UDN(): srv.dev.UDN()}
	for _, urn := range append([]string{"upnp:rootdevice", srv.dev.URN()}, srv.dev.ServiceURNs()...) {
		caps[urn] = srv.dev.UDN() + "::" + urn
	}

	return caps
}

func (srv *Server) commonHeader() http.Header {
	return http.Header{
		"CACHE-CONTROL":     []string{CacheControlDirective},
		"LOCATION":          []string{srv.loc.String()},
		"SERVER":            []string{ServerName},
		"BOOTID.UPNP.ORG":   []string{strconv.Itoa(int(time.Now().Unix()))},
		"CONFIGID.UPNP.ORG": []string{"1"},
	}
}

func (srv *Server) sendNotification(nts string) error {
	switch nts {
	case "ssdp:alive", "ssdp:byebye":
	case "ssdp:update":
		return fmt.Errorf("NTS %s not implemented", nts)
	default:
		return fmt.Errorf("invalid NTS: %s", nts)
	}

	for t, usn := range srv.capabilities() {
		req := &http.Request{
			Method: "NOTIFY",
			URL:    &url.URL{Opaque: "*"},
			Host:   MulticastIPv4Addr,
			Header: srv.commonHeader(),
		}

		req.Header.Set("NTS", nts)

		req.Header.Set("NT", t)
		req.Header.Set("USN", usn)

		buf := new(bytes.Buffer)
		req.Write(buf)

		_, err := srv.conn.WriteTo(buf.Bytes(), srv.addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (srv *Server) handleRequest(req *http.Request, raddr *net.UDPAddr) error {
	if req.Method != "M-SEARCH" {
		return fmt.Errorf("unsupported method: %s", req.Method)
	}

	if man := req.Header.Get("MAN"); man != `"ssdp:discover"` {
		return fmt.Errorf("unexpected MAN: %s", man)
	}

	st := req.Header.Get("ST")
	if st == "" {
		return errors.New("ST is empty")
	}

	log.Printf("Received %s request from %s, ST=%s\n", req.Method, raddr, st)

	n := 0
	for t, usn := range srv.capabilities() {
		if st == t || st == "ssdp:all" {
			resp := &http.Response{
				StatusCode:    http.StatusOK,
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        srv.commonHeader(),
				ContentLength: -1,
				Uncompressed:  true,
			}

			resp.Header.Set("ST", t)
			resp.Header.Set("USN", usn)

			buf := new(bytes.Buffer)
			resp.Write(buf)

			_, err := srv.conn.WriteTo(buf.Bytes(), raddr)
			if err != nil {
				return err
			}

			n++
		}
	}

	if n == 0 {
		return fmt.Errorf("ST %s not found", st)
	}

	return nil
}
