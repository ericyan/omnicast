package av

import (
	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/upnp"
)

// NewMediaRenderer returns a MediaRenderer UPnP device.
//
// Spec: http://upnp.org/specs/av/UPnP-av-MediaRenderer-v1-Device.pdf
func NewMediaRenderer(name string, player omnicast.MediaPlayer) (*upnp.Device, error) {
	dev := upnp.NewDevice(name, "MediaRenderer", 1)

	dev.RegisterService(AVTransport(player))

	return dev, nil
}
