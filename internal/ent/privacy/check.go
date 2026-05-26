package access

import (
	"errors"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// Allow is a wrapper around privacy decision errors that will return true for privacy.Allow
func Allow(err error) bool {
	return errors.Is(err, privacy.Allow)
}

// Deny is a wrapper around privacy decision errors that will return true for privacy.Deny
func Deny(err error) bool {
	return errors.Is(err, privacy.Deny)
}
