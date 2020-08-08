package gcast

import (
	"encoding/json"
	"log"
	"time"

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

// MediaSession represents the current status of a single session.
type MediaSession struct {
	MediaSessionID         int               `json:"mediaSessionId"`
	Media                  *MediaInformation `json:"media,omitempty"`
	PlaybackRate           float32           `json:"playbackRate"`
	PlayerState            string            `json:"playerState"`
	IdleReason             string            `json:"idleReason,omitempty"`
	CurrentTime            float64           `json:"currentTime"`
	SupportedMediaCommands int               `json:"supportedMediaCommands"`
}

// MediaStatus represents the current status of the media artifact with
// respect to the session.
//
// https://developers.google.com/cast/docs/reference/messages#MediaStatus
type MediaStatus struct {
	Status []*MediaSession `json:"status"`
}

// Receiver represents a Google Cast device.
type Receiver struct {
	*DeviceInfo

	ch     *castv2.Channel
	events chan *castv2.Msg

	app        *ReceiverApplication
	vol        *ReceiverVolume
	session    *MediaSession
	lastUpdate time.Time
}

func (r *Receiver) updateReceiverStatus(msg *castv2.Msg) error {
	rs := new(ReceiverStatus)
	if err := json.Unmarshal([]byte(msg.Payload), &rs); err != nil {
		return err
	}

	var app *ReceiverApplication
	if apps := rs.Status.Applications; len(apps) > 0 {
		app = apps[0]
	} else {
		app = nil
	}

	if r.app != app {
		r.app = app
		r.session = nil
	}

	r.vol = rs.Status.Volume

	return nil
}

func (r *Receiver) updateMediaStatus(msg *castv2.Msg) error {
	ms := new(MediaStatus)
	if err := json.Unmarshal([]byte(msg.Payload), &ms); err != nil {
		return err
	}

	r.lastUpdate = time.Now()
	for _, s := range ms.Status {
		// The media element will only be returned if it has changed.
		if s.Media == nil && r.session != nil {
			s.Media = r.session.Media
		}

		r.session = s
	}

	return nil
}

// Connect makes a connection to the receiver.
func (r *Receiver) Connect() error {
	if r.IsConnected() {
		return nil
	}

	if r.events == nil {
		r.events = make(chan *castv2.Msg)
		go func() {
			for msg := range r.events {
				var h castv2.Header
				if err := json.Unmarshal([]byte(msg.Payload), &h); err != nil {
					continue
				}

				switch h.Type {
				case castv2.TypeReceiverStatus:
					r.updateReceiverStatus(msg)
				case castv2.TypeMediaStatus:
					r.updateMediaStatus(msg)
				}
			}
		}()
	}

	ch, err := castv2.Dial(r.TCPAddr())
	if err != nil {
		return err
	}
	r.ch = ch

	r.ch.Subscribe(r.events)

	// Request receiver status to update state
	respCh := make(chan *castv2.Msg)
	err = r.ch.Request(
		castv2.PlatformSenderID,
		castv2.PlatformReceiverID,
		castv2.NamespaceReceiver,
		castv2.NewRequest(castv2.TypeGetStatus),
		respCh,
	)
	if err != nil {
		return err
	}

	return r.updateReceiverStatus(<-respCh)
}

// IsConnected returns true if there is an active connection to the
// receiver device.
func (r *Receiver) IsConnected() bool {
	if r.ch == nil || r.ch.IsClosed() {
		return false
	}

	return true
}

// Application returns the current running receiver application, if any.
func (r *Receiver) Application() *ReceiverApplication {
	if !r.IsConnected() {
		return nil
	}

	return r.app
}

// Volume returns the receiver volume, if known.
func (r *Receiver) Volume() *ReceiverVolume {
	if !r.IsConnected() {
		return nil
	}

	return r.vol
}

// GetSession returns the last known status of the media session.
func (r *Receiver) Session(senderID string) (*MediaSession, time.Time) {
	if !r.IsConnected() {
		log.Println("gcast: connection lost, reconnecting...")
		if err := r.Connect(); err != nil {
			log.Println("gcast: failed to reconnect.", err)
			return nil, r.lastUpdate
		}
	}

	if r.app == nil || r.app.IsIdleScreen {
		return nil, r.lastUpdate
	}

	if r.session == nil || time.Since(r.lastUpdate).Seconds() > 30 {
		respCh := make(chan *castv2.Msg)
		err := r.ch.Request(
			senderID,
			r.app.SessionID,
			castv2.NamespaceMedia,
			castv2.NewRequest(castv2.TypeGetStatus),
			respCh,
		)
		if err != nil {
			log.Println("gcast: failed to update media status.", err)
			return nil, r.lastUpdate
		}

		r.updateMediaStatus(<-respCh)
	}

	return r.session, r.lastUpdate
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
	if !r.IsConnected() {
		return nil
	}

	return r.ch.Close()
}
