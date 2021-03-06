package mpris

import (
	"net/url"
	"time"

	"github.com/godbus/dbus/v5"
)

// MediaMetadata is a mapping from metadata attribute names to values.
//
// https://www.freedesktop.org/wiki/Specifications/mpris-spec/metadata/
type MediaMetadata map[string]dbus.Variant

// TrackID returns the track ID as a D-Bus object path.
func (m MediaMetadata) TrackID() dbus.ObjectPath {
	return m["mpris:trackid"].Value().(dbus.ObjectPath)
}

// Title returns the descriptive title of the content.
func (m MediaMetadata) Title() string {
	v, ok := m["xesam:title"]
	if !ok {
		return ""
	}

	return v.Value().(string)
}

// Subtitle returns the descriptive subtitle of the content. Usually,
// the name of the album.
func (m MediaMetadata) Subtitle() string {
	v, ok := m["xesam:album"]
	if !ok {
		return ""
	}

	return v.Value().(string)
}

// MediaDuration returns the duration of the media.
func (m MediaMetadata) MediaDuration() time.Duration {
	v, ok := m["mpris:length"]
	if !ok {
		return time.Duration(0)
	}

	return time.Duration(v.Value().(int64)) * time.Microsecond
}

// MediaURL returns the URL of the media content.
func (m MediaMetadata) MediaURL() *url.URL {
	v, ok := m["xesam:url"]
	if !ok {
		return nil
	}

	u, err := url.Parse(v.Value().(string))
	if err != nil {
		return nil
	}

	return u
}

// ImageURL returns the URL of the image.
func (m MediaMetadata) ImageURL() *url.URL {
	v, ok := m["mpris:artUrl"]
	if !ok {
		return nil
	}

	u, err := url.Parse(v.Value().(string))
	if err != nil {
		return nil
	}

	return u
}
