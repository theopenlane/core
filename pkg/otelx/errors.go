package otelx

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownProvider = errors.New("unknown provider")
	ErrInvalidConfig   = errors.New("failed parsing trace config(s)")
)

func newUnknownProviderError(provider string) error {
	return fmt.Errorf("%w: %s", ErrUnknownProvider, provider)
}

func newTraceConfigError(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidConfig, err)
}
