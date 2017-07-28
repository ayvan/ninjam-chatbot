package models

import (
	"fmt"
	"bytes"
)

// ServerAuthReply
//0x01
type ServerAuthReply struct {
	Flag         uint8
	ErrorMessage []byte // NUL-terminated
	MaxChannels  uint8
}

func (sac *ServerAuthReply) Unmarshal(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Input data error: %s", r)
			return
		}
	}()

	sac.Flag = uint8(data[0])

	nulTerminator := bytes.Index(data[1:], []byte{0x0})

	sac.ErrorMessage = data[1:nulTerminator+1]

	sac.MaxChannels = data[nulTerminator+2]

	return nil
}
