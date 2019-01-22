package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ClientSetChannelInfo struct {
	Channels []ChannelInfo
}

type ChannelInfo struct {
	Name   string // NUL-terminated
	Volume int16  // (dB gain, 0=0dB, 10=1dB, -30=-3dB, etc)
	Pan    int8   // [-128, 127]
	Flags  uint8
}

func (c *ClientSetChannelInfo) Marshal() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Marshal error: %s", r)
			return
		}
	}()

	for _, channel := range c.Channels {
		channelsData := make([]byte, 0)

		cName := append([]byte(channel.Name), 0) // NUL-terminate string
		channelsData = append(channelsData, cName...)

		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, channel.Volume)
		if err != nil {
			err = fmt.Errorf("binary.Write failed: %s", err)
		}

		err = binary.Write(buf, binary.LittleEndian, channel.Pan)
		if err != nil {
			err = fmt.Errorf("binary.Write failed: %s", err)
		}

		channelsData = append(channelsData, buf.Bytes()...)

		channelsData = append(channelsData, channel.Flags)

		cps := make([]byte, 2)
		binary.LittleEndian.PutUint16(cps, uint16(6)) // i don't know why it's always 6...

		data = append(data, cps...)
		data = append(data, channelsData...)
	}

	return
}
