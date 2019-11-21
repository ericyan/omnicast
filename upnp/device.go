package upnp

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type Device struct {
	UUID [16]byte
	Name string
	Type string

	Services map[string]*Service
}

func NewDevice(data []byte) *Device {
	dev := new(Device)
	dev.UUID = md5.Sum(data)
	dev.Services = make(map[string]*Service)

	return dev
}

func (dev *Device) RegisterService(svc *Service) {
	if svc != nil {
		dev.Services[svc.Type] = svc
	}
}

func (dev *Device) UDN() string {
	var buf [5 + 36]byte
	copy(buf[:], "uuid:")
	hex.Encode(buf[5:], dev.UUID[:4])
	buf[13] = '-'
	hex.Encode(buf[14:18], dev.UUID[4:6])
	buf[18] = '-'
	hex.Encode(buf[19:23], dev.UUID[6:8])
	buf[23] = '-'
	hex.Encode(buf[24:28], dev.UUID[8:10])
	buf[28] = '-'
	hex.Encode(buf[29:], dev.UUID[10:])

	return string(buf[:])
}

func (dev *Device) URN() string {
	return dev.Type
}

func (dev *Device) ServiceURNs() []string {
	urns := make([]string, 0, len(dev.Services))
	for _, svc := range dev.Services {
		urns = append(urns, svc.URN())
	}

	return urns
}

func (dev *Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Add("Content-Type", "application/xml")
		dev.writeDevice(w)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/services/") {
		st := r.URL.Path[len("/services/"):len(r.URL.Path)]

		svc, ok := dev.Services[st]
		if !ok {
			log.Printf("Service %s not found\n", st)
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodPost {
			req, err := ParseHTTPRequest(r)
			if err != nil {
				log.Println(err)
				return
			}

			err = svc.HandleRequest(req).WriteTo(w)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
	}

	http.NotFound(w, r)
}

const deviceTemplate = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>0</minor>
  </specVersion>
  <device>
    <deviceType>{{.Type}}</deviceType>
    <UDN>{{.UDN}}</UDN>
    <friendlyName>{{.Name}}</friendlyName>
    <manufacturer>Eric Yan</manufacturer>
    <manufacturerURL>https://ericyan.me/</manufacturerURL>
    <modelName>Omnicast</modelName>
    <modelDescription>DLNA media renderer written in Go</modelDescription>
    <modelNumber>0.1</modelNumber>
    <modelURL>http://github.com/ericyan/omnicast</modelURL>
    <dlna:X_DLNADOC xmlns:dlna="urn:schemas-dlna-org:device-1-0">DMR-1.50</dlna:X_DLNADOC>
    <serviceList>
    {{- range $path, $svc := .Services }}
      <service>
        <serviceType>{{$svc.URN}}</serviceType>
        <serviceId>urn:upnp-org:serviceId:{{$svc.Type}}</serviceId>
        <controlURL>/services/{{$path}}</controlURL>
        <eventSubURL>/services/{{$path}}/events</eventSubURL>
        <SCPDURL>/services/{{$path}}</SCPDURL>
      </service>
    {{- end}}
    </serviceList>
  </device>
</root>`

func (dev *Device) writeDevice(w io.Writer) error {
	tpl := template.Must(template.New("device").Parse(deviceTemplate))

	return tpl.Execute(w, dev)
}

type Service struct {
	Type    string
	Version uint
	Actions map[string]func(*SOAPRequest, *SOAPResponse)
}

func (svc *Service) URN() string {
	return "urn:schemas-upnp-org:service:" + svc.Type + ":" + strconv.Itoa(int(svc.Version))
}

func (svc *Service) RegisterAction(name string, handler func(*SOAPRequest, *SOAPResponse)) {
	svc.Actions[name] = handler
}

func (svc *Service) HandleRequest(req *SOAPRequest) *SOAPResponse {
	resp := new(SOAPResponse)
	resp.Action = req.Action
	resp.Args = make(map[string]string)

	handler, ok := svc.Actions[req.Action.Name]
	if !ok {
		resp.Error = ErrActionNotImplemented
		return resp
	}

	handler(req, resp)
	return resp
}
