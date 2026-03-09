package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomString generates a random bytes of the given length, encoded to base64.
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Errorf("failed to generate random string: %w", err))
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
