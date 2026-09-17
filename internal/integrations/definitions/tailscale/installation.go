package tailscale

import (
	"context"

	"github.com/samber/lo"
	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// resolveInstallationMetadata derives Tailscale tailnet metadata from the installation credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	built, err := Client{}.Build(ctx, types.ClientBuildRequest{Credentials: req.Credentials})
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	tailnet, err := resolveTailnet(ctx, built.(*tsclient.Client))
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		ClientID: cred.ClientID,
		Tailnet:  tailnet,
	}, true, nil
}

// resolveTailnet derives the tailnet name the credential is scoped to from the member users it can list
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
