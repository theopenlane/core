package mixin

import (
	"fmt"

	"entgo.io/ent"
)

// UnexpectedMutationTypeError is returned when an unexpected mutation type is parsed
type UnexpectedMutationTypeError struct {
	MutationType ent.Mutation
}

// Error returns the UnexpectedAuditError in string format
func (e *UnexpectedMutationTypeError) Error() string {
	return fmt.Sprintf("unexpected mutation type: %T", e.MutationType)
}

func newUnexpectedMutationTypeError(arg ent.Mutation) *UnexpectedMutationTypeError {
	return &UnexpectedMutationTypeError{
		MutationType: arg,
	}
}
