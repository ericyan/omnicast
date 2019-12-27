package types

import "net/url"

// Metadata implements the omnicast.MediaMetadata interface.
type Metadata struct{}

// Title returns the descriptive title of the content.
func (m *Metadata) Title() string {
	return "" // TODO
}

// Subtitle returns the descriptive subtitle of the content.
func (m *Metadata) Subtitle() string {
	return "" // TODO
}

// ImageURL returns the URL of the image.
func (m *Metadata) ImageURL() *url.URL {
	return nil // TODO
}

// UnmarshalText fills the Struct with media metadata described in the
// DIDL-Lite XML fragment.
func (m *Metadata) UnmarshalText(didl []byte) error {
	return nil // TODO
}
