package av

import (
	"strconv"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/upnp"
	"github.com/ericyan/omnicast/upnp/internal/soap"
)

// RenderingControl returns a RenderingControl UPnP service for the Player.
//
// Spec: http://upnp.org/specs/av/UPnP-av-RenderingControl-v1-Service.pdf
func RenderingControl(player omnicast.MediaPlayer) *upnp.Service {
	svc := upnp.NewService("RenderingControl", 1)

	var (
		ErrInvalidInstanceID = &soap.Error{702, "Invalid InstanceID"}
	)

	svc.RegisterAction("GetVolume", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}
		if req.Args["Channel"] != "Master" {
			resp.Error = soap.ErrInvalidArgs
			return
		}

		vol := int(player.VolumeLevel() * 100)
		resp.Args["CurrentVolume"] = strconv.Itoa(vol)
	})

	svc.RegisterAction("SetVolume", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}
		if req.Args["Channel"] != "Master" {
			resp.Error = soap.ErrInvalidArgs
			return
		}

		vol, err := strconv.Atoi(req.Args["DesiredVolume"])
		if err != nil {
			resp.Error = soap.ErrInvalidArgs
			return
		}

		player.SetVolumeLevel(float64(vol) / 100.0)
	})

	return svc
}
