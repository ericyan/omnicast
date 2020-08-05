package gcast

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/grandcat/zeroconf"
)

type DeviceInfo struct {
	Name  string `json:"name"`
	Model string

	IPv4 net.IP
	IPv6 net.IP
	Port int
}

func GetDeviceInfo(ip net.IP) (*DeviceInfo, error) {
	host := &net.TCPAddr{IP: ip, Port: 8008}
	endpoint := &url.URL{
		Scheme: "http",
		Host:   host.String(),
		Path:   "/setup/eureka_info",
	}

	resp, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info DeviceInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Discover returns a channel with DeviceInfo found via mDNS.
func Discover(ctx context.Context) (<-chan *DeviceInfo, error) {
	resolv, err := zeroconf.NewResolver()
	if err != nil {
		return nil, err
	}

	mdnsCh := make(chan *zeroconf.ServiceEntry)
	go func() {
		if err := resolv.Browse(ctx, "_googlecast._tcp", "local", mdnsCh); err != nil {
			return
		}
	}()

	devCh := make(chan *DeviceInfo)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(devCh)
				return
			case mdns := <-mdnsCh:
				if mdns == nil {
					continue
				}

				dev := new(DeviceInfo)

				if len(mdns.AddrIPv4) > 0 {
					dev.IPv4 = mdns.AddrIPv4[0]
				}
				if len(mdns.AddrIPv6) > 0 {
					dev.IPv6 = mdns.AddrIPv6[0]
				}

				dev.Port = mdns.Port

				for _, value := range mdns.Text {
					if kv := strings.SplitN(value, "=", 2); len(kv) == 2 {
						key, val := kv[0], kv[1]

						switch key {
						case "fn":
							dev.Name = val
						case "md":
							dev.Model = val
						}
					}
				}

				devCh <- dev
			}
		}
	}()

	return devCh, nil
}
