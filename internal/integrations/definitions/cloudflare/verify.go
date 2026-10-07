package cloudflare

import (
	"context"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/accounts"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify probes the Cloudflare token and returns the account metadata it is scoped to
func verify(ctx context.Context, _ types.ConnectionRequest[CredentialSchema], c *CloudflareClient) (InstallationMetadata, error) {
	res, err := c.Accounts.Tokens.Verify(ctx, accounts.TokenVerifyParams{
		AccountID: cf.F(c.Config.AccountID),
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: token verification failed")
		return InstallationMetadata{}, ErrTokenVerificationFailed
	}

	if res.Status != accounts.TokenVerifyResponseStatusActive {
		return InstallationMetadata{}, ErrTokenNotActive
	}

	return InstallationMetadata{AccountID: c.Config.AccountID}, nil
}
