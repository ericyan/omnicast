package gcast

import (
	"errors"
	"mime"
	"net/url"
	"path/filepath"
	"time"
)

// Errors used by the Sender.
var (
	ErrReceiverNotReady = errors.New("receiver not ready")
	ErrInvalidMedia     = errors.New("invalid media")
)

// A Sender is a sender app instance that controls media playback on the
// receiver. Its ID, which should be unique, is used to identify itself
// when communicating with the receiver.
type Sender struct {
	ID string

	ReceiverApp    *ReceiverApplication
	ReceiverVolume *ReceiverVolume

	MediaSessionID int
	MediaInfo      *MediaInformation

	r *Receiver

	playbackState    string
	playbackPosition float64
	playbackRate     float32
}

// NewSender returns a new Sender with given ID.
func NewSender(id string) *Sender {
	return &Sender{ID: id}
}

// ConnectTo makes a connection to the Receiver.
func (s *Sender) ConnectTo(raddr string) error {
	s.r = NewReceiver(raddr)

	err := s.r.Connect()
	if err != nil {
		return err
	}

	s.r.OnStatusUpdate(s.updateReceiverStatus)
	s.r.OnMediaStatusUpdate(s.updateMediaStatus)

	s.getReceiverStatus()
	s.getMediaStatus()

	return nil
}

func (s *Sender) updateReceiverStatus(rs *ReceiverStatus) {
	if apps := rs.Status.Applications; len(apps) > 0 {
		s.ReceiverApp = apps[0]
	} else {
		s.ReceiverApp = nil
	}

	s.ReceiverVolume = rs.Status.Volume
}

func (s *Sender) getReceiverStatus() error {
	rs, err := s.r.GetStatus()
	if err != nil {
		return err
	}

	s.updateReceiverStatus(rs)

	return nil
}

func (s *Sender) updateMediaStatus(ms *MediaStatus) {
	sess := ms.Status[0]
	s.MediaSessionID = sess.MediaSessionID
	s.MediaInfo = sess.Media

	s.playbackState = sess.PlayerState
	s.playbackPosition = sess.CurrentTime
	s.playbackRate = sess.PlaybackRate
}

func (s *Sender) getMediaStatus() error {
	if s.ReceiverApp == nil || s.ReceiverApp.IsIdleScreen {
		return ErrReceiverNotReady
	}

	ms, err := s.r.GetMediaStatus(s.ID, s.ReceiverApp.SessionID)
	if err != nil {
		return err
	}
	if len(ms.Status) == 0 {
		return ErrReceiverNotReady
	}

	s.updateMediaStatus(ms)

	return nil
}

func (s *Sender) ensureAppLaunched(appID string) error {
	for {
		select {
		case <-time.After(2 * time.Second):
			return ErrReceiverNotReady
		default:
			if s.ReceiverApp != nil && s.ReceiverApp.AppID == DefaultReceiverAppID {
				return nil
			}
		}
	}
}

// Load casts media to the receiver and starts playback.
func (s *Sender) Load(mediaURL *url.URL) error {
	if !mediaURL.IsAbs() {
		return ErrInvalidMedia
	}

	ext := filepath.Ext(mediaURL.EscapedPath())
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if s.ReceiverApp == nil || s.ReceiverApp.AppID != DefaultReceiverAppID {
		err := s.r.Launch(DefaultReceiverAppID)
		if err != nil {
			return err
		}
	}
	if err := s.ensureAppLaunched(DefaultReceiverAppID); err != nil {
		return err
	}

	mediaInfo := &MediaInformation{
		ContentID:   mediaURL.String(),
		ContentType: contentType,
		StreamType:  "BUFFERED",
	}

	return s.r.Load(s.ID, s.ReceiverApp.SessionID, mediaInfo)
}

// IsIdle returns true if the recevier has media playback stopped.
func (s *Sender) IsIdle() bool {
	return s.playbackState == "IDLE"
}

// IsPlaying returns true if the recevier is actively playing content.
func (s *Sender) IsPlaying() bool {
	return s.playbackState == "PLAYING"
}

// IsPaused returns true if playback is paused due to user request.
func (s *Sender) IsPaused() bool {
	return s.playbackState == "PAUSED"
}

// IsBuffering returns true if playback is effectively paused due to
// buffer underflow.
func (s *Sender) IsBuffering() bool {
	return s.playbackState == "BUFFERING"
}

// PlaybackPosition returns the current position of media playback from
// the beginning of media content. For live streams, it returns the time
// since playback started.
func (s *Sender) PlaybackPosition() time.Duration {
	return time.Duration(s.playbackPosition * float64(time.Second))
}

// PlaybackRate returns the ratio of speed that media is played at.
func (s *Sender) PlaybackRate() float32 {
	return s.playbackRate
}

// Close closes the connected receiver, if any.
func (s *Sender) Close() error {
	if s.r == nil {
		return nil
	}

	return s.r.Close()
}
