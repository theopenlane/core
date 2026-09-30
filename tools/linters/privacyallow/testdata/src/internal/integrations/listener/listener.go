package listener

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/theopenlane/iam/auth"
)

func bare(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow) // want "privacy.Allow must be wrapped in a scoped caller"
}

func assignedThenWrapped(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow) // want "privacy.Allow must be wrapped in a scoped caller"

	return auth.WithCaller(allowCtx, &auth.Caller{})
}

func withCaller(ctx context.Context) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{OrganizationID: "org"})
}

func integrationCaller(ctx context.Context) context.Context {
	return auth.EnsureIntegrationCaller(privacy.DecisionContext(ctx, privacy.Allow), "org")
}
