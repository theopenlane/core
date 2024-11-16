package plan

import (
	"errors"
)

var (
	// ErrNotFound is returned when the plan is not found
	ErrNotFound = errors.New("plan not found")
	// ErrInvalidUUID is returned when the syntax of uuid is invalid
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
	// ErrInvalidName is returned when the plan name is invalid
	ErrInvalidName = errors.New("plan name is invalid")
	// ErrInvalidDetail is returned when the plan detail is invalid
	ErrInvalidDetail = errors.New("invalid plan detail")
)
