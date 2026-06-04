package googledrive

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives Google Drive installation metadata from the credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := driveCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to resolve drive credential")

		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cred.AccessToken})

	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to create drive service")

		return InstallationMetadata{}, false, nil
	}

	about, err := svc.About.Get().Fields("user(emailAddress,displayName)").Context(ctx).Do()
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googledrive: failed to fetch about information")

		return InstallationMetadata{}, false, nil
	}

	if about.User == nil || about.User.EmailAddress == "" {
		return InstallationMetadata{}, false, nil
	}

	meta := InstallationMetadata{
		Domain: domainFromEmail(about.User.EmailAddress),
	}

	return meta, true, nil
}

// domainFromEmail extracts the domain portion from an email address
func domainFromEmail(email string) string {
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			return email[i+1:]
		}
	}

	return email
}
