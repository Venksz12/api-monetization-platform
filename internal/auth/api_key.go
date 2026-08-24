package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrInvalidKey = errors.New("invalid api key")

func Generate() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plain = "amp_" + hex.EncodeToString(b)
	hash = Hash(plain)
	return
}

func Hash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func Normalize(key string) string {
	return strings.TrimSpace(key)
}
