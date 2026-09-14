package security

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

func Authenticate(address, clientID string, secret []byte) (*AuthResult, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("%w: empty secret", ErrAuthenticationFailed)
	}
	conn, err := net.DialTimeout("tcp", address, HandshakeTimeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionTimeout, err)
	}
	result, err := AuthenticateWithConn(conn, clientID, secret)
	if err != nil {
		_ = conn.Close()
	}
	return result, err
}

func AuthenticateWithConn(conn net.Conn, clientID string, secret []byte) (*AuthResult, error) {
	return authenticateWithConn(conn, clientID, secret, HandshakeTimeout)
}

func authenticateWithConn(conn net.Conn, clientID string, secret []byte, timeout time.Duration) (*AuthResult, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("%w: empty secret", ErrAuthenticationFailed)
	}
	if clientID == "" {
		return nil, fmt.Errorf("%w: empty client ID", ErrAuthenticationFailed)
	}
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	if err := WritePacket(conn, Packet{Version: ProtocolVersion, Type: MessageHello, Payload: encodeHello(clientID)}); err != nil {
		return nil, classifyIOError(err)
	}
	challengePacket, err := ReadPacket(conn)
	if err != nil {
		return nil, classifyIOError(err)
	}
	if challengePacket.Version != ProtocolVersion {
		return nil, fmt.Errorf("%w: client v%d, server v%d", ErrUnsupportedProtocol, ProtocolVersion, challengePacket.Version)
	}
	if challengePacket.Type == messageError {
		return nil, decodeServerError(challengePacket.Payload)
	}
	if challengePacket.Type != messageChallenge {
		return nil, fmt.Errorf("%w: expected challenge", ErrInvalidPacket)
	}
	challenge, err := decodeChallenge(challengePacket.Payload)
	if err != nil {
		return nil, err
	}
	counter := challenge.TimeCounter
	secretKey, err := base64.StdEncoding.DecodeString(string(secret))
	if err != nil || len(secretKey) == 0 {
		return nil, fmt.Errorf("%w: secret is not valid base64", ErrAuthenticationFailed)
	}
	authPayload := encodeAuth(clientID, counter, challenge.Nonce, CalculateHMAC(secretKey, clientID, counter, challenge.Nonce))
	if err := WritePacket(conn, Packet{Version: ProtocolVersion, Type: messageAuth, Payload: authPayload}); err != nil {
		return nil, classifyIOError(err)
	}
	resultPacket, err := ReadPacket(conn)
	if err != nil {
		return nil, classifyIOError(err)
	}
	if resultPacket.Version != ProtocolVersion {
		return nil, fmt.Errorf("%w: client v%d, server v%d", ErrUnsupportedProtocol, ProtocolVersion, resultPacket.Version)
	}
	if resultPacket.Type == messageError {
		return nil, decodeServerError(resultPacket.Payload)
	}
	if resultPacket.Type != messageAuthSuccess {
		return nil, fmt.Errorf("%w: expected auth success", ErrInvalidPacket)
	}
	result, err := decodeSuccess(resultPacket.Payload)
	if err != nil {
		return nil, err
	}
	result.ServerTimeOffset = float64(time.Now().UnixNano())/1e9 - float64(challenge.ServerTime)
	result.Connection = &SecureConn{conn: conn}
	return result, nil
}

func encodeHello(clientID string) []byte {
	clientBytes := []byte(clientID)
	payload := make([]byte, 2+len(clientBytes))
	binary.BigEndian.PutUint16(payload, uint16(len(clientBytes)))
	copy(payload[2:], clientBytes)
	return payload
}
func encodeAuth(clientID string, counter int64, nonce, digest []byte) []byte {
	_ = clientID
	_ = counter
	payload := make([]byte, 2)
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], uint16(len(nonce)))
	payload = append(payload[:0], length[:]...)
	payload = append(payload, nonce...)
	binary.BigEndian.PutUint16(length[:], uint16(len(digest)))
	payload = append(payload, length[:]...)
	payload = append(payload, digest...)
	return payload
}
func decodeChallenge(payload []byte) (Challenge, error) {
	if len(payload) < 18 {
		return Challenge{}, fmt.Errorf("%w: malformed challenge", ErrInvalidPacket)
	}
	nonceLength := int(binary.BigEndian.Uint16(payload))
	if nonceLength < 16 || nonceLength > 64 || len(payload) != 2+nonceLength+16 {
		return Challenge{}, fmt.Errorf("%w: malformed challenge", ErrInvalidPacket)
	}
	return Challenge{
		Nonce:       append([]byte(nil), payload[2:2+nonceLength]...),
		ServerTime:  int64(binary.BigEndian.Uint64(payload[2+nonceLength:])),
		TimeCounter: int64(binary.BigEndian.Uint64(payload[10+nonceLength:])),
	}, nil
}
func decodeSuccess(payload []byte) (*AuthResult, error) {
	if len(payload) < 3 {
		return nil, fmt.Errorf("%w: malformed auth success", ErrInvalidPacket)
	}
	success := payload[0] != 0
	reasonLength := int(binary.BigEndian.Uint16(payload[1:]))
	if len(payload) != 3+reasonLength {
		return nil, fmt.Errorf("%w: malformed auth success", ErrInvalidPacket)
	}
	result := &AuthResult{Reason: string(payload[3:])}
	if !success {
		switch result.Reason {
		case "INVALID_HMAC":
			return nil, ErrInvalidHMAC
		case "TIMESTAMP_OUT_OF_RANGE":
			return nil, ErrExpiredChallenge
		case "NONCE_ALREADY_USED":
			return nil, ErrReplayDetected
		default:
			return nil, fmt.Errorf("%w: %s", ErrServerRejected, result.Reason)
		}
	}
	return result, nil
}
func decodeServerError(payload []byte) error {
	if len(payload) == 0 {
		return ErrServerRejected
	}
	switch payload[0] {
	case errorUnsupportedVersion:
		return ErrUnsupportedProtocol
	case errorInvalidHMAC:
		return ErrInvalidHMAC
	case errorExpiredChallenge:
		return ErrExpiredChallenge
	case errorReplay:
		return ErrReplayDetected
	default:
		return ErrServerRejected
	}
}
func classifyIOError(err error) error {
	if errors, ok := err.(net.Error); ok && errors.Timeout() {
		return fmt.Errorf("%w: %v", ErrConnectionTimeout, err)
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("%w: server closed the secure RCON connection before sending a complete response; verify the gateway address and protocol framing", ErrAuthenticationFailed)
	}
	return err
}
