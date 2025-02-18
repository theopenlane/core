package hooks

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
)

// newAuthSession creates a new auth session struct
func newAuthSession(sc *sessions.SessionConfig, tm *tokens.TokenManager) authmanager.Config {
	as := authmanager.Config{}

	as.SetSessionConfig(sc)
	as.SetTokenManager(tm)

	return as
}

// updateUserAuthSession updates the user session with the new org ID
// and sets updated auth cookies
func updateUserAuthSession(ctx context.Context, as authmanager.Config, newOrgID string) error {
	au, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	user, err := generated.FromContext(ctx).User.Get(ctx, au.SubjectID)
	if err != nil {
		return err
	}

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	// generate a new auth session with the new org ID
	// this will also set the session cookie
	out, err := as.GenerateUserAuthSessionWithOrg(ec, user, newOrgID)
	if err != nil {
		return err
	}

	// // add the organization ID to the authenticated user context
	// if err := auth.SetOrganizationIDInAuthContext(ctx, newOrgID); err != nil {
	// 	return err
	// }

	// set the auth cookies
	auth.SetAuthCookies(ec.Response().Writer, out.AccessToken, out.RefreshToken, *as.GetSessionConfig().CookieConfig)

	// update the context with the new tokens and session
	auth.SetAccessTokenContext(ec, out.AccessToken)
	auth.SetRefreshTokenContext(ec, out.RefreshToken)

	return err
}
