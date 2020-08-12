package mpris

import (
	"net/url"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// MediaMetadata is a mapping from metadata attribute names to values.
//
// https://www.freedesktop.org/wiki/Specifications/mpris-spec/metadata/
type MediaMetadata map[string]dbus.Variant

// TrackID returns the raw mpris:trackid without parsing it.
func (m MediaMetadata) TrackID() dbus.Variant {
	return m["mpris:trackid"]
}

// Title returns the descriptive title of the content.
func (m MediaMetadata) Title() string {
	v, ok := m["xesam:title"]
	if !ok {
		return ""
	}

	return v.String()
}

// Subtitle returns the descriptive subtitle of the content. Usually,
// the name of the album.
func (m MediaMetadata) Subtitle() string {
	v, ok := m["xesam:album"]
	if !ok {
		return ""
	}

	return v.String()
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

	u, err := url.Parse(strings.Trim(v.String(), `"`))
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

	u, err := url.Parse(strings.Trim(v.String(), `"`))
	if err != nil {
		return nil
	}

	return u
}
