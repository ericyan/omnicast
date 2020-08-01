package gcast

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

type DeviceInfo struct {
	Name string `json:"name"`
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
