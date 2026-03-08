package keystore

import (
	"context"
	"errors"
	"time"

	"github.com/theopenlane/core/common/models"
)

// credentialGetter retrieves the currently persisted credential set.
type credentialGetter func(ctx context.Context) (models.CredentialSet, error)

// credentialMinter mints/refreshed credential sets; previous is populated when refresh is triggered.
type credentialMinter func(ctx context.Context, previous models.CredentialSet) (models.CredentialSet, error)

// resolveCredentialWithPolicy standardizes credential resolution across client pools
func resolveCredentialWithPolicy(ctx context.Context, force bool, now func() time.Time, get credentialGetter, mint credentialMinter) (models.CredentialSet, error) {
	if force {
		return mint(ctx, models.CredentialSet{})
	}

	payload, err := get(ctx)
	if err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return mint(ctx, models.CredentialSet{})
		}

		return models.CredentialSet{}, err
	}

	if credentialNeedsRefresh(payload, now) {
		return mint(ctx, payload)
	}

	return payload, nil
}

// credentialNeedsRefresh determines if a credential needs refreshing based on OAuth expiry.
func credentialNeedsRefresh(payload models.CredentialSet, now func() time.Time) bool {
	if payload.OAuthExpiry == nil || payload.OAuthExpiry.IsZero() {
		return false
	}

	refreshAt := payload.OAuthExpiry.Add(-cacheSkew)

	return now().After(refreshAt)
}
