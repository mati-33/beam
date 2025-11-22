package absorber

import (
	"fmt"
	"net"

	p "github.com/mati-33/beam/internal/protocol"
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

func (a *Absorber) Receive() (p.Message, error) {
	m, err := p.ReadMessage(a.conn)
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

	err := p.WriteMessage(a.conn, m)
	if err != nil {
		return fmt.Errorf("failed to send message to emitter: %v", err)
	}
	return nil
}
