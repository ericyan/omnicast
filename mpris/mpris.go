package mpris

import (
	"errors"
	"net/url"
	"strings"

	"github.com/godbus/dbus"
)

const (
	DBusPath      = "org.mpris.MediaPlayer2"
	DBusInterface = "/org/mpris/MediaPlayer2"
)

type Player struct {
	bo dbus.BusObject
}

func NewPlayer() (*Player, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	ret := conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0)
	if ret.Err != nil {
		return nil, ret.Err
	}
	if len(ret.Body) == 0 {
		return nil, errors.New("DBus connection error")
	}

	var dest string
	for _, s := range ret.Body[0].([]string) {
		if strings.HasPrefix(s, DBusPath) {
			dest = s
		}
	}
	if dest == "" {
		return nil, errors.New("no mpris player instance found")
	}

	return &Player{conn.Object(dest, DBusInterface)}, nil
}

func (p *Player) Call(method string, args ...interface{}) ([]interface{}, error) {
	ret := p.bo.Call(DBusPath+"."+method, 0, args...)
	return ret.Body, ret.Err
}

func (p *Player) Load(media *url.URL) {
	p.Call("Player.OpenUri", media.String())
}

func (p *Player) Play() {
	p.Call("Player.Play")
}

func (p *Player) Pause() {
	p.Call("Player.Pause")
}
