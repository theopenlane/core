package googledrive

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/ssoutils"
)

// verify resolves the connected Google account through the Drive About API
func verify(ctx context.Context, _ types.ConnectionRequest[googleDriveCred], c DriveClient) (InstallationMetadata, error) {
	about, err := c.Svc.About.Get().Fields("user(emailAddress,permissionId)").Context(ctx).Do()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("googledrive: about lookup failed")

		return InstallationMetadata{}, fmt.Errorf("%w: %v", ErrHealthCheckFailed, err)
	}

	if about.User == nil || about.User.EmailAddress == "" || about.User.PermissionId == "" {
		return InstallationMetadata{}, ErrAccountUnresolved
	}

	return InstallationMetadata{
		AccountID: about.User.PermissionId,
		Domain:    ssoutils.EmailDomain(about.User.EmailAddress),
	}, nil
}
