package security

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestPacketRoundTrip(t *testing.T) {
	input := Packet{Version: ProtocolVersion, Type: messageHello, Payload: []byte("moses")}
	var buffer bytes.Buffer
	if err := WritePacket(&buffer, input); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPacket(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != input.Version || got.Type != input.Type || !bytes.Equal(got.Payload, input.Payload) {
		t.Fatalf("packet = %#v, want %#v", got, input)
	}
}

func TestReadPacketRejectsOversizedPacket(t *testing.T) {
	header := make([]byte, headerSize)
	copy(header, magic[:])
	binary.BigEndian.PutUint32(header[6:], MaxPacketSize+1)
	_, err := ReadPacket(bytes.NewReader(header))
	if !errors.Is(err, ErrInvalidPacket) {
		t.Fatalf("error = %v, want ErrInvalidPacket", err)
	}
}

func TestReadPacketRejectsMalformedPacket(t *testing.T) {
	_, err := ReadPacket(bytes.NewReader([]byte("bad packet")))
	if !errors.Is(err, ErrInvalidPacket) {
		t.Fatalf("error = %v, want malformed packet error", err)
	}
}
