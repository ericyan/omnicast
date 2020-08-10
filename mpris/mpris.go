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
	trackID := 0 // FIXME: get the actual track ID.

	p.call("Player.SetPosition", trackID, pos.Microseconds())
}
