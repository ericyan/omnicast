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
	"time"

	"github.com/rakyll/statik/fs"

	_ "github.com/ericyan/omnicast/upnp/internal/scpd"
	"github.com/ericyan/omnicast/upnp/internal/soap"
)

type Device struct {
	Name    string
	Type    string
	Version uint

	uuid     [16]byte
	services map[string]*Service
}

func NewDevice(name, deviceType string, ver uint) *Device {
	return &Device{
		Name:     name,
		Type:     deviceType,
		Version:  ver,
		uuid:     md5.Sum([]byte(name + deviceType)),
		services: make(map[string]*Service),
	}
}

func (dev *Device) RegisterService(svc *Service) {
	if svc != nil {
		dev.services[svc.Type] = svc
	}
}

func (dev *Device) UDN() string {
	var buf [5 + 36]byte
	copy(buf[:], "uuid:")
	hex.Encode(buf[5:], dev.uuid[:4])
	buf[13] = '-'
	hex.Encode(buf[14:18], dev.uuid[4:6])
	buf[18] = '-'
	hex.Encode(buf[19:23], dev.uuid[6:8])
	buf[23] = '-'
	hex.Encode(buf[24:28], dev.uuid[8:10])
	buf[28] = '-'
	hex.Encode(buf[29:], dev.uuid[10:])

	return string(buf[:])
}

func (dev *Device) URN() string {
	return "urn:schemas-upnp-org:device:" + dev.Type + ":" + strconv.Itoa(int(dev.Version))
}

func (dev *Device) Services() map[string]*Service {
	return dev.services
}

func (dev *Device) ServiceURNs() []string {
	urns := make([]string, 0, len(dev.services))
	for _, svc := range dev.services {
		urns = append(urns, svc.URN())
	}

	return urns
}

func (dev *Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

	if r.URL.Path == "/" {
		w.Header().Add("Content-Type", "application/xml")
		dev.writeDevice(w)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/services/") {
		st := r.URL.Path[len("/services/"):len(r.URL.Path)]

		svc, ok := dev.services[st]
		if !ok {
			log.Printf("Service %s not found\n", st)
			http.NotFound(w, r)
			return
		}

		if r.Method == http.MethodGet {
			scpd, err := fs.New()
			if err != nil {
				log.Println(err)
				http.NotFound(w, r)
				return
			}

			filename := "/" + st + ".xml"
			f, err := scpd.Open(filename)
			if err != nil {
				log.Println(err)
				http.NotFound(w, r)
				return
			}

			http.ServeContent(w, r, filename, time.Time{}, f)
			return
		}

		if r.Method == http.MethodPost {
			req, err := soap.ParseHTTPRequest(r)
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
    <deviceType>{{.URN}}</deviceType>
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

	actions map[string]func(*soap.Request, *soap.Response)
}

func NewService(serviceType string, ver uint) *Service {
	return &Service{
		Type:    serviceType,
		Version: ver,
		actions: make(map[string]func(*soap.Request, *soap.Response)),
	}
}

func (svc *Service) URN() string {
	return "urn:schemas-upnp-org:service:" + svc.Type + ":" + strconv.Itoa(int(svc.Version))
}

func (svc *Service) RegisterAction(name string, handler func(*soap.Request, *soap.Response)) {
	svc.actions[name] = handler
}

func (svc *Service) HandleRequest(req *soap.Request) *soap.Response {
	resp := new(soap.Response)
	resp.Action = req.Action
	resp.Args = make(map[string]string)

	handler, ok := svc.actions[req.Action.Name]
	if !ok {
		resp.Error = soap.ErrActionNotImplemented
		return resp
	}

	handler(req, resp)
	return resp
}
