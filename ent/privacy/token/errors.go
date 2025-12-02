package token

import (
	"errors"
)

var (
	ErrIncorrectEmail   = errors.New("privacy token has incorrect email")
	ErrInvalidTokenType = errors.New("invalid token type")
)
