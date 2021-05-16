package fileshare

import (
	"crypto/rand"
	"encoding/hex"
)

func Bytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func Hex(n int) string {
	return hex.EncodeToString(Bytes(n))
}
