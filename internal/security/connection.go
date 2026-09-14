package security

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type SecureConn struct {
	conn net.Conn
	mu   sync.Mutex
}

func (c *SecureConn) Execute(command string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.conn.SetDeadline(time.Now().Add(HandshakeTimeout)); err != nil {
		return "", err
	}
	if err := WritePacket(c.conn, Packet{Version: ProtocolVersion, Type: MessageCommand, Payload: []byte(command)}); err != nil {
		return "", err
	}
	packet, err := ReadPacket(c.conn)
	if err != nil {
		return "", err
	}
	if packet.Version != ProtocolVersion || packet.Type != MessageCommandResult {
		return "", fmt.Errorf("%w: expected command result", ErrInvalidPacket)
	}
	if len(packet.Payload) < 3 {
		return "", fmt.Errorf("%w: malformed command result", ErrInvalidPacket)
	}
	textLength := int(binary.BigEndian.Uint16(packet.Payload[1:]))
	if len(packet.Payload) != 3+textLength {
		return "", fmt.Errorf("%w: malformed command result", ErrInvalidPacket)
	}
	if packet.Payload[0] == 0 {
		return "", fmt.Errorf("%w: %s", ErrServerRejected, string(packet.Payload[3:]))
	}
	return string(packet.Payload[3:]), nil
}

func (c *SecureConn) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
