package security

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestSecureConnExecute(t *testing.T) {
	client, server := net.Pipe()
	secure := &SecureConn{conn: client}
	defer secure.Close()
	defer server.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		packet, err := ReadPacket(server)
		if err != nil || packet.Type != MessageCommand || string(packet.Payload) != "list" {
			return
		}
		payload := make([]byte, 3+len("players: 2"))
		payload[0] = 1
		binary.BigEndian.PutUint16(payload[1:], uint16(len("players: 2")))
		copy(payload[3:], "players: 2")
		_ = WritePacket(server, Packet{Version: ProtocolVersion, Type: MessageCommandResult, Payload: payload})
	}()

	result, err := secure.Execute("list")
	if err != nil {
		t.Fatal(err)
	}
	if result != "players: 2" {
		t.Fatalf("result = %q, want players: 2", result)
	}
	<-done
}
