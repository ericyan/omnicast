package gcast

import (
	"errors"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/ericyan/omnicast"
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

	r *Receiver
}

// NewSender returns a new Sender and connects to the Receiver.
func NewSender(id string, r *Receiver) (*Sender, error) {
	s := &Sender{ID: id, r: r}

	if err := s.r.Connect(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sender) ensureAppLaunched(appID string) error {
	if s.r.Application() == nil || s.r.Application().AppID != DefaultReceiverAppID {
		err := s.r.Launch(DefaultReceiverAppID)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-time.After(2 * time.Second):
			return ErrReceiverNotReady
		default:
			if s.r.Application() != nil && s.r.Application().AppID == DefaultReceiverAppID {
				return nil
			}
		}
	}
}

// Load casts media to the receiver and starts playback.
func (s *Sender) Load(mediaURL *url.URL, mediaMetadata omnicast.MediaMetadata) error {
	if !mediaURL.IsAbs() {
		return ErrInvalidMedia
	}

	ext := filepath.Ext(mediaURL.EscapedPath())
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.ensureAppLaunched(DefaultReceiverAppID); err != nil {
		return err
	}

	metadata := MediaMetadata{"type": 0}
	if mediaMetadata != nil {
		if mediaMetadata.Title() != "" {
			metadata["title"] = mediaMetadata.Title()
		}

		if mediaMetadata.Subtitle() != "" {
			metadata["subtitle"] = mediaMetadata.Subtitle()
		}

		if mediaMetadata.ImageURL() != nil {
			metadata["images"] = []map[string]string{
				map[string]string{"url": mediaMetadata.ImageURL().String()},
			}
		}
	}

	mediaInfo := &MediaInformation{
		ContentID:   mediaURL.String(),
		ContentType: contentType,
		Metadata:    metadata,
		StreamType:  "BUFFERED",
	}

	return s.r.Load(s.ID, mediaInfo)
}

// MediaURL returns the URL of current loaded media.
func (s *Sender) MediaURL() *url.URL {
	if s.IsIdle() {
		return nil
	}

	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return nil
	}

	switch s.r.Application().AppID {
	case YouTubeReceiverAppID:
		return &url.URL{
			Scheme: "https",
			Host:   "youtu.be",
			Path:   ms.Media.ContentID,
		}
	default:
		u, err := url.Parse(ms.Media.ContentID)
		if err != nil {
			return nil
		}
		return u
	}
}

// MediaMetadata returns the metadata of current loaded media.
func (s *Sender) MediaMetadata() omnicast.MediaMetadata {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return nil
	}

	return ms.Media.Metadata
}

// MediaDuration returns the duration of current loaded media.
func (s *Sender) MediaDuration() time.Duration {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return time.Duration(0)
	}

	return time.Duration(ms.Media.Duration * float64(time.Second))
}

// PlayerState return the current playback state.
func (s *Sender) PlayerState() string {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return "IDLE"
	}

	return ms.PlayerState
}

// IsIdle returns true if the recevier device does not have an receiver
// app running or have media playback stopped.
func (s *Sender) IsIdle() bool {
	return s.PlayerState() == "IDLE"
}

// IsPlaying returns true if the recevier is actively playing content.
func (s *Sender) IsPlaying() bool {
	return s.PlayerState() == "PLAYING"
}

// IsPaused returns true if playback is paused due to user request.
func (s *Sender) IsPaused() bool {
	return s.PlayerState() == "PAUSED"
}

// IsBuffering returns true if playback is effectively paused due to
// buffer underflow.
func (s *Sender) IsBuffering() bool {
	return s.PlayerState() == "BUFFERING"
}

// PlaybackPosition returns the current position of media playback from
// the beginning of media content. For live streams, it returns the time
// since playback started.
func (s *Sender) PlaybackPosition() time.Duration {
	ms, ts := s.r.Session(s.ID)
	if ms == nil {
		return time.Duration(0)
	}

	pos := ms.CurrentTime * float64(time.Second)
	if ms.PlayerState == "PLAYING" {
		t := time.Since(ts).Seconds()
		pos += t * float64(ms.PlaybackRate)
	}

	return time.Duration(pos)
}

// PlaybackRate returns the ratio of speed that media is played at.
func (s *Sender) PlaybackRate() float32 {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return 0
	}

	return ms.PlaybackRate
}

// Play begins playback of the loaded media content from the current
// playback position.
func (s *Sender) Play() {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return
	}

	s.r.Play(s.ID, ms.MediaSessionID)
}

// Pause pauses playback of the current content.
func (s *Sender) Pause() {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return
	}

	s.r.Pause(s.ID, ms.MediaSessionID)
}

// Stop stops the playback and unload the current content
func (s *Sender) Stop() {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return
	}

	s.r.Stop(s.ID, ms.MediaSessionID)
}

// SeekTo sets the current playback position to pos,
func (s *Sender) SeekTo(pos time.Duration) {
	ms, _ := s.r.Session(s.ID)
	if ms == nil {
		return
	}

	s.r.Seek(s.ID, ms.MediaSessionID, pos.Seconds())
}

// VolumeLevel returns receiver volume as a number between 0.0 and 1.0.
func (s *Sender) VolumeLevel() float64 {
	if s.r.Volume() == nil {
		return 0.0
	}

	return s.r.Volume().Level
}

// IsMuted returns true if the receiver is muted.
func (s *Sender) IsMuted() bool {
	if s.r.Volume() == nil {
		return false
	}

	return s.r.Volume().Muted
}

// SetVolumeLevel sets receiver volume level.
func (s *Sender) SetVolumeLevel(level float64) {
	s.r.SetVolume(&ReceiverVolume{Level: level})
}

// Mute mutes the receiver.
func (s *Sender) Mute() {
	s.r.SetVolume(&ReceiverVolume{Muted: true})
}

// Unmute unmutes the receiver.
func (s *Sender) Unmute() {
	s.r.SetVolume(&ReceiverVolume{Muted: false})
}

// Close closes the connected receiver, if any.
func (s *Sender) Close() error {
	if s.r == nil {
		return nil
	}

	return s.r.Close()
}
