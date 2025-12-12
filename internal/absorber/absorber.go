package absorber

import (
	"errors"
	"io"
	"net"

	p "github.com/mati-33/beam/internal/protocol"
)

type Absorber struct {
	conn net.Conn
}

func New(addr string) (*Absorber, error) {
	conn, err := net.Dial("tcp", addr+":3000")
	if err != nil {
		return &Absorber{}, err
	}
	return &Absorber{conn}, nil
}

func (a *Absorber) Close() error {
	return a.conn.Close()
}

func (a *Absorber) Receive() (p.Message, error) {
	m, err := p.ReadMessage(a.conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return p.Message{}, errors.New("emitter disconnected")
		}
		return p.Message{}, err
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

	return p.WriteMessage(a.conn, m)
}
