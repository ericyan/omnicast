package omnicast

import (
	"net/url"
)

type Player interface {
	Load(media *url.URL)
	Play()
	Pause()
}
