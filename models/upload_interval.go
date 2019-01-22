package models

import (
	"encoding/binary"
	"fmt"
)

// ClientUploadIntervalBegin
// 0x83
type ClientUploadIntervalBegin struct {
	GUID          [16]byte
	EstimatedSize uint32
	FourCC        [4]byte
	ChannelIndex  uint8
}

// ClientUploadIntervalWrite
// 0x84
type ClientUploadIntervalWrite struct {
	GUID      [16]uint8
	Flags     uint8
	AudioData []byte
}

func (c *ClientUploadIntervalBegin) Marshal() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Marshal error: %s", r)
			return
		}
	}()

	c.FourCC = [4]byte{}

	copy(c.FourCC[:], []byte("OGGv")[0:4])

	es := make([]byte, 4)
	binary.LittleEndian.PutUint32(es, c.EstimatedSize)

	data = append(data, c.GUID[:]...)
	data = append(data, es...)
	data = append(data, c.FourCC[:]...)
	data = append(data, c.ChannelIndex)

	return
}

func (c *ClientUploadIntervalWrite) Marshal() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Marshal error: %s", r)
			return
		}
	}()

	data = append(data, c.GUID[:]...)
	data = append(data, c.Flags)
	data = append(data, c.AudioData...)

	return
}
