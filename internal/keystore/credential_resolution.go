package keystore

import (
	"context"
	"errors"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

// credentialGetter retrieves the currently persisted credential payload
type credentialGetter func(ctx context.Context) (types.CredentialPayload, error)

// credentialMinter mints/refreshed a credential payload; previous is populated when refresh is triggered
type credentialMinter func(ctx context.Context, previous types.CredentialPayload) (types.CredentialPayload, error)

// resolveCredentialWithPolicy standardizes credential resolution across client pools
func resolveCredentialWithPolicy(ctx context.Context, force bool, now func() time.Time, get credentialGetter, mint credentialMinter) (types.CredentialPayload, error) {
	if force {
		return mint(ctx, types.CredentialPayload{})
	}

	payload, err := get(ctx)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return mint(ctx, types.CredentialPayload{})
		}

		return types.CredentialPayload{}, err
	}

	if credentialNeedsRefresh(payload, now) {
		return mint(ctx, payload)
	}

	return payload, nil
}

// credentialNeedsRefresh determines if a credential needs refreshing based on its expiry
func credentialNeedsRefresh(payload types.CredentialPayload, now func() time.Time) bool {
	if payload.Token == nil || payload.Token.Expiry.IsZero() {
		return false
	}
	if now == nil {
		now = time.Now
	}

	refreshAt := payload.Token.Expiry.Add(-cacheSkew)

	return now().After(refreshAt)
}
