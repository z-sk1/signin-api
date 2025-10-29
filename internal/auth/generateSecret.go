package auth

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

func getOrCreateSecret() []byte {
	data, err := os.ReadFile("secret.txt")
	if err == nil {
		return data
	}

	// Create one if not found
	b := make([]byte, 64)
	rand.Read(b)
	hexSecret := hex.EncodeToString(b)
	os.WriteFile("internal/auth/secret.txt", []byte(hexSecret), 0600)
	return []byte(hexSecret)
}
