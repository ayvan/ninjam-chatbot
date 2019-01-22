package models

import (
	"fmt"
	"bytes"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

type ServerUserInfoChangeNotify struct {
	UserInfos []UserInfo
}

type UserInfo struct {
	Active       uint8
	ChannelIndex uint8
	Volume       [2]byte
	Pan          byte
	Flags        uint8
	Name         []byte
	Channels     [][]byte
}

func (s *ServerUserInfoChangeNotify) Unmarshal(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Input data error: %s", r)
			logrus.Debug(string(debug.Stack()))
			return
		}
	}()

	if s.UserInfos == nil {
		s.UserInfos = make([]UserInfo, 0)
	}

	for {
		if len(data) < 7 {
			break
		}

		userInfo := UserInfo{}
		userInfo.Active = uint8(data[0])
		userInfo.ChannelIndex = uint8(data[1])

		copy(userInfo.Volume[:], data[2:4])

		userInfo.Pan = data[4]
		userInfo.Flags = uint8(data[5])

		data = data[6:]

		nulTerminator := bytes.Index(data, []byte{0x0})

		userInfo.Name = data[:nulTerminator]

		data = data[nulTerminator+1:]

		nulTerminator = bytes.Index(data, []byte{0x0})

		channel := data[:nulTerminator]

		data = data[nulTerminator+1:]

		if userInfo.Channels == nil {
			userInfo.Channels = make([][]byte, 0)
		}

		userInfo.Channels = append(userInfo.Channels, channel)

		nulTerminator = bytes.Index(data, []byte{0x0})

		s.UserInfos = append(s.UserInfos, userInfo)

		if nulTerminator == -1 {
			break
		}
	}

	return nil
}
