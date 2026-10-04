package privacy

import (
	"context"
	"errors"
	"fmt"
)

var Allow = errors.New("allow")

func Allowf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

func DecisionContext(ctx context.Context, _ error) context.Context {
	return ctx
}
