package models

const (
	ServerAuthChallengeType        uint8 = 0x00
	ServerAuthReplyType            uint8 = 0x01
	ServerUserInfoChangeNotifyType uint8 = 0x03
	ClientAuthUserType             uint8 = 0x80
	ChatMessageType                uint8 = 0xC0
	ClientKeepaliveType            uint8 = 0xfd
)

const (
	MSG   = "MSG"
	JOIN  = "JOIN"
	PART  = "PART"
	ADMIN = "ADMIN"
)
