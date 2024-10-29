package internal

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidLink   = errors.New("invalid link")
	ErrFailedRequest = errors.New("failed to create request")
	ErrFailedGetLink = errors.New("failed to get link")
)
