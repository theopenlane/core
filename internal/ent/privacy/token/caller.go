package token

import (
	"context"

	"github.com/theopenlane/iam/auth"
)

const tokenContextBypassCaps = auth.CapBypassOrgFilter

func withTokenContextBypassCaller(ctx context.Context) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	return auth.WithCaller(ctx, caller.WithCapabilities(tokenContextBypassCaps))
}
