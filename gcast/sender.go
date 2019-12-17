package gcast

import (
	"errors"
)

// Errors used by the Sender.
var (
	ErrReceiverNotReady = errors.New("receiver not ready")
)

// A Sender is a sender app instance that controls media playback on the
// receiver. Its ID, which should be unique, is used to identify itself
// when communicating with the receiver.
type Sender struct {
	ID string

	ReceiverApp    *ReceiverApplication
	ReceiverVolume *ReceiverVolume

	MediaSessionID int
	MediaInfo      *MediaInformation

	r *Receiver
}

// NewSender returns a new Sender with given ID.
func NewSender(id string) *Sender {
	return &Sender{ID: id}
}

// ConnectTo makes a connection to the Receiver.
func (s *Sender) ConnectTo(raddr string) error {
	s.r = NewReceiver(raddr)

	err := s.r.Connect()
	if err != nil {
		return err
	}

	s.r.OnStatusUpdate(s.updateReceiverStatus)
	s.r.OnMediaStatusUpdate(s.updateMediaStatus)

	s.getReceiverStatus()
	s.getMediaStatus()

	return nil
}

func (s *Sender) updateReceiverStatus(rs *ReceiverStatus) {
	if apps := rs.Status.Applications; len(apps) > 0 {
		s.ReceiverApp = apps[0]
	} else {
		s.ReceiverApp = nil
	}

	s.ReceiverVolume = rs.Status.Volume
}

func (s *Sender) getReceiverStatus() error {
	rs, err := s.r.GetStatus()
	if err != nil {
		return err
	}

	s.updateReceiverStatus(rs)

	return nil
}

func (s *Sender) updateMediaStatus(ms *MediaStatus) {
	sess := ms.Status[0]
	s.MediaSessionID = sess.MediaSessionID
	s.MediaInfo = sess.Media
}

func (s *Sender) getMediaStatus() error {
	if s.ReceiverApp == nil || s.ReceiverApp.IsIdleScreen {
		return ErrReceiverNotReady
	}

	ms, err := s.r.GetMediaStatus(s.ID, s.ReceiverApp.SessionID)
	if err != nil {
		return err
	}
	if len(ms.Status) == 0 {
		return ErrReceiverNotReady
	}

	s.updateMediaStatus(ms)

	return nil
}

// Close closes the connected receiver, if any.
func (s *Sender) Close() error {
	if s.r == nil {
		return nil
	}

	return s.r.Close()
}
