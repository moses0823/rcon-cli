package security

import (
	"encoding/hex"
	"testing"
)

func TestCalculateHMAC(t *testing.T) {
	nonce := []byte("server-nonce")
	got := CalculateHMAC([]byte("secret"), "moses", 57, nonce)
	want, _ := hex.DecodeString("ce37ef96492f6cbb077a794805b37603a35bcb021ec9cc11339176d071438389")
	if string(got) != string(want) {
		t.Fatalf("HMAC = %x, want %x", got, want)
	}
	if !VerifyHMAC(want, got) || VerifyHMAC(want, []byte("invalid")) {
		t.Fatal("HMAC verification did not use the expected digest")
	}
}

func TestCalculateHMACChangesWithNonce(t *testing.T) {
	first := CalculateHMAC([]byte("secret"), "moses", 57, []byte("nonce-a"))
	second := CalculateHMAC([]byte("secret"), "moses", 57, []byte("nonce-b"))
	if VerifyHMAC(first, second) {
		t.Fatal("changing nonce must change the HMAC")
	}
}
