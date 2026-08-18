package urlx

import (
	"net/url"
	"strings"
)

const (
	// schemeHTTP is the plaintext web scheme
	schemeHTTP = "http"
	// schemeHTTPS is the TLS web scheme applied when input lacks a scheme
	schemeHTTPS = "https"
	// schemeSeparator splits a scheme from the remainder of a URL
	schemeSeparator = "://"
)

// Parse parses rawURL after trimming surrounding whitespace, applying the https
// scheme when rawURL has none, and requires the result to contain a hostname and
// an http or https scheme
func Parse(rawURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, ErrEmptyURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Hostname() == "" {
		if strings.Contains(trimmed, schemeSeparator) {
			if err != nil {
				return nil, ErrInvalidURL
			}

			return nil, ErrMissingHost
		}

		if parsed, err = url.Parse(schemeHTTPS + schemeSeparator + trimmed); err != nil {
			return nil, ErrInvalidURL
		}
	}

	if parsed.Hostname() == "" {
		return nil, ErrMissingHost
	}

	if parsed.Scheme == "" {
		parsed.Scheme = schemeHTTPS
	}

	if !isHTTPScheme(parsed.Scheme) {
		return nil, ErrUnsupportedScheme
	}

	return parsed, nil
}

// ParseAbsolute parses rawURL after trimming surrounding whitespace, requiring an
// explicit http or https scheme and a host
func ParseAbsolute(rawURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, ErrEmptyURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, ErrInvalidURL
	}

	if !isHTTPScheme(parsed.Scheme) {
		return nil, ErrUnsupportedScheme
	}

	if parsed.Host == "" {
		return nil, ErrMissingHost
	}

	return parsed, nil
}

// NormalizeHostname extracts the hostname from a URL or raw host input,
// lowercasing it and stripping any trailing dots
func NormalizeHostname(rawURL string) (string, error) {
	parsed, err := Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := strings.TrimRight(strings.ToLower(parsed.Hostname()), ".")
	if host == "" {
		return "", ErrMissingHost
	}

	return host, nil
}

// WithoutQuery returns rawURL with its query string and fragment removed,
// returning rawURL unchanged when it cannot be parsed
func WithoutQuery(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String()
}

// isHTTPScheme reports whether scheme is http or https
func isHTTPScheme(scheme string) bool {
	return scheme == schemeHTTP || scheme == schemeHTTPS
}
