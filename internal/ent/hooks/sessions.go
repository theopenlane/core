package hooks

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/logx"
)

// updateUserAuthSession updates the user session with the new org ID
// and sets updated auth cookies
func updateUserAuthSession(ctx context.Context, am *authmanager.Client, newOrgID string) error {
	if am == nil {
		logx.FromContext(ctx).Error().Msg("auth manager is nil, unable to update user auth session")

		return ErrInternalServerError
	}

	sessionCaller, ok := auth.CallerFromContext(ctx)
	if !ok || sessionCaller == nil {
		return auth.ErrNoAuthUser
	}

	user, err := generated.FromContext(ctx).User.
		Query().
		WithSetting().
		Where(user.ID(sessionCaller.SubjectID)).
		Only(ctx)
	if err != nil {
		return err
	}

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	// generate a new auth session with the new org ID
	// this will also set the session cookie
	out, err := am.GenerateUserAuthSessionWithOrg(ctx, ec.Response().Writer, user, newOrgID)
	if err != nil {
		return err
	}

	// add the organization ID to the authenticated user context
	sessionCaller.OrganizationID = newOrgID

	err = auth.SetOrganizationIDInAuthContext(ctx, newOrgID)
	if err != nil {
		return err
	}

	// set the auth cookies
	auth.SetAuthCookies(ec.Response().Writer, out.AccessToken, out.RefreshToken, *am.GetSessionConfig().CookieConfig)

	// update the context with the new tokens and session
	auth.WithContextValue(ctx, auth.AccessTokenKey, out.AccessToken)
	auth.WithContextValue(ctx, auth.RefreshTokenKey, out.RefreshToken)

	return err
}
