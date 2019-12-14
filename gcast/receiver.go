package gcast

import (
	"encoding/json"

	"github.com/ericyan/omnicast/gcast/internal/castv2"
)

// ReceiverApplication represents an instance of receiver application.
type ReceiverApplication struct {
	AppID               string              `json:"appId"`
	Name                string              `json:"displayName"`
	IconURL             string              `json:"iconUrl"`
	StatusText          string              `json:"statusText"`
	IsIdleScreen        bool                `json:"isIdleScreen"`
	SupportedNamespaces []map[string]string `json:"namespaces"`
	SessionID           string              `json:"sessionId"`
	TransportID         string              `json:"transportId"`
}

// ReceiverVolume represents the volume of the receiver device.
type ReceiverVolume struct {
	ControlType  string  `json:"controlType"`
	Level        float64 `json:"level"`
	Muted        bool    `json:"muted"`
	StepInterval float64 `json:"stepInterval"`
}

// ReceiverStatus represents the devices status of the receiver.
type ReceiverStatus struct {
	castv2.Header
	Status struct {
		Applications []*ReceiverApplication `json:"applications,omitempty"`
		Volume       *ReceiverVolume        `json:"volume"`
	} `json:"status"`
}

// Receiver represents a Google Cast device.
type Receiver struct {
	ch *castv2.Channel
}

// Connect makes a connection to the receiver.
func Connect(addr string) (*Receiver, error) {
	ch, err := castv2.NewChannel(addr)
	if err != nil {
		return nil, err
	}

	return &Receiver{ch}, nil
}

// GetStatus returns the devices status of the receiver.
func (r *Receiver) GetStatus() (*ReceiverStatus, error) {
	respCh := make(chan *castv2.Msg)
	err := r.ch.Request(
		castv2.PlatformSenderID,
		castv2.PlatformReceiverID,
		castv2.NamespaceReceiver,
		castv2.NewRequest(castv2.TypeGetStatus),
		respCh,
	)
	if err != nil {
		return nil, err
	}
	resp := <-respCh

	var rs ReceiverStatus
	if err := json.Unmarshal([]byte(resp.Payload), &rs); err != nil {
		return nil, err
	}

	return &rs, nil
}

// Close closes the connection to the receiver.
func (r *Receiver) Close() error {
	return r.ch.Close()
}
