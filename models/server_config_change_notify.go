package models

import (
	"encoding/binary"
	"fmt"
)

// ServerAuthReply
//0x02
type ServerConfigChangeNotify struct {
	BPM uint16
	BPI uint16
}

func (sac *ServerConfigChangeNotify) Unmarshal(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Input data error: %s", r)
			return
		}
	}()

	sac.BPM = binary.LittleEndian.Uint16(data[:2])
	sac.BPI = binary.LittleEndian.Uint16(data[2:4])

	return nil
}
