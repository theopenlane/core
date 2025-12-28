package domain

import (
	"errors"
	"net/url"
	"strings"
)

var (
	// ErrEmptyHostname indicates the input does not contain a hostname
	ErrEmptyHostname = errors.New("hostname is required")
	// ErrInvalidHostname indicates the input could not be parsed into a hostname
	ErrInvalidHostname = errors.New("invalid hostname")
)

// NormalizeHostname extracts the hostname from a URL or raw host input, lowercases it,
// and removes trailing dots
func NormalizeHostname(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", ErrEmptyHostname
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", ErrInvalidHostname
	}

	host := parsed.Hostname()
	if host == "" {
		if strings.Contains(trimmed, "://") {
			return "", ErrInvalidHostname
		}

		parsed, err = url.Parse("http://" + trimmed)
		if err != nil {
			return "", ErrInvalidHostname
		}

		host = parsed.Hostname()
	}

	host = strings.TrimRight(strings.ToLower(host), ".")
	if host == "" {
		return "", ErrInvalidHostname
	}

	return host, nil
}
