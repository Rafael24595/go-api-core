package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenerateRawToken() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func HashToken(secret, token string) string {
	tkn := fmt.Sprintf("%s:%s", secret, token)
	sum := sha256.Sum256([]byte(tkn))
	return hex.EncodeToString(sum[:])
}
