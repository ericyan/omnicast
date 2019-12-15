package gcast

// A Sender is a sender app instance that controls media playback on the
// receiver. Its ID, which should be unique, is used to identify itself
// when communicating with the receiver.
type Sender struct {
	ID string

	r *Receiver
}

// NewSender returns a new Sender with given ID.
func NewSender(id string) *Sender {
	return &Sender{ID: id}
}

// ConnectTo makes a connection to the Receiver.
func (s *Sender) ConnectTo(raddr string) error {
	s.r = NewReceiver(raddr)

	return s.r.Connect()
}
