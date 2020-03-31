package token

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRefresh will generate a cryptographically random base64 string
func GenerateRefresh(length int) (string, error) {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("could not generate rand bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
