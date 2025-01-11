package schema

import (
	"errors"
	"fmt"

	"entgo.io/ent"
)

var (
	// ErrInvalidTokenSize is returned when session token size is invalid
	ErrInvalidTokenSize = errors.New("invalid token size")

	// ErrContainsSpaces is returned when field contains spaces
	ErrContainsSpaces = errors.New("field should not contain spaces")

	// ErrUnexpectedMutationType is returned when an unexpected mutation type is encountered
	ErrUnexpectedMutationType = errors.New("unexpected mutation type")
)

// UnexpectedAuditError is returned when an unexpected audit log call is received
type UnexpectedAuditError struct {
	MutationType ent.Mutation
}

// Error returns the UnexpectedAuditError in string format
func (e *UnexpectedAuditError) Error() string {
	return fmt.Sprintf("unexpected audit log call from mutation type: %T", e.MutationType)
}

func newUnexpectedAuditError(arg ent.Mutation) *UnexpectedAuditError {
	return &UnexpectedAuditError{
		MutationType: arg,
	}
}
