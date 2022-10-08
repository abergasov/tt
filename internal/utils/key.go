package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateKey_A(src string) string {
	hash := sha256.Sum256([]byte(src))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func GenerateKey_B(src string) string {
	hash := sha256.Sum256([]byte(src))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateKey hex encoding little faster than base64 and gives more friendly output
func GenerateKey(src string) string {
	hash := sha256.Sum256([]byte(src))
	return hex.EncodeToString(hash[:])
}
