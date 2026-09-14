package security

import "time"

const (
	ProtocolVersion  uint16 = 1
	DefaultTimeStep  int64  = 30
	MaxPacketSize    uint32 = 4096
	HandshakeTimeout        = 5 * time.Second
)

const (
	MessageHello         uint8 = 1
	MessageChallenge           = 2
	MessageAuth                = 3
	MessageAuthSuccess         = 4
	MessageCommand             = 5
	MessageCommandResult       = 6
	MessageError               = 7
)

const (
	ErrorUnsupportedVersion byte = iota + 1
	ErrorInvalidHMAC
	ErrorExpiredChallenge
	ErrorReplay
	ErrorRejected
)

const (
	messageHello            = MessageHello
	messageChallenge        = MessageChallenge
	messageAuth             = MessageAuth
	messageAuthSuccess      = MessageAuthSuccess
	messageError            = MessageError
	errorUnsupportedVersion = ErrorUnsupportedVersion
	errorInvalidHMAC        = ErrorInvalidHMAC
	errorExpiredChallenge   = ErrorExpiredChallenge
	errorReplay             = ErrorReplay
	errorRejected           = ErrorRejected
)

type Packet struct {
	Version uint16
	Type    uint8
	Payload []byte
}

type Challenge struct {
	Nonce       []byte
	ServerTime  int64
	TimeCounter int64
}

type AuthResult struct {
	RCONAddress      string
	SessionToken     []byte
	ServerTimeOffset float64
	Reason           string
	Connection       *SecureConn
}
