package emailbranding

import "errors"

// ErrNameRequired is returned when --name is missing
var ErrNameRequired = errors.New("--name is required")
