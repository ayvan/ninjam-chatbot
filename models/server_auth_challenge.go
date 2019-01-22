package models

import (
	"time"
	"fmt"
	"github.com/sirupsen/logrus"
	"bytes"
	"encoding/binary"
)

// ServerAuthChallenge
type ServerAuthChallenge struct {
	Challenge          [8]uint8
	ServerCapabilities uint32
	ProtocolVersion    uint32
	LicenseAgreement   []byte // NUL-terminated
}

func (sac *ServerAuthChallenge) Unmarshal(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Input data error: %s", r)
			return
		}
	}()

	sac.Challenge = [8]uint8{}
	for i := 0; i < 8; i++ {
		sac.Challenge[i] = uint8(data[i])
	}

	sac.ServerCapabilities = binary.LittleEndian.Uint32(data[8:12])

	sac.ProtocolVersion = binary.LittleEndian.Uint32(data[12:16])

	nulTerminator := bytes.Index(data[16:], []byte{0x0})

	if nulTerminator != -1 {
		sac.LicenseAgreement = data[16:nulTerminator+16]
	}

	return nil
}

func (sac *ServerAuthChallenge) KeepAliveInterval() (interval time.Duration, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Input data error: %s", r)
			return
		}
	}()

	b := make([]byte, 4)

	binary.LittleEndian.PutUint32(b, sac.ServerCapabilities)

	i := binary.LittleEndian.Uint16(b[1:])

	logrus.Infof("Keep alive interval: %ds", i)

	return time.Second * time.Duration(i), nil
}

func (sac *ServerAuthChallenge) HasAgreement() bool {
	return hasBit(sac.ServerCapabilities, 0)
}
