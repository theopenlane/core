package serveropts

import "errors"

// ErrNoSigningKeys is returned when no signing keys are found in the key directory
var ErrNoSigningKeys = errors.New("no signing keys found in key directory")
