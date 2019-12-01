package av

import (
	"log"
	"net/url"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/upnp"
	"github.com/ericyan/omnicast/upnp/internal/soap"
)

var ErrInvalidInstanceID = &soap.Error{719, "Invalid InstanceID"}

// AVTransport returns an AVTransport UPnP service for the Player.
//
// Spec: http://upnp.org/specs/av/UPnP-av-AVTransport-v1-Service.pdf
func AVTransport(player omnicast.Player) *upnp.Service {
	svc := upnp.NewService("AVTransport", 1)

	svc.RegisterAction("SetAVTransportURI", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		if uri, ok := req.Args["CurrentURI"]; ok {
			media, err := url.Parse(uri)
			if err != nil {
				log.Println(err)

				resp.Error = soap.ErrInvalidArgs
				return
			}

			player.Load(media)
		}
	})

	svc.RegisterAction("Play", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		player.Play()
	})

	svc.RegisterAction("Pause", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		player.Pause()
	})

	return svc
}
