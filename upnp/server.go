package upnp

import (
	"net/http"
	"net/url"

	"golang.org/x/sync/errgroup"

	"github.com/ericyan/omnicast/upnp/internal/ssdp"
)

type Server struct {
	ss *ssdp.Server
	hs *http.Server
}

func NewServer(dev *Device, addr string) (*Server, error) {
	loc := &url.URL{Scheme: "http", Host: addr, Path: "/"}
	ss, err := ssdp.NewServer(dev, loc)
	if err != nil {
		return nil, err
	}

	hs := &http.Server{
		Addr:    addr,
		Handler: dev,
	}

	return &Server{ss, hs}, nil
}

func (srv *Server) ListenAndServe() error {
	var g errgroup.Group

	g.Go(srv.ss.ListenAndServe)
	g.Go(srv.hs.ListenAndServe)

	return g.Wait()
}

func (srv *Server) Close() error {
	var g errgroup.Group

	g.Go(srv.hs.Close)
	g.Go(srv.ss.Close)

	return g.Wait()
}
