package tailscale

import (
	"context"

	"github.com/samber/lo"
	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify derives the Tailscale tailnet metadata from the credential's member users
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *tsclient.Client) (InstallationMetadata, error) {
	tailnet, err := resolveTailnet(ctx, c)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		ClientID: req.Credential.ClientID,
		Tailnet:  tailnet,
	}, nil
}

// resolveTailnet derives the tailnet name from the credential's member users
func resolveTailnet(ctx context.Context, client *tsclient.Client) (string, error) {
	memberType := tsclient.UserTypeMember

	users, err := client.Users().List(ctx, &memberType, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("tailscale: failed listing users to resolve tailnet")
		return "", ErrUsersFetchFailed
	}

	user, ok := lo.Find(users, func(u tsclient.User) bool { return u.TailnetID != "" })
	if !ok {
		return "", ErrTailnetUnresolved
	}

	return user.TailnetID, nil
}
