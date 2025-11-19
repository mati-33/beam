package absorber

import (
	"fmt"
	"net"

	p "github.com/mati-33/beam/internal/protocol"
	t "github.com/mati-33/beam/internal/transport"
)

type Absorber struct {
	conn net.Conn
}

func New() (*Absorber, error) {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		return &Absorber{}, fmt.Errorf("failed to connect to tcp server: %v", err)
	}
	return &Absorber{conn}, nil
}

func (a *Absorber) Close() error {
	return a.conn.Close()
}

func (a *Absorber) Receive(msgType p.MessageType) (p.Message, error) {
	switch msgType {
	case p.BC:
		panic("absorber cannot receive BC message")
	case p.OK, p.NO, p.FC, p.FI:
	default:
		panic("unknown message type")
	}

	m, err := t.ReadMessage(a.conn)
	if err != nil {
		return p.Message{}, fmt.Errorf("failed to receive message from emitter: %v", err)
	}
	return m, nil
}

func (a *Absorber) Send(m p.Message) error {
	switch m.Type {
	case p.FI, p.FC:
		panic("absorber cannot send FI/FC message")
	case p.OK, p.NO, p.BC:
	default:
		panic("unknown message type")
	}

	err := t.WriteMessage(a.conn, m)
	if err != nil {
		return fmt.Errorf("failed to send message to emitter: %v", err)
	}
	return nil
}
