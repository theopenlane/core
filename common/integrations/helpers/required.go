package helpers

import "strings"

// RequireString returns err when the provided value is empty after trimming.
func RequireString(value string, err error) error {
	if strings.TrimSpace(value) == "" {
		return err
	}
	return nil
}

// RequiredString returns the trimmed string value for a key or err when missing.
func RequiredString(data map[string]any, key string, err error) (string, error) {
	if len(data) == 0 {
		return "", err
	}
	value, ok := data[key]
	if !ok {
		return "", err
	}
	trimmed := StringFromAny(value)
	if trimmed == "" {
		return "", err
	}
	return trimmed, nil
}
