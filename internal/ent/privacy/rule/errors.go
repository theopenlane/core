package rule

import "errors"

var (
	ErrRequiredScopeNotSet = errors.New("the provided token does not have the required scopes for the request")
)
