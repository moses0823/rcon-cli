package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

func CalculateHMAC(secret []byte, clientID string, timeCounter int64, nonce []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(clientID))
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(timeCounter))
	mac.Write(counter[:])
	mac.Write(nonce)
	return mac.Sum(nil)
}

func VerifyHMAC(expected, actual []byte) bool {
	return hmac.Equal(expected, actual)
}
