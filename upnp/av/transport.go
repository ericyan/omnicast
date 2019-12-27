package av

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/ericyan/omnicast"
	"github.com/ericyan/omnicast/upnp"
	"github.com/ericyan/omnicast/upnp/internal/soap"
	"github.com/ericyan/omnicast/upnp/internal/types"
)

// Action-specific errors defined in AVTransport:1 service spec.
var (
	ErrSeekModeNotSupported = &soap.Error{710, "Seek mode not supported"}
	ErrIllegalSeekTarget    = &soap.Error{711, "Illegal seek target"}
	ErrInvalidInstanceID    = &soap.Error{719, "Invalid InstanceID"}
)

// AVTransport returns an AVTransport UPnP service for the Player.
//
// Spec: http://upnp.org/specs/av/UPnP-av-AVTransport-v1-Service.pdf
func AVTransport(player omnicast.MediaPlayer) *upnp.Service {
	svc := upnp.NewService("AVTransport", 1)

	svc.RegisterAction("SetAVTransportURI", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}
		if _, ok := req.Args["CurrentURI"]; !ok {
			resp.Error = soap.ErrInvalidArgs
			return
		}

		mediaURL, err := url.Parse(req.Args["CurrentURI"])
		if err != nil {
			log.Println(err)

			resp.Error = soap.ErrInvalidArgs
			return
		}

		mediaMetadata := new(types.Metadata)
		if didl, ok := req.Args["CurrentURIMetaData"]; ok && didl != "" {
			if err := mediaMetadata.UnmarshalText([]byte(didl)); err != nil {
				log.Println("parsing metadata failed:", err)
			}
		}

		player.Load(mediaURL, mediaMetadata)
	})

	svc.RegisterAction("GetMediaInfo", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		if !player.IsIdle() {
			resp.Args["NrTracks"] = "1"
			resp.Args["MediaDuration"] = types.FormatDuration(player.MediaDuration())
			resp.Args["CurrentURI"] = player.MediaURL().String()
		} else {
			resp.Args["NrTracks"] = "0"
			resp.Args["MediaDuration"] = "00:00:00"
			resp.Args["CurrentURI"] = ""
		}

		resp.Args["CurrentURIMetadata"] = "NOT_IMPLEMENTED"
		resp.Args["NextURI"] = "NOT_IMPLEMENTED"
		resp.Args["NextURIMetadata"] = "NOT_IMPLEMENTED"
		resp.Args["PlayMedium"] = "UNKNOWN"
		resp.Args["RecordMedium"] = "NOT_IMPLEMENTED"
		resp.Args["WriteStatus"] = "NOT_IMPLEMENTED"
	})

	svc.RegisterAction("GetTransportInfo", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		var state string
		switch {
		case player.IsPlaying():
			state = "PLAYING"
		case player.IsIdle():
			state = "NO_MEDIA_PRESENT"
		case player.IsPaused():
			state = "PAUSED_PLAYBACK"
		case player.IsBuffering():
			state = "TRANSITIONING"
		}

		resp.Args["CurrentTransportState"] = state
		resp.Args["CurrentTransportStatus"] = "OK"
		resp.Args["CurrentSpeed"] = types.ParseFloat32(player.PlaybackRate()).String()
	})

	svc.RegisterAction("GetPositionInfo", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		var pos time.Duration
		if !player.IsIdle() {
			pos = player.PlaybackPosition()

			resp.Args["Track"] = "1"
			resp.Args["TrackURI"] = player.MediaURL().String()
			resp.Args["TrackDuration"] = types.FormatDuration(player.MediaDuration())
		} else {
			pos = 0

			resp.Args["Track"] = "0"
			resp.Args["TrackURI"] = ""
			resp.Args["TrackDuration"] = "00:00:00"
		}

		resp.Args["TrackMetaData"] = "NOT_IMPLEMENTED"

		resp.Args["RelTime"] = types.FormatDuration(pos)
		resp.Args["AbsTime"] = types.FormatDuration(pos)
		resp.Args["RelCount"] = strconv.Itoa(int(pos.Seconds()))
		resp.Args["AbsCount"] = strconv.Itoa(int(pos.Seconds()))
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

	svc.RegisterAction("Stop", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		player.Stop()
	})

	svc.RegisterAction("Seek", func(req *soap.Request, resp *soap.Response) {
		if req.Args["InstanceID"] != "0" {
			resp.Error = ErrInvalidInstanceID
			return
		}

		switch req.Args["Unit"] {
		case "ABS_TIME", "REL_TIME":
			pos, err := types.ParseDuration(req.Args["Target"])
			if err != nil {
				resp.Error = ErrIllegalSeekTarget
				return
			}

			player.SeekTo(pos)
		default:
			resp.Error = ErrSeekModeNotSupported
			return
		}
	})

	return svc
}
