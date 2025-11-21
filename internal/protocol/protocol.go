package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

type MessageType byte

const (
	OK MessageType = 0x01
	NO MessageType = 0x02
	BC MessageType = 0x03
	FI MessageType = 0x04
	FC MessageType = 0x05
)

var startingSequence = [...]byte{0x06, 0x07, 0x08, 0x09}

type Message struct {
	Type    MessageType
	Payload []byte
}

func NewOK() *Message         { return &Message{Type: OK} }
func NewNO() *Message         { return &Message{Type: NO} }
func NewBC(p []byte) *Message { return &Message{Type: BC, Payload: p} }
func NewFI(p []byte) *Message { return &Message{Type: FI, Payload: p} }
func NewFC(p []byte) *Message { return &Message{Type: FC, Payload: p} }

func WriteMessage(w io.Writer, m Message) error {
	b := make([]byte, 0, 7+len(m.Payload)+4)
	b = append(b, startingSequence[:]...)
	b = append(b, byte(m.Type))
	b = binary.BigEndian.AppendUint16(b, uint16(len(m.Payload)))
	b = append(b, m.Payload...)
	b = binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(m.Payload))

	_, err := w.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

func ReadMessage(r io.Reader) (Message, error) {
	header := make([]byte, 7)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return Message{}, fmt.Errorf("failed to read message header: %v", err)
	}

	if !bytes.Equal(header[:4], startingSequence[:]) {
		return Message{}, errors.New("invalid starting byte sequence")
	}

	mt := MessageType(header[4])
	switch mt {
	case OK, NO, BC, FI, FC:
	default:
		return Message{}, errors.New("invalid message type")
	}

	payloadLen := binary.BigEndian.Uint16(header[5:7])
	rest := make([]byte, payloadLen+4)
	_, err = io.ReadFull(r, rest)
	if err != nil {
		return Message{}, fmt.Errorf("failed to read message payload: %v", err)
	}

	payload := rest[:payloadLen]
	crc := binary.BigEndian.Uint32(rest[len(rest)-4:])

	if crc32.ChecksumIEEE(payload) != crc {
		return Message{}, errors.New("invalid payload checksum")
	}

	return Message{
		Type:    mt,
		Payload: payload,
	}, nil
}
