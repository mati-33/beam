package transport

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	p "github.com/mati-33/beam/internal/protocol"
)

func WriteMessage(dst io.Writer, m p.Message) error {
	b := p.Encode(m)
	_, err := io.Copy(dst, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	return nil
}

func ReadMessage(src io.Reader) (p.Message, error) {
	header := make([]byte, 7)

	_, err := io.ReadFull(src, header)
	if err != nil {
		return p.Message{}, fmt.Errorf("failed to read message header: %v", err)
	}

	payloadLen := binary.BigEndian.Uint16(header[5:7])

	rest := make([]byte, payloadLen+4)
	_, err = io.ReadFull(src, rest)
	if err != nil {
		return p.Message{}, fmt.Errorf("failed to read message payload: %v", err)
	}

	m, err := p.Decode(append(header, rest...))
	if err != nil {
		return p.Message{}, fmt.Errorf("failed to decode message bytes: %v", err)
	}

	return m, nil
}
