package genhelpers

import "errors"

// ErrLoadPackages is returned when Go packages cannot be loaded for comment extraction
var ErrLoadPackages = errors.New("failed to load packages")
