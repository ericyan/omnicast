// Package castv2 provides a low-level implementation of Google Cast V2
// protocol.
package castv2

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/ericyan/omnicast/gcast/internal/castv2/cast_channel"
)

// Sender and receiver IDs to use for platform messages.
const (
	PlatformSenderID   = "sender-0"
	PlatformReceiverID = "receiver-0"
)

// Reserved message namespaces for internal messages.
const (
	NamespaceConnection = "urn:x-cast:com.google.cast.tp.connection"
	NamespaceHeartbeat  = "urn:x-cast:com.google.cast.tp.heartbeat"
	NamespaceReceiver   = "urn:x-cast:com.google.cast.receiver"
	NamespaceMedia      = "urn:x-cast:com.google.cast.media"
)

// Cast application protocol message types.
const (
	TypeConnect        = "CONNECT"
	TypeClose          = "CLOSE"
	TypePing           = "PING"
	TypePong           = "PONG"
	TypeGetStatus      = "GET_STATUS"
	TypeReceiverStatus = "RECEIVER_STATUS"
	TypeMediaStatus    = "MEDIA_STATUS"
	TypeLaunch         = "LAUNCH"
	TypeLoad           = "LOAD"
	TypePlay           = "PLAY"
	TypePause          = "PAUSE"
	TypeStop           = "STOP"
	TypeSeek           = "SEEK"
)

// Msg is a Cast V2 protocol data unit with textual payload.
type Msg struct {
	SourceID      string
	DestinationID string
	Namespace     string
	Payload       string
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (m *Msg) UnmarshalBinary(data []byte) error {
	cm := new(cast_channel.CastMessage)
	if err := proto.Unmarshal(data, cm); err != nil {
		return err
	}

	m.SourceID = cm.GetSourceId()
	m.DestinationID = cm.GetDestinationId()
	m.Namespace = cm.GetNamespace()
	m.Payload = cm.GetPayloadUtf8()

	if cm.GetPayloadType() != cast_channel.CastMessage_STRING {
		return errors.New("unspported payload type")
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (m *Msg) MarshalBinary() ([]byte, error) {
	cm := &cast_channel.CastMessage{
		ProtocolVersion: cast_channel.CastMessage_CASTV2_1_0.Enum(),
		SourceId:        &m.SourceID,
		DestinationId:   &m.DestinationID,
		Namespace:       &m.Namespace,
		PayloadType:     cast_channel.CastMessage_STRING.Enum(),
		PayloadUtf8:     &m.Payload,
	}

	return proto.Marshal(cm)
}

// String implements the fmt.Stringer interface.
func (m *Msg) String() string {
	return fmt.Sprintf("%s -> %s [%s] %s", m.SourceID, m.DestinationID, m.Namespace, m.Payload)
}

// Header contains the required fields in most payload types.
type Header struct {
	RequestID uint64 `json:"requestId,omitempty"`
	Type      string `json:"type"`
}

// SetRequestID sets the requestId header.
func (h *Header) SetRequestID(id uint64) {
	h.RequestID = id
}

// Request represents a request payload.
type Request interface {
	SetRequestID(id uint64)
}

// NewRequest returns a new request of given type.
func NewRequest(reqType string) Request {
	return &Header{Type: reqType}
}
