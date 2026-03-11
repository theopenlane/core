package keystore

import (
	"context"
	"errors"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

// credentialGetter retrieves the currently persisted credential set.
type credentialGetter func(ctx context.Context) (types.CredentialSet, error)

// credentialMinter mints/refreshed credential sets; previous is populated when refresh is triggered.
type credentialMinter func(ctx context.Context, previous types.CredentialSet) (types.CredentialSet, error)

// resolveCredentialWithPolicy standardizes credential resolution across client pools
func resolveCredentialWithPolicy(ctx context.Context, force bool, now func() time.Time, get credentialGetter, mint credentialMinter) (types.CredentialSet, error) {
	if force {
		return mint(ctx, types.CredentialSet{})
	}

	payload, err := get(ctx)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return mint(ctx, types.CredentialSet{})
		}

		return types.CredentialSet{}, err
	}

	if credentialNeedsRefresh(payload, now) {
		return mint(ctx, payload)
	}

	return payload, nil
}

// credentialNeedsRefresh determines if a credential needs refreshing based on OAuth expiry.
func credentialNeedsRefresh(payload types.CredentialSet, now func() time.Time) bool {
	if payload.OAuthExpiry == nil || payload.OAuthExpiry.IsZero() {
		return false
	}

	refreshAt := payload.OAuthExpiry.Add(-cacheSkew)

	return now().After(refreshAt)
}
