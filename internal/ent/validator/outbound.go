package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/core/v2/pkg/urlx"
)

var (
	// ErrURLNotPublic is returned when a url targets a private, loopback, link-local, or internal host
	ErrURLNotPublic = errors.New("url must be a public address")
	// ErrHeaderNotAllowed is returned when a request header is reserved for cloud metadata services
	ErrHeaderNotAllowed = errors.New("header is not allowed")
)

// blockedOutboundHeaders are lowercased request headers that cloud metadata services require
var blockedOutboundHeaders = map[string]struct{}{
	"metadata-flavor":                      {},
	"x-google-metadata-request":            {},
	"metadata":                             {},
	"x-aws-ec2-metadata-token":             {},
	"x-aws-ec2-metadata-token-ttl-seconds": {},
}

// ValidatePublicURL validates that a url the server will call out to targets a public address
func ValidatePublicURL() func(u string) error {
	return func(u string) error {
		if _, err := urlx.ValidatePublicURL(u); err != nil {
			return fmt.Errorf("%w: %w", ErrURLNotPublic, err)
		}

		return nil
	}
}

// ValidateOutboundHeaders validates that headers the server will send do not include cloud metadata headers
func ValidateOutboundHeaders() func(headers map[string]string) error {
	return func(headers map[string]string) error {
		for name := range headers {
			if _, blocked := blockedOutboundHeaders[strings.ToLower(strings.TrimSpace(name))]; blocked {
				return fmt.Errorf("%w: %s", ErrHeaderNotAllowed, name)
			}
		}

		return nil
	}
}
