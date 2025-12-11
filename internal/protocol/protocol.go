package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"strconv"
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
func NewFC(p []byte) *Message { return &Message{Type: FC, Payload: p} }

func NewFI(name string, size int64) *Message {
	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(size))
	p = append(p, []byte(name)...)
	return &Message{Type: FI, Payload: p}
}

type FileInfo struct {
	Name string
	Size int64
}

func DecodeFIPayload(p []byte) FileInfo {
	s := int64(binary.BigEndian.Uint64(p[:8]))
	n := string(p[8:])
	return FileInfo{Name: n, Size: s}
}

func DecodeBCPayload(p []byte) string {
	v := binary.BigEndian.Uint16(p)
	return fmt.Sprintf("%02x", v)
}

func EncodeBCPayload(v string) ([]byte, error) {
	i, err := strconv.ParseUint(v, 16, 8)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(i))

	return b, nil
}

func WriteMessage(w io.Writer, m Message) error {
	b := make([]byte, 0, 7+len(m.Payload)+4)
	b = append(b, startingSequence[:]...)
	b = append(b, byte(m.Type))
	b = binary.BigEndian.AppendUint16(b, uint16(len(m.Payload)))
	b = append(b, m.Payload...)
	b = binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(m.Payload))

	_, err := w.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func ReadMessage(r io.Reader) (Message, error) {
	header := make([]byte, 7)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return Message{}, fmt.Errorf("failed to read message header: %w", err)
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
		return Message{}, fmt.Errorf("failed to read message payload: %w", err)
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
