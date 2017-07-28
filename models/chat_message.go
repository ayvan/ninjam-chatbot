package models

import (
	"bytes"
	"fmt"
)

// ChatMessage
// 0xc0
type ChatMessage struct {
	Command []byte // NUL-terminated
	Arg1    []byte // NUL-terminated
	Arg2    []byte // NUL-terminated
	Arg3    []byte // NUL-terminated
	Arg4    []byte // NUL-terminated
}

func (cm *ChatMessage) Marshal() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Marshal error: %s", r)
			return
		}
	}()

	data = append(data, cm.Command...)
	data = append(data, byte(0))

	data = append(data, cm.Arg1...)
	data = append(data, byte(0))

	data = append(data, cm.Arg2...)
	data = append(data, byte(0))

	data = append(data, cm.Arg3...)
	data = append(data, byte(0))

	data = append(data, cm.Arg4...)
	data = append(data, byte(0))

	return
}

func (cm *ChatMessage) Unmarshal(data []byte) (err error) {

	if len(data) == 0 {
		return
	}

	nulTerminator := bytes.Index(data, []byte{0x0})

	cm.Command = data[:nulTerminator]

	data = data[nulTerminator+1:]

	if len(data) == 0 {
		return
	}

	nulTerminator = bytes.Index(data, []byte{0x0})

	cm.Arg1 = data[:nulTerminator]

	data = data[nulTerminator+1:]

	if len(data) == 0 {
		return
	}

	nulTerminator = bytes.Index(data, []byte{0x0})

	cm.Arg2 = data[:nulTerminator]

	data = data[nulTerminator+1:]

	if len(data) == 0 {
		return
	}

	nulTerminator = bytes.Index(data, []byte{0x0})

	cm.Arg3 = data[:nulTerminator]

	data = data[nulTerminator+1:]

	if len(data) == 0 {
		return
	}

	nulTerminator = bytes.Index(data, []byte{0x0})

	cm.Arg4 = data[:nulTerminator]

	return nil
}
