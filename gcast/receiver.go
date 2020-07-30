package gcast

import (
	"encoding/json"

	"github.com/ericyan/omnicast/gcast/internal/castv2"
)

// Common receiver app IDs.
const (
	DefaultReceiverAppID = "CC1AD845"
	YouTubeReceiverAppID = "233637DE"
)

// ReceiverApplication represents an instance of receiver application.
type ReceiverApplication struct {
	AppID               string              `json:"appId"`
	Name                string              `json:"displayName"`
	IconURL             string              `json:"iconUrl"`
	StatusText          string              `json:"statusText"`
	IsIdleScreen        bool                `json:"isIdleScreen"`
	SupportedNamespaces []map[string]string `json:"namespaces"`
	SessionID           string              `json:"sessionId"`
	TransportID         string              `json:"transportId"`
}

// ReceiverVolume represents the volume of the receiver device.
type ReceiverVolume struct {
	ControlType  string  `json:"controlType,omitempty"`
	Level        float64 `json:"level,omitempty"`
	Muted        bool    `json:"muted,omitempty"`
	StepInterval float64 `json:"stepInterval,omitempty"`
}

// ReceiverStatus represents the devices status of the receiver.
type ReceiverStatus struct {
	castv2.Header
	Status struct {
		Applications []*ReceiverApplication `json:"applications,omitempty"`
		Volume       *ReceiverVolume        `json:"volume"`
	} `json:"status"`
}

// MediaInformation represents a media stream.
//
// Ref: https://developers.google.com/cast/docs/reference/messages#MediaInformation
type MediaInformation struct {
	ContentID   string        `json:"contentId"`
	ContentType string        `json:"contentType"`
	StreamType  string        `json:"streamType"`
	Metadata    MediaMetadata `json:"metadata,omitempty"`
	Duration    float64       `json:"duration,omitempty"`
}

// MediaStatus represents the current status of the media artifact with
// respect to the session.
//
// https://developers.google.com/cast/docs/reference/messages#MediaStatus
type MediaStatus struct {
	Status []struct {
		MediaSessionID         int               `json:"mediaSessionId"`
		Media                  *MediaInformation `json:"media,omitempty"`
		PlaybackRate           float32           `json:"playbackRate"`
		PlayerState            string            `json:"playerState"`
		IdleReason             string            `json:"idleReason,omitempty"`
		CurrentTime            float64           `json:"currentTime"`
		SupportedMediaCommands int               `json:"supportedMediaCommands"`
	} `json:"status"`
}

// Receiver represents a Google Cast device.
type Receiver struct {
	Addr string

	ch *castv2.Channel

	rsSubID     int
	rsListeners []func(*ReceiverStatus)
	msSubID     int
	msListeners []func(*MediaStatus)

	app *ReceiverApplication
	vol *ReceiverVolume
}

// NewReceiver returns a new Receiver instance.
func NewReceiver(addr string) *Receiver {
	return &Receiver{Addr: addr}
}

// Connect makes a connection to the receiver.
func (r *Receiver) Connect() error {
	if r.ch != nil && !r.ch.IsClosed() {
		return nil
	}

	ch, err := castv2.NewChannel(r.Addr)
	if err != nil {
		return err
	}
	r.ch = ch

	rsSubCh := make(chan *castv2.Msg)
	r.rsSubID, err = r.ch.Subscribe(castv2.TypeReceiverStatus, rsSubCh)
	if err != nil {
		return err
	}

	go func() {
		for msg := range rsSubCh {
			rs := new(ReceiverStatus)
			if err := json.Unmarshal([]byte(msg.Payload), &rs); err != nil {
				continue
			}

			if apps := rs.Status.Applications; len(apps) > 0 {
				r.app = apps[0]
			} else {
				r.app = nil
			}

			r.vol = rs.Status.Volume

			for _, f := range r.rsListeners {
				f(rs)
			}
		}
	}()

	// Request receiver status for the initial state
	r.ch.Request(
		castv2.PlatformSenderID,
		castv2.PlatformReceiverID,
		castv2.NamespaceReceiver,
		castv2.NewRequest(castv2.TypeGetStatus),
		rsSubCh,
	)

	msSubCh := make(chan *castv2.Msg)
	r.msSubID, err = r.ch.Subscribe(castv2.TypeMediaStatus, msSubCh)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msSubCh {
			ms, err := parseMediaStatus(msg)
			if err != nil {
				continue
			}

			for _, f := range r.msListeners {
				f(ms)
			}
		}
	}()

	return nil
}

// Application returns the current running receiver application, if any.
func (r *Receiver) Application() *ReceiverApplication {
	return r.app
}

// Volume returns the receiver volume, if known.
func (r *Receiver) Volume() *ReceiverVolume {
	return r.vol
}

// OnStatusUpdate registers an event listener for status updates.
func (r *Receiver) OnStatusUpdate(listener func(*ReceiverStatus)) {
	r.rsListeners = append(r.rsListeners, listener)
}

func parseMediaStatus(msg *castv2.Msg) (*MediaStatus, error) {
	var ms MediaStatus
	if err := json.Unmarshal([]byte(msg.Payload), &ms); err != nil {
		return nil, err
	}

	return &ms, nil
}

// GetMediaStatus retrieves the media status for all media sessions.
func (r *Receiver) GetMediaStatus(senderID string) (*MediaStatus, error) {
	respCh := make(chan *castv2.Msg)
	err := r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		castv2.NewRequest(castv2.TypeGetStatus),
		respCh,
	)
	if err != nil {
		return nil, err
	}
	resp := <-respCh

	return parseMediaStatus(resp)
}

// OnMediaStatusUpdate registers an event listener for media status
// updates.
func (r *Receiver) OnMediaStatusUpdate(listener func(*MediaStatus)) {
	r.msListeners = append(r.msListeners, listener)
}

// Launch starts an new receiver application.
func (r *Receiver) Launch(appID string) error {
	req := &struct {
		castv2.Header
		AppID string `json:"appId"`
	}{}

	req.Type = castv2.TypeLaunch
	req.AppID = appID

	return r.ch.Request(
		castv2.PlatformSenderID,
		castv2.PlatformReceiverID,
		castv2.NamespaceReceiver,
		req,
		nil,
	)
}

// SetVolume sets the receiver volume.
func (r *Receiver) SetVolume(vol *ReceiverVolume) error {
	req := &struct {
		castv2.Header
		Volume *ReceiverVolume `json:"volume"`
	}{}

	req.Type = castv2.TypeSetVolume
	req.Volume = vol

	return r.ch.Request(
		castv2.PlatformSenderID,
		castv2.PlatformReceiverID,
		castv2.NamespaceReceiver,
		req,
		nil,
	)
}

// Load loads new content into the media player.
//
// Ref: https://developers.google.com/cast/docs/reference/messages#Load
func (r *Receiver) Load(senderID string, media *MediaInformation) error {
	req := &struct {
		castv2.Header
		Media *MediaInformation `json:"media"`
	}{}

	req.Type = castv2.TypeLoad
	req.Media = media

	return r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		req,
		nil,
	)
}

// Play begins playback of the loaded media content from the current
// playback position.
//
// https://developers.google.com/cast/docs/reference/messages#Play
func (r *Receiver) Play(senderID string, mediaSessionID int) error {
	req := &struct {
		castv2.Header
		MediaSessionID int `json:"mediaSessionId"`
	}{}

	req.Type = castv2.TypePlay
	req.MediaSessionID = mediaSessionID

	return r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		req,
		nil,
	)
}

// Pause pauses playback of the current content.
//
// https://developers.google.com/cast/docs/reference/messages#Pause
func (r *Receiver) Pause(senderID string, mediaSessionID int) error {
	req := &struct {
		castv2.Header
		MediaSessionID int `json:"mediaSessionId"`
	}{}

	req.Type = castv2.TypePause
	req.MediaSessionID = mediaSessionID

	return r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		req,
		nil,
	)
}

// Stop stops the playback and unload the current content
//
// https://developers.google.com/cast/docs/reference/messages#Stop
func (r *Receiver) Stop(senderID string, mediaSessionID int) error {
	req := &struct {
		castv2.Header
		MediaSessionID int `json:"mediaSessionId"`
	}{}

	req.Type = castv2.TypeStop
	req.MediaSessionID = mediaSessionID

	return r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		req,
		nil,
	)
}

// Seek sets the current playback position to pos, which is the number
// of seconds since beginning of content
func (r *Receiver) Seek(senderID string, mediaSessionID int, pos float64) error {
	req := &struct {
		castv2.Header
		MediaSessionID int     `json:"mediaSessionId"`
		CurrentTime    float64 `json:"currentTime"`
	}{}

	req.Type = castv2.TypeSeek
	req.MediaSessionID = mediaSessionID
	req.CurrentTime = pos

	return r.ch.Request(
		senderID,
		r.app.SessionID,
		castv2.NamespaceMedia,
		req,
		nil,
	)
}

// Close closes the connection to the receiver.
func (r *Receiver) Close() error {
	r.ch.Unsubscribe(castv2.TypeReceiverStatus, r.rsSubID)
	r.ch.Unsubscribe(castv2.TypeMediaStatus, r.msSubID)

	return r.ch.Close()
}
