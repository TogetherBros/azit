package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func NewRefreshToken() (plain string, hash string) {
	b := make([]byte, 32)
	rand.Read(b)

	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, HashRefreshToken(plain)
}

func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
