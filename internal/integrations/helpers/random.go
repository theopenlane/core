package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// RandomState generates a URL-safe random string using crypto/rand
func RandomState(bytes int) (string, error) {
	if bytes <= 0 {
		bytes = 32
	}

	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("integrations/helpers: random state: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
