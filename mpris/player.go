package mpris

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/ericyan/omnicast"
)

const (
	DBusPath      = "org.mpris.MediaPlayer2"
	DBusInterface = "/org/mpris/MediaPlayer2"
)

// Discover returns MPRIS players available.
func Discover() ([]string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	var names []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return nil, err
	}

	var dests []string
	for _, name := range names {
		if strings.HasPrefix(name, DBusPath) {
			dests = append(dests, name)
		}
	}

	if len(dests) == 0 {
		return nil, errors.New("no mpris player instance found")
	}

	return dests, nil
}

// Player represents a MPRIS player.
type Player struct {
	dest string
	bo   dbus.BusObject
}

// NewPlayer returns a new player.
func NewPlayer(dest string) (*Player, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	bo := conn.Object(dest, DBusInterface)

	return &Player{dest, bo}, nil
}

// Name returns the name of the player instace.
func (p *Player) Name() string {
	return p.dest
}

// call invokes a MPRIS method.
func (p *Player) call(method string, args ...interface{}) ([]interface{}, error) {
	call := p.bo.Call(DBusPath+"."+method, 0, args...)
	return call.Body, call.Err
}

// Load opens media and starts playback.
func (p *Player) Load(media *url.URL, metadata omnicast.MediaMetadata) error {
	_, err := p.call("Player.OpenUri", media.String())
	return err
}

// metadata returns the MPRIS metadata.
func (p *Player) metadata() MediaMetadata {
	v, err := p.bo.GetProperty(DBusPath + ".Player.Metadata")
	if err != nil {
		return nil
	}

	m := v.Value().(map[string]dbus.Variant)
	return MediaMetadata(m)
}

// MediaURL returns the URL of current loaded media.
func (p *Player) MediaURL() *url.URL {
	return p.metadata().MediaURL()
}

// MediaMetadata returns the metadata of current loaded media.
func (p *Player) MediaMetadata() omnicast.MediaMetadata {
	return p.metadata()
}

// MediaDuration returns the duration of current loaded media.
func (p *Player) MediaDuration() time.Duration {
	return p.metadata().MediaDuration()
}

// Play starts or resumes playback.
func (p *Player) Play() {
	p.call("Player.Play")
}

// Pause pauses playback of the current content.
func (p *Player) Pause() {
	p.call("Player.Pause")
}

// Stop stops the playback and resets the playback position.
func (p *Player) Stop() {
	p.call("Player.Stop")
}

// SeekTo sets the current playback position to pos.
func (p *Player) SeekTo(pos time.Duration) {
	trackID := p.metadata().TrackID()

	p.call("Player.SetPosition", trackID, pos.Microseconds())
}

// PlaybackStatus return the current playback status.
func (p *Player) PlaybackStatus() string {
	v, err := p.bo.GetProperty(DBusPath + ".Player.PlaybackStatus")
	if err != nil {
		return "UNKNOWN"
	}

	return v.String()
}

// IsIdle returns true if the media playback stopped.
func (p *Player) IsIdle() bool {
	return p.PlaybackStatus() == "Stopped"
}

// IsPlaying returns true if the player is actively playing content.
func (p *Player) IsPlaying() bool {
	return p.PlaybackStatus() == "Playing"
}

// IsPaused returns true if playback is paused.
func (p *Player) IsPaused() bool {
	return p.PlaybackStatus() == "Paused"
}

// IsBuffering always returns false as the MPRIS API does not provide
// this information.
func (p *Player) IsBuffering() bool {
	return false
}

// PlaybackPosition returns the current position of media playback from
// the beginning of media content.
func (p *Player) PlaybackPosition() time.Duration {
	v, err := p.bo.GetProperty(DBusPath + ".Player.Position")
	if err != nil {
		return time.Duration(0)
	}

	pos := v.Value().(int64)
	return time.Duration(pos) * time.Microsecond
}

// PlaybackRate returns the ratio of speed that media is played at.
func (p *Player) PlaybackRate() float32 {
	v, err := p.bo.GetProperty(DBusPath + ".Player.Rate")
	if err != nil {
		return 0
	}

	return float32(v.Value().(float64))
}
