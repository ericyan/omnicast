package types

import (
	"encoding/xml"
	"net/url"

	"github.com/ericyan/omnicast/upnp/internal/didl"
)

// Metadata implements the omnicast.MediaMetadata interface.
type Metadata map[string]string

// Title returns the descriptive title of the content.
func (m Metadata) Title() string {
	return m["title"]
}

// Subtitle returns the descriptive subtitle of the content.
func (m Metadata) Subtitle() string {
	return m["dc:creator"]
}

// ImageURL returns the URL of the image.
func (m Metadata) ImageURL() *url.URL {
	uri, ok := m["albumArtURI"]
	if !ok {
		return nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil
	}

	return u
}

// UnmarshalText fills the map with media metadata contained in the
// DIDL-Lite XML fragment.
func (m Metadata) UnmarshalText(data []byte) error {
	doc := new(didl.Document)
	if err := xml.Unmarshal(data, doc); err != nil {
		return err
	}

	if i := len(doc.Items); i > 0 {
		for _, v := range doc.Items[i-1].Values {
			m[v.Type()] = v.String()
		}
	}

	return nil
}
