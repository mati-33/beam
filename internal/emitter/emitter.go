package emitter

import (
	"fmt"
	"net"

	p "github.com/mati-33/beam/internal/protocol"
	t "github.com/mati-33/beam/internal/transport"
)

type Emitter struct {
	l    net.Listener
	conn net.Conn
}

func New() (*Emitter, error) {
	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		return &Emitter{}, fmt.Errorf("failed to start tcp server: %v", err)
	}
	return &Emitter{l: l}, nil
}

func (e *Emitter) Close() error {
	if e.conn != nil {
		_ = e.conn.Close()
	}
	return e.l.Close()
}

func (e *Emitter) AcceptAbsorber() error {
	a, err := e.l.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept absorber: %v", err)
	}
	e.conn = a
	return nil
}

func (e *Emitter) Receive(msgType p.MessageType) (p.Message, error) {
	if e.conn == nil {
		panic("called Receive before AcceptAbsorber")
	}

	switch msgType {
	case p.FC, p.FI:
		panic("emitter cannot receive FC/FI messages")
	case p.OK, p.NO, p.BC:
	default:
		panic("unknown message type")
	}

	m, err := t.ReadMessage(e.conn)
	if err != nil {
		return p.Message{}, fmt.Errorf("failed to receive message from absorber: %v", err)
	}
	return m, nil
}
func (e *Emitter) Send(m p.Message) error {
	if e.conn == nil {
		panic("called Send before AcceptAbsorber")
	}

	switch m.Type {
	case p.BC:
		panic("emitter cannot send BC message")
	case p.OK, p.NO, p.FI, p.FC:
	default:
		panic("unknown message type")
	}

	err := t.WriteMessage(e.conn, m)
	if err != nil {
		return fmt.Errorf("failed to send message to absorber: %v", err)
	}
	return nil
}
