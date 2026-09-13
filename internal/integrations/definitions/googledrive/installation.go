package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/ssoutils"
)

// installationRef builds the typed installation metadata handle for the Google Drive definition,
// closing over the operator OAuth config needed to refresh the stored credential's access token
func installationRef(cfg Config) types.InstallationRef[InstallationMetadata] {
	return types.NewInstallationRef(func(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
		return resolveInstallationMetadata(ctx, cfg, req)
	})
}

// resolveInstallationMetadata derives Google Drive installation metadata from the credential
func resolveInstallationMetadata(ctx context.Context, cfg Config, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := driveCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to resolve drive credential")

		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	svc, err := drive.NewService(ctx, option.WithTokenSource(tokenSource(ctx, cfg, cred)))
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to create drive service")

		return InstallationMetadata{}, false, nil
	}

	about, err := svc.About.Get().Fields("user(emailAddress,permissionId)").Context(ctx).Do()
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to fetch about information")

		return InstallationMetadata{}, false, nil
	}

	if about.User == nil || about.User.EmailAddress == "" || about.User.PermissionId == "" {
		return InstallationMetadata{}, false, nil
	}

	meta := InstallationMetadata{
		AccountID: about.User.PermissionId,
		Domain:    ssoutils.EmailDomain(about.User.EmailAddress),
	}

	return meta, true, nil
}
