package hooks

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
)

// newAuthSession creates a new auth session struct
func newAuthSession(db *generated.Client) authmanager.Config {
	return *authmanager.New(db)
}

// updateUserAuthSession updates the user session with the new org ID
// and sets updated auth cookies
func updateUserAuthSession(ctx context.Context, as authmanager.Config, newOrgID string) error {
	au, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	user, err := generated.FromContext(ctx).User.
		Query().
		WithSetting().
		Where(user.ID(au.SubjectID)).
		Only(ctx)
	if err != nil {
		return err
	}

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	ec.SetRequest(ec.Request().WithContext(ctx))

	// generate a new auth session with the new org ID
	// this will also set the session cookie
	out, err := as.GenerateUserAuthSessionWithOrg(ec, user, newOrgID)
	if err != nil {
		return err
	}

	// add the organization ID to the authenticated user context
	if err := auth.SetOrganizationIDInAuthContext(ctx, newOrgID); err != nil {
		return err
	}

	// set the auth cookies
	auth.SetAuthCookies(ec.Response().Writer, out.AccessToken, out.RefreshToken, *as.GetSessionConfig().CookieConfig)

	// update the context with the new tokens and session
	auth.SetAccessTokenContext(ec, out.AccessToken)
	auth.SetRefreshTokenContext(ec, out.RefreshToken)

	return err
}
