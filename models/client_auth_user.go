package models

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"crypto/sha1"
	"encoding/binary"
)

// ClientAuthUser
// 0x80
type ClientAuthUser struct {
	PasswordHash       [20]uint8
	Username           []byte // NUL-terminated
	ClientCapabilities uint32
	ClientVersion      uint32
}

func NewClientAuthUser(username, password string, authAgreement bool, challenge [8]uint8) *ClientAuthUser {
	cau := &ClientAuthUser{}
	up := []byte(username + ":" + password)

	sha1Sum := sha1.Sum(up)
	upSha1 := make([]byte, 0)
	upSha1 = append(upSha1, sha1Sum[0:20]...)
	upSha1 = append(upSha1, []byte(challenge[0:8])...)

	cau.PasswordHash = sha1.Sum(upSha1)
	cau.Username = []byte(username)
	cau.ClientVersion = 0x00020000

	if authAgreement {
		cau.ClientCapabilities = 0x1
	}

	return cau
}

func (cau *ClientAuthUser) Marshal() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Marshal error: %s", r)
			return
		}
	}()

	for _, b := range cau.PasswordHash {
		data = append(data, byte(b))
	}

	logrus.Info("Username:", cau.Username)

	data = append(data, cau.Username...)
	data = append(data, byte(0x0))

	cc := make([]byte, 4)
	cv := make([]byte, 4)

	binary.LittleEndian.PutUint32(cc, cau.ClientCapabilities)
	binary.LittleEndian.PutUint32(cv, cau.ClientVersion)

	data = append(data, cc...)
	data = append(data, cv...)

	return
}
