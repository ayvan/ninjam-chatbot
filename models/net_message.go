package models

import (
	"encoding/binary"
)

type Unmarshaler interface {
	Unmarshal(data []byte) error
}

type Marshaler interface {
	Marshal() ([]byte, error)
}

// NetMessage
type NetMessage struct {
	Type       uint8
	Length     uint32
	InPayload  Unmarshaler
	OutPayload Marshaler
	RawData    []byte
}

func NewNetMessage(t uint8) *NetMessage {
	nm := &NetMessage{}

	nm.Type = t

	return nm
}

func NewInNetMessage(header [5]byte) *NetMessage {

	nm := &NetMessage{}

	nm.Type = uint8(header[0])

	nm.Length = binary.LittleEndian.Uint32(header[1:])

	return nm
}

func (nm *NetMessage) Marshal() (data []byte, err error) {

	payloadBytes, err := nm.OutPayload.Marshal()

	if err != nil {
		return nil, err
	}

	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(payloadBytes)))
	responseMessageHeader := []byte{nm.Type}
	responseMessageHeader = append(responseMessageHeader, length...)

	result := append(responseMessageHeader, payloadBytes...)

	return result, nil
}

func (nm *NetMessage) Unmarshal(data []byte) error {
	nm.RawData = data
	switch nm.Type {
	case ServerAuthChallengeType:
		nm.InPayload = &ServerAuthChallenge{}
		return nm.InPayload.Unmarshal(data)
	case ServerAuthReplyType:
		nm.InPayload = &ServerAuthReply{}
		return nm.InPayload.Unmarshal(data)
	case ChatMessageType:
		nm.InPayload = &ChatMessage{}
		return nm.InPayload.Unmarshal(data)
	case ServerUserInfoChangeNotifyType:
		nm.InPayload = &ServerUserInfoChangeNotify{}
		return nm.InPayload.Unmarshal(data)
	}

	return nil
}

func hasBit(n uint32, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}
