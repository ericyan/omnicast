package gcast

import "net/url"

// MediaMetadata represents a generic media artifact.
//
// Ref: https://developers.google.com/cast/docs/reference/messages#GenericMediaMetadata
//      https://developers.google.com/cast/docs/reference/messages#Image
type MediaMetadata map[string]interface{}

// Title returns the descriptive title of the content.
func (m MediaMetadata) Title() string {
	if _, ok := m["title"]; !ok {
		return ""
	}

	return m["title"].(string)
}

// Subtitle returns the descriptive subtitle of the content.
func (m MediaMetadata) Subtitle() string {
	if _, ok := m["subtitle"]; !ok {
		return ""
	}

	return m["subtitle"].(string)
}

// ImageURL returns the URL of the image.
func (m MediaMetadata) ImageURL() *url.URL {
	if _, ok := m["images"]; !ok {
		return nil
	}

	images := m["images"].([]interface{})
	if len(images) == 0 {
		return nil
	}

	img := images[0].(map[string]interface{})
	if _, ok := img["url"]; !ok {
		return nil
	}

	u, err := url.Parse(img["url"].(string))
	if err != nil {
		return nil
	}

	return u
}
