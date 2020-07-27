package omnicast

import (
	"net/url"
	"time"
)

// MediaPlayer is a generic media player.
type MediaPlayer interface {
	MediaLoader
	MediaInfoReporter
	PlaybackStateReporter
	PlaybackController
	VolumeReporter
	VolumeController
}

// MediaMetadata describes a media artefact.
type MediaMetadata interface {
	Title() string
	Subtitle() string
	ImageURL() *url.URL
}

// MediaLoader loads the media for playback.
type MediaLoader interface {
	Load(media *url.URL, metadata MediaMetadata) error
}

// MediaInfoReporter provides information for the current media.
type MediaInfoReporter interface {
	MediaURL() *url.URL
	MediaMetadata() MediaMetadata
	MediaDuration() time.Duration
}

// PlaybackStateReporter retrieves media playback state.
type PlaybackStateReporter interface {
	IsIdle() bool
	IsPlaying() bool
	IsPaused() bool
	IsBuffering() bool
	PlaybackPosition() time.Duration
	PlaybackRate() float32
}

// PlaybackController provides methods for controlling media playback.
type PlaybackController interface {
	Play()
	Pause()
	Stop()
	SeekTo(pos time.Duration)
}

// VolumeReporter retrieves volume settings of audio output.
type VolumeReporter interface {
	VolumeLevel() float64
	IsMuted() bool
}

// VolumeController provides methods for adjusting volume settings.
type VolumeController interface {
	SetVolumeLevel(level float64)
	Mute()
	Unmute()
}
