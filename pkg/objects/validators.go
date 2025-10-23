package objects

import (
	"fmt"
	"strings"
)

// ValidationFunc is a type that can be used to dynamically validate a file
type ValidationFunc func(f File) error

// MimeTypeValidator makes sure we only accept a valid mimetype.
// It takes in an array of supported mimes
// MimeTypeValidator is a validator factory that ensures the file's content type matches one of the provided types
// When validation fails it wraps ErrUnsupportedMimeType and includes a normalized mime type without charset parameters
func MimeTypeValidator(validMimeTypes ...string) ValidationFunc {
	return func(f File) error {
		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(strings.ToLower(mimeType), f.ContentType) {
				return nil
			}
		}

		return fmt.Errorf("%w: %s", ErrUnsupportedMimeType, f.ContentType)
	}
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
