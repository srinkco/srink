package randomiser

import (
	"crypto/rand"
	"encoding/hex"
)

func GetString(digit int) string {
	buf := make([]byte, digit)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func GetBytes(digit int) []byte {
	src := make([]byte, digit)
	rand.Read(src)
	dst := make([]byte, hex.EncodedLen(digit))
	hex.Encode(dst, src)
	return dst
}
