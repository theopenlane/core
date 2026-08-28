package graphapi

import (
	"context"
	"time"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
)

// WaitForIdentityHolderLink polls until the directory account has an identity_holder_id set or times out
func WaitForIdentityHolderLink(ctx context.Context, client *generated.Client, accountID string) (*generated.DirectoryAccount, error) {
	return WaitForIdentityHolderLinkWithTimeout(ctx, client, accountID, defaultPollTimeout)
}

// WaitForIdentityHolderLinkWithTimeout polls until the directory account has an identity_holder_id set or times out
func WaitForIdentityHolderLinkWithTimeout(ctx context.Context, client *generated.Client, accountID string, timeout time.Duration) (*generated.DirectoryAccount, error) {
	return pollUntil(ctx, timeout,
		func() (*generated.DirectoryAccount, error) {
			return client.DirectoryAccount.Query().Where(directoryaccount.IDEQ(accountID)).Only(ctx)
		},
		func(account *generated.DirectoryAccount) bool {
			return account.IdentityHolderID != nil && *account.IdentityHolderID != ""
		},
	)
}
