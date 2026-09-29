package privacy

import (
	"context"
	"errors"
)

var Allow = errors.New("allow")

func DecisionContext(ctx context.Context, _ error) context.Context {
	return ctx
}
