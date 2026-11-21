package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	iamoauth "github.com/theopenlane/iam/oauthserver"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/oauthserver"
	"github.com/theopenlane/core/v2/internal/oauthserver/store"
	"github.com/theopenlane/core/v2/pkg/middleware/graphapi"
)

// oauthAccessTokenAuthentication is what getTokenType returns for OAuth access tokens, the caller still gets PAT or API token authentication
const oauthAccessTokenAuthentication auth.AuthenticationType = "oauth_access_token" //nolint:gosec

var (
	// ErrOAuthUserNotMember is returned when a personal connection's user has left the organization
	ErrOAuthUserNotMember = errors.New("the user who connected this application is no longer a member of the organization")
	// ErrOAuthObjectLimited is returned when an application limited to specific objects calls outside the GraphQL API
	ErrOAuthObjectLimited = errors.New("this application is limited to specific objects and can only use the GraphQL API")
)

// checkOAuthAccessToken validates an OAuth access token and returns the caller and grant for the request
// personal connections act as their user like a PAT, organization installs act as a service subject like an API token
func checkOAuthAccessToken(ctx context.Context, conf *Options, token string) (*auth.Caller, oauthserver.CallerGrant, error) {
	var none oauthserver.CallerGrant

	// the OAuth authorization server is disabled, so no OAuth token can be valid
	if conf.oauthTokens == nil {
		return nil, none, rout.ErrInvalidCredentials
	}

	internalCtx := auth.WithInternalReadCrossOrgContext(ctx)

	t, err := store.Redeem(internalCtx, conf.DBClient, conf.oauthTokens, token, iamoauth.KindAccess)

	switch {
	case errors.Is(err, iamoauth.ErrTokenInvalid):
		return nil, none, rout.ErrInvalidCredentials
	case errors.Is(err, store.ErrTokenConsumed):
		return nil, none, ErrTokenRevoked
	case errors.Is(err, store.ErrTokenExpired):
		return nil, none, rout.ErrExpiredCredentials
	case err != nil:
		return nil, none, err
	}

	grant := t.Edges.OauthGrant

	marker := oauthserver.CallerGrant{
		ID:       grant.ID,
		Kind:     grant.Kind,
		ClientID: grant.ClientID,
		Access:   enums.OAuthAccessLevelWrite,
	}

	var caller *auth.Caller

	switch grant.Kind {
	case enums.OAuthGrantKindUser:
		// read from the grant, not the token, so a narrower re-consent applies at once
		marker.Access = grant.AccessLevel
		marker.Scopes = grant.Scopes

		role := getOrgRoleFunc(internalCtx, conf.DBClient, grant.UserID, grant.OwnerID)
		if role == nil {
			return nil, none, ErrOAuthUserNotMember
		}

		user, err := conf.DBClient.User.Get(internalCtx, grant.UserID)
		if err != nil {
			return nil, none, rout.ErrInvalidCredentials
		}

		// applications never inherit system admin capabilities
		caller = &auth.Caller{
			SubjectID:          user.ID,
			SubjectName:        getSubjectName(user),
			SubjectEmail:       user.Email,
			OrganizationID:     grant.OwnerID,
			OrganizationIDs:    []string{grant.OwnerID},
			OrganizationRole:   *role,
			AuthenticationType: auth.PATAuthentication,
		}
	case enums.OAuthGrantKindOrganization:
		caller = &auth.Caller{
			SubjectID:          grant.ID,
			SubjectName:        fmt.Sprintf("app: %s", grant.ClientName),
			OrganizationID:     grant.OwnerID,
			OrganizationIDs:    []string{grant.OwnerID},
			AuthenticationType: auth.APITokenAuthentication,
		}
	default:
		return nil, none, rout.ErrInvalidCredentials
	}

	// usage tracking must not block the request
	if err := conf.DBClient.OauthGrant.UpdateOneID(grant.ID).SetLastUsedAt(time.Now()).Exec(auth.WithInternalOperationContext(ctx)); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("grant_id", grant.ID).Msg("unable to update oauth grant last used time")
	}

	return caller, marker, nil
}

// isOAuthAccessDenied reports whether err came from enforceOAuthAccess
func isOAuthAccessDenied(err error) bool {
	return errors.Is(err, ErrOAuthObjectLimited) || errors.Is(err, rule.ErrOAuthAppReadOnly)
}

// enforceOAuthAccess rejects requests outside an application's access before they reach a handler
// read-only apps may only read, and apps limited to specific objects may only use GraphQL, where every field is checked
func enforceOAuthAccess(c echo.Context, g oauthserver.CallerGrant) error {
	method := c.Request().Method
	isRead := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
	isGraphQL := graphapi.IsGraphQLRequest(c)

	if g.ReadLimited() && !isGraphQL && method != http.MethodOptions {
		return ErrOAuthObjectLimited
	}

	if g.Access != enums.OAuthAccessLevelWrite {
		if isRead || graphapi.CheckGraphReadRequest(c) {
			return nil
		}

		return rule.ErrOAuthAppReadOnly
	}

	if g.WriteLimited() && !isRead && !isGraphQL {
		return ErrOAuthObjectLimited
	}

	return nil
}
