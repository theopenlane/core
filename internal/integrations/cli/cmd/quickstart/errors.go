//go:build examples

package quickstart

import "errors"

// ErrRecipientResolution is returned when no target email can be resolved from
// either --to or the authenticated user config
var ErrRecipientResolution = errors.New("no recipient resolved; set --to or openlane.auth.email")
