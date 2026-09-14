package security

import (
	"encoding/binary"
	"fmt"
	"io"
)

const headerSize = 11

var magic = [4]byte{'R', 'S', 'E', 'C'}

func WritePacket(w io.Writer, packet Packet) error {
	if uint32(len(packet.Payload)) > MaxPacketSize {
		return fmt.Errorf("%w: packet too large", ErrInvalidPacket)
	}
	header := make([]byte, headerSize)
	copy(header[:4], magic[:])
	binary.BigEndian.PutUint16(header[4:], packet.Version)
	header[6] = packet.Type
	binary.BigEndian.PutUint32(header[7:], uint32(len(packet.Payload)))
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(packet.Payload)
	return err
}

func ReadPacket(r io.Reader) (Packet, error) {
	var header [headerSize]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return Packet{}, fmt.Errorf("%w: %v", ErrInvalidPacket, err)
	}
	if string(header[:4]) != string(magic[:]) {
		return Packet{}, fmt.Errorf("%w: invalid magic", ErrInvalidPacket)
	}
	version := binary.BigEndian.Uint16(header[4:])
	typeByte := header[6]
	length := binary.BigEndian.Uint32(header[7:])
	if length > MaxPacketSize {
		return Packet{}, fmt.Errorf("%w: packet too large", ErrInvalidPacket)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Packet{}, fmt.Errorf("%w: %v", ErrInvalidPacket, err)
	}
	return Packet{Version: version, Type: typeByte, Payload: payload}, nil
}
