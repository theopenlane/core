package outofscope

import (
	"context"

	"entgo.io/ent/privacy"
)

func bare(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow)
}
