package authz

import (
	"context"

	"entgo.io/ent/privacy"
)

func generated(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow)
}
