package auth

import "context"

type Caller struct {
	OrganizationID string
}

func WithCaller(ctx context.Context, _ *Caller) context.Context {
	return ctx
}

func EnsureIntegrationCaller(ctx context.Context, _ string) context.Context {
	return ctx
}
