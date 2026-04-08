package controls

import "errors"

var (
	// ErrStandardNotFound is the error message returned when a standard cannot be found based on the provided filter options during control cloning
	ErrStandardNotFound = errors.New("standard not found, unable to clone controls")
)
