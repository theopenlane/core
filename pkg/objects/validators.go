package objects

import (
	"fmt"
	"mime"
	"strings"
)

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f File) error

// MimeTypeValidator returns a validator that accepts a file when its content type
// matches one of validMimeTypes. Both sides are normalized: structured-parameter
// suffixes (e.g. "; charset=utf-8") are stripped before comparison so an allowlist
// entry of "text/html" matches an incoming "text/html; charset=utf-8" returned by
// MIME sniffers.
func MimeTypeValidator(validMimeTypes ...string) ValidationFunc {
	return func(f File) error {
		got := normalizeMimeType(f.ContentType)

		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(normalizeMimeType(mimeType), got) {
				return nil
			}
		}

		return fmt.Errorf("%w: %s", ErrUnsupportedMimeType, f.ContentType)
	}
}

// normalizeMimeType strips any media-type parameters (e.g. "; charset=utf-8") and
// returns a lowercased "type/subtype". When the value isn't parseable as a media
// type, it falls back to the lowercased input so validation degrades to a strict
// equality compare rather than silently accepting.
func normalizeMimeType(s string) string {
	t, _, err := mime.ParseMediaType(s)
	if err != nil {
		return strings.ToLower(s)
	}

	return t
}

// ChainValidators returns a validator that accepts multiple validating criteria
func ChainValidators(validators ...ValidationFunc) ValidationFunc {
	return func(f File) error {
		for _, validator := range validators {
			if err := validator(f); err != nil {
				return err
			}
		}

		return nil
	}
}
