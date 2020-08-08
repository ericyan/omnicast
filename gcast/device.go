package gcast

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/grandcat/zeroconf"
)

// DeviceCapability represents one of the defined device capabilities.
type DeviceCapability uint

// String returns the string representation of the device capability.
func (c DeviceCapability) String() string {
	switch c {
	case None:
		return "none"
	case VideoOut:
		return "video_out"
	case VideoIn:
		return "video_in"
	case AudioOut:
		return "audio_out"
	case AudioIn:
		return "audio_in"
	case DevMode:
		return "dev_mode"
	case MultizoneGroup:
		return "multizone_group"
	default:
		return strconv.Itoa(int(c))
	}
}

// Defined Google Cast device capabilities.
//
// Source: https://github.com/chromium/chromium/blob/master/components/cast_channel/cast_socket.h#L46
const (
	None           DeviceCapability = 0
	VideoOut       DeviceCapability = 1 << 0
	VideoIn        DeviceCapability = 1 << 1
	AudioOut       DeviceCapability = 1 << 2
	AudioIn        DeviceCapability = 1 << 3
	DevMode        DeviceCapability = 1 << 4
	MultizoneGroup DeviceCapability = 1 << 5
)

type DeviceInfo struct {
	UUID  uuid.UUID
	Name  string `json:"name"`
	Model string

	IPv4 net.IP
	IPv6 net.IP
	Port int

	capabilities DeviceCapability
}

// TCPAddr returns IPv4 and Port as net.TCPAddr.
func (d *DeviceInfo) TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{IP: d.IPv4, Port: d.Port}
}

// Capabilities returns a list of device capabilities.
func (d *DeviceInfo) Capabilities() []DeviceCapability {
	result := make([]DeviceCapability, 0)

	if d.capabilities&VideoOut != 0 {
		result = append(result, VideoOut)
	}

	if d.capabilities&VideoIn != 0 {
		result = append(result, VideoIn)
	}

	if d.capabilities&AudioOut != 0 {
		result = append(result, AudioOut)
	}

	if d.capabilities&AudioIn != 0 {
		result = append(result, AudioIn)
	}

	if d.capabilities&DevMode != 0 {
		result = append(result, DevMode)
	}

	if d.capabilities&MultizoneGroup != 0 {
		result = append(result, MultizoneGroup)
	}

	return result
}

// CapableOf returns true if the device has all given capabilities.
func (d *DeviceInfo) CapableOf(capabilities ...DeviceCapability) bool {
	var mask DeviceCapability
	for _, c := range capabilities {
		mask |= c
	}

	return d.capabilities&mask == mask
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
						case "id":
							dev.UUID, _ = uuid.Parse(val)
						case "fn":
							dev.Name = val
						case "md":
							dev.Model = val
						case "ca":
							ca, err := strconv.Atoi(val)
							if err != nil {
								dev.capabilities = None
							}

							dev.capabilities = DeviceCapability(ca)
						}
					}
				}

				devCh <- dev
			}
		}
	}()

	return devCh, nil
}
