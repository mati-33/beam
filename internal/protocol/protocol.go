package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
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

func NewMessage(t MessageType, p []byte) *Message {
	switch t {
	case OK, NO:
		if len(p) != 0 {
			panic("OK/NO messages cannot have payload")
		}
	case BC, FI, FC:
		if len(p) == 0 {
			panic("BC/FI/FC messages require payload")
		}
	default:
		panic("unknown message type")
	}
	return &Message{t, p}
}

func Encode(m Message) []byte {
	b := make([]byte, 0, 7+len(m.Payload)+4)
	b = append(b, startingSequence[:]...)
	b = append(b, byte(m.Type))
	b = binary.BigEndian.AppendUint16(b, uint16(len(m.Payload)))
	b = append(b, m.Payload...)
	b = binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(m.Payload))
	return b
}

func Decode(b []byte) (Message, error) {
	if len(b) < 11 {
		return Message{}, errors.New("message too short")
	}

	if !bytes.Equal(b[:4], startingSequence[:]) {
		return Message{}, errors.New("invalid starting byte sequence")
	}

	mt := MessageType(b[4])
	switch mt {
	case OK, NO, BC, FI, FC:
	default:
		return Message{}, errors.New("invalid message type")
	}

	payloadLen := binary.BigEndian.Uint16(b[5:7])
	if len(b) != 7+int(payloadLen)+4 {
		return Message{}, errors.New("payload length missmatch")
	}

	payload := b[7 : 7+payloadLen]
	crc := binary.BigEndian.Uint32(b[len(b)-4:])

	if crc32.ChecksumIEEE(payload) != crc {
		return Message{}, errors.New("invalid payload checksum")
	}

	return Message{
		Type:    mt,
		Payload: payload,
	}, nil
}
