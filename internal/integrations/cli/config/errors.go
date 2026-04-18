package config

import (
	"slices"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// IsAlreadyExistsError checks if a GraphQL error response indicates an
// ALREADY_EXISTS or USER_EXISTS error
func IsAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	if HasAnyErrorCode(err, "ALREADY_EXISTS", "USER_EXISTS") {
		return true
	}

	msg := err.Error()

	return strings.Contains(msg, "already exists") || strings.Contains(msg, "one trust center at a time")
}

// HasAnyErrorCode checks if a GraphQL error response includes any of the provided codes
func HasAnyErrorCode(err error, codes ...string) bool {
	if err == nil || len(codes) == 0 {
		return false
	}

	gqlErr, ok := err.(*clientv2.ErrorResponse)
	if !ok {
		return false
	}

	return lo.SomeBy(*gqlErr.GqlErrors, func(e *gqlerror.Error) bool {
		code, ok := e.Extensions["code"].(string)
		return ok && slices.Contains(codes, code)
	})
}
