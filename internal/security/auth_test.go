package security

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"
)

func challengePayload(nonce []byte, serverTime, timeCounter int64) []byte {
	payload := make([]byte, 2+len(nonce)+16)
	binary.BigEndian.PutUint16(payload, uint16(len(nonce)))
	copy(payload[2:], nonce)
	binary.BigEndian.PutUint64(payload[2+len(nonce):], uint64(serverTime))
	binary.BigEndian.PutUint64(payload[10+len(nonce):], uint64(timeCounter))
	return payload
}

func successPayload(success bool, reason string) []byte {
	payload := make([]byte, 3+len(reason))
	if success {
		payload[0] = 1
	}
	binary.BigEndian.PutUint16(payload[1:], uint16(len(reason)))
	copy(payload[3:], reason)
	return payload
}

func runAuthServer(t *testing.T, secret, expectedClient string, response func(Packet) Packet) net.Conn {
	t.Helper()
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		hello, err := ReadPacket(server)
		if err != nil {
			return
		}
		clientLength := int(binary.BigEndian.Uint16(hello.Payload))
		if len(hello.Payload) != 2+clientLength || string(hello.Payload[2:]) != expectedClient {
			_ = WritePacket(server, response(Packet{Version: ProtocolVersion, Type: messageError, Payload: []byte{errorRejected}}))
			return
		}
		nonce := []byte("nonce-kept-byte-for-byte")
		serverTime := time.Now().Unix()
		_ = WritePacket(server, Packet{Version: ProtocolVersion, Type: messageChallenge, Payload: challengePayload(nonce, serverTime, serverTime/DefaultTimeStep)})
		auth, err := ReadPacket(server)
		if err != nil {
			return
		}
		nonceLength := int(binary.BigEndian.Uint16(auth.Payload))
		digestLengthOffset := 2 + nonceLength
		if len(auth.Payload) < digestLengthOffset+2 {
			return
		}
		digestLength := int(binary.BigEndian.Uint16(auth.Payload[digestLengthOffset:]))
		digestStart := digestLengthOffset + 2
		secretKey, _ := base64.StdEncoding.DecodeString(secret)
		counter := serverTime / DefaultTimeStep
		valid := nonceLength == len(nonce) && string(auth.Payload[2:2+nonceLength]) == string(nonce) && digestLength == sha256Size &&
			VerifyHMAC(CalculateHMAC(secretKey, expectedClient, counter, nonce), auth.Payload[digestStart:])
		if !valid {
			_ = WritePacket(server, Packet{Version: ProtocolVersion, Type: messageError, Payload: []byte{errorInvalidHMAC}})
			return
		}
		_ = WritePacket(server, response(Packet{Version: ProtocolVersion, Type: messageAuthSuccess, Payload: successPayload(true, "SUCCESS")}))
	}()
	return client
}

const sha256Size = 32

func TestAuthenticateSuccess(t *testing.T) {
	conn := runAuthServer(t, base64.StdEncoding.EncodeToString([]byte("secret")), "moses", func(packet Packet) Packet { return packet })
	defer conn.Close()
	result, err := AuthenticateWithConn(conn, "moses", []byte(base64.StdEncoding.EncodeToString([]byte("secret"))))
	if err != nil {
		t.Fatal(err)
	}
	if result.Reason != "SUCCESS" || result.Connection == nil {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestAuthenticateWrongSecret(t *testing.T) {
	conn := runAuthServer(t, base64.StdEncoding.EncodeToString([]byte("secret")), "moses", func(packet Packet) Packet { return packet })
	defer conn.Close()
	_, err := AuthenticateWithConn(conn, "moses", []byte(base64.StdEncoding.EncodeToString([]byte("wrong"))))
	if !errors.Is(err, ErrInvalidHMAC) {
		t.Fatalf("error = %v, want invalid HMAC", err)
	}
}

func TestAuthenticateWrongClientID(t *testing.T) {
	conn := runAuthServer(t, "secret", "moses", func(packet Packet) Packet { return packet })
	defer conn.Close()
	_, err := AuthenticateWithConn(conn, "other", []byte("secret"))
	if !errors.Is(err, ErrServerRejected) {
		t.Fatalf("error = %v, want server rejection", err)
	}
}

func TestAuthenticateUnsupportedVersion(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	go func() {
		defer server.Close()
		_, _ = ReadPacket(server)
		_ = WritePacket(server, Packet{Version: ProtocolVersion + 1, Type: messageChallenge, Payload: challengePayload([]byte("nonce"), time.Now().Unix(), time.Now().Unix()/DefaultTimeStep)})
	}()
	_, err := AuthenticateWithConn(client, "moses", []byte("secret"))
	if !errors.Is(err, ErrUnsupportedProtocol) {
		t.Fatalf("error = %v, want unsupported protocol", err)
	}
}

func TestAuthenticateReplayResponse(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	go func() {
		defer server.Close()
		_, _ = ReadPacket(server)
		_ = WritePacket(server, Packet{Version: ProtocolVersion, Type: messageError, Payload: []byte{errorReplay}})
	}()
	_, err := AuthenticateWithConn(client, "moses", []byte("secret"))
	if !errors.Is(err, ErrReplayDetected) {
		t.Fatalf("error = %v, want replay error", err)
	}
}

func TestAuthenticateTimeoutAndEmptySecret(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	_, err := authenticateWithConn(client, "moses", []byte("secret"), 10*time.Millisecond)
	if !errors.Is(err, ErrConnectionTimeout) {
		t.Fatalf("error = %v, want timeout", err)
	}
	_, err = AuthenticateWithConn(client, "moses", nil)
	if !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("error = %v, want authentication failure", err)
	}
}
