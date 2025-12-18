package emitter

import (
	"errors"
	"io"
	"net"

	p "github.com/mati-33/beam/internal/protocol"
)

type Emitter struct {
	l    net.Listener
	conn net.Conn
}

func New() (*Emitter, error) {
	l, err := net.Listen("tcp", ":9001")
	if err != nil {
		return &Emitter{}, err
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
		return err
	}
	e.conn = a
	return nil
}

func (e *Emitter) ClearAbsorber() {
	_ = e.conn.Close()
	e.conn = nil
}

func (e *Emitter) Receive() (p.Message, error) {
	if e.conn == nil {
		panic("called Receive before AcceptAbsorber")
	}

	m, err := p.ReadMessage(e.conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return p.Message{}, errors.New("absorber disconnected")
		}
		return p.Message{}, err
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

	return p.WriteMessage(e.conn, m)
}
