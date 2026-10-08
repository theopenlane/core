package listener

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/theopenlane/iam/auth"
)

func bare(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow) // want "privacy.Allow decision contexts are not allowed"
}

func allowf(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allowf("cleanup")) // want "privacy.Allow decision contexts are not allowed"
}

func assignedThenWrapped(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow) // want "privacy.Allow decision contexts are not allowed"

	return auth.WithCaller(allowCtx, &auth.Caller{})
}

func withCaller(ctx context.Context) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{OrganizationID: "org"}) // want "privacy.Allow decision contexts are not allowed"
}

func integrationCaller(ctx context.Context) context.Context {
	return auth.EnsureIntegrationCaller(privacy.DecisionContext(ctx, privacy.Allow), "org") // want "privacy.Allow decision contexts are not allowed"
}
