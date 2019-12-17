package castv2

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

// A vconn is a virtual connection represented by a pair of source and
// destination ID.
type vconn struct {
	LocalID  string
	RemoteID string
}

func (vc vconn) buildPayload(msgType string) string {
	return `{ "type": "` + msgType + `" }`
}

func (vc vconn) NewMsg(namespace, payload string) *Msg {
	return &Msg{vc.LocalID, vc.RemoteID, namespace, payload}
}

func (vc vconn) NewConnectMsg() *Msg {
	return vc.NewMsg(NamespaceConnection, vc.buildPayload(TypeConnect))
}

func (vc vconn) NewCloseMsg() *Msg {
	return vc.NewMsg(NamespaceConnection, vc.buildPayload(TypeClose))
}

func (vc vconn) NewPingMsg() *Msg {
	return vc.NewMsg(NamespaceHeartbeat, vc.buildPayload(TypePing))
}

func (vc vconn) NewPongMsg() *Msg {
	return vc.NewMsg(NamespaceHeartbeat, vc.buildPayload(TypePong))
}

// Channel represents a cast channel to the receiver device.
//
// It also manages the virtual connections. If a messages will be sent
// to a new source and destination ID pair, a virtual connection will be
// automatically established and keeped alive.
type Channel struct {
	conn        *tls.Conn
	done        chan struct{}
	vconns      map[vconn]struct{}
	heartbeat   *time.Ticker
	lastReqID   uint64
	pendingReqs map[uint64]chan *Msg
	// FIXME: make this thread-safe
	subscriptions map[string][]chan *Msg
}

// NewChannel connects to the receiver device at the given address over
// TLS and returns a Channel.
func NewChannel(addr string) (*Channel, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	c := &Channel{
		conn:          conn,
		done:          make(chan struct{}),
		vconns:        make(map[vconn]struct{}),
		heartbeat:     time.NewTicker(5 * time.Second),
		lastReqID:     0,
		pendingReqs:   make(map[uint64]chan *Msg),
		subscriptions: make(map[string][]chan *Msg),
	}

	go c.listen()
	go c.keepalive()

	return c, nil
}

// readMsg reads a message from the channel and blocks until it returns.
func (c *Channel) readMsg() (*Msg, error) {
	// Each message is prefixed withs its length as a big-endian uint32.
	var n uint32
	if err := binary.Read(c.conn, binary.BigEndian, &n); err != nil {
		return nil, err
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		return nil, err
	}

	msg := new(Msg)
	err := msg.UnmarshalBinary(buf)

	return msg, err
}

// writeMsg sends the message over the wire.
func (c *Channel) writeMsg(msg *Msg) error {
	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	if err := binary.Write(c.conn, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}

func (c *Channel) listen() {
	for {
		select {
		case <-c.done:
			close(c.done)
			return
		default:
			msg, err := c.readMsg()
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
					continue
				}

				log.Println(err)
				return
			}

			var p Header
			if err := json.Unmarshal([]byte(msg.Payload), &p); err != nil {
				log.Printf("unexpected payload: %s\n", msg.Payload)
				continue
			}

			if msg.DestinationID == "*" {
				if subs, ok := c.subscriptions[p.Type]; ok {
					for _, sub := range subs {
						if sub != nil {
							sub <- msg
						}
					}
				}

				continue
			}

			vc := vconn{msg.DestinationID, msg.SourceID}
			if _, ok := c.vconns[vc]; ok {
				switch msg.Namespace {
				case NamespaceHeartbeat:
					if p.Type == TypePing {
						c.writeMsg(vc.NewPongMsg())
					}

					continue
				case NamespaceConnection:
					if p.Type == TypeClose {
						delete(c.vconns, vc)
					}

					continue
				}

				if ch, ok := c.pendingReqs[p.RequestID]; ok {
					ch <- msg
					delete(c.pendingReqs, p.RequestID)
					continue
				}
			}

			log.Println("Unhandled msg:", msg)
		}
	}
}

func (c *Channel) keepalive() {
	for range c.heartbeat.C {
		for vc := range c.vconns {
			c.writeMsg(vc.NewPingMsg())
		}
	}
}

// Close terminates all established virtual connections and then closes
// the underying TLS connection.
func (c *Channel) Close() error {
	// Stop heartbeats
	c.heartbeat.Stop()

	// Close all virtual connections
	for vc := range c.vconns {
		delete(c.vconns, vc)
		c.writeMsg(vc.NewCloseMsg())
	}

	// Stop listening
	c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	c.done <- struct{}{}
	<-c.done

	return c.conn.Close()
}

// Request sends a request
func (c *Channel) Request(srcID, descID, namespace string, req Request, respCh chan *Msg) error {
	vc := vconn{srcID, descID}
	if _, ok := c.vconns[vc]; !ok {
		err := c.writeMsg(vc.NewConnectMsg())
		if err != nil {
			return err
		}

		c.vconns[vc] = struct{}{}
	}

	reqID := atomic.AddUint64(&c.lastReqID, 1)
	req.SetRequestID(reqID)

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if respCh != nil {
		c.pendingReqs[reqID] = respCh
	}

	return c.writeMsg(vc.NewMsg(namespace, string(payload)))
}

// Subscribe registers a subscription to a particular type of broadcast
// messages. It returns an ID for identifying the subscription (along
// with the message type) when unsubscribing.
func (c *Channel) Subscribe(msgType string, subCh chan *Msg) (int, error) {
	if msgType != TypeReceiverStatus && msgType != TypeMediaStatus {
		return 0, errors.New("message type unavailable for subscribe")
	}

	if _, ok := c.subscriptions[msgType]; !ok {
		c.subscriptions[msgType] = make([]chan *Msg, 0)
	}
	c.subscriptions[msgType] = append(c.subscriptions[msgType], subCh)

	return len(c.subscriptions[msgType]), nil
}

// Unsubscribe unregisters the subscription and closes the subscription
// channel.
func (c *Channel) Unsubscribe(msgType string, subID int) error {
	if subs, ok := c.subscriptions[msgType]; ok {
		if len(subs) > subID && subs[subID] != nil {
			subs[subID] = nil
			return nil

		}
	}

	return errors.New("subscription not found")
}
