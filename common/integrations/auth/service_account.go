package auth

import (
	"encoding/json"
	"strings"
)

// NormalizeServiceAccountKey trims and unwraps JSON-encoded service account keys.
func NormalizeServiceAccountKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}
