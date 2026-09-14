package security

import "errors"

var (
	ErrAuthenticationFailed = errors.New("secure rcon authentication failed")
	ErrInvalidHMAC          = errors.New("invalid HMAC")
	ErrExpiredChallenge     = errors.New("expired challenge")
	ErrReplayDetected       = errors.New("replay detected")
	ErrUnsupportedProtocol  = errors.New("unsupported secure rcon protocol")
	ErrInvalidPacket        = errors.New("invalid secure rcon packet")
	ErrConnectionTimeout    = errors.New("secure rcon connection timeout")
	ErrServerRejected       = errors.New("secure rcon server rejected authentication")
)
