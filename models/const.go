package models

// Message types:
// https://github.com/wahjam/wahjam/wiki/Ninjam-Protocol

const (
	ServerAuthChallengeType        uint8 = 0x00
	ServerAuthReplyType            uint8 = 0x01
	ServerConfigChangeNotifyType   uint8 = 0x02
	ServerUserInfoChangeNotifyType uint8 = 0x03
	ClientAuthUserType             uint8 = 0x80
	ClientSetUsermaskType          uint8 = 0x81
	ClientSetChannelInfoType       uint8 = 0x82
	ClientUploadIntervalBeginType  uint8 = 0x83
	ClientUploadIntervalWriteType  uint8 = 0x84
	ChatMessageType                uint8 = 0xC0
	ClientKeepaliveType            uint8 = 0xfd
)

const (
	MSG   = "MSG"
	JOIN  = "JOIN"
	PART  = "PART"
	ADMIN = "ADMIN"
)
