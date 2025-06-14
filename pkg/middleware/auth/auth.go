package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	api "github.com/theopenlane/core/pkg/models"
)

// SessionSkipperFunc is the function that determines if the session check should be skipped
// due to the request being a PAT or API Token auth request
var SessionSkipperFunc = func(c echo.Context) bool {
	return auth.GetAuthTypeFromEchoContext(c) != auth.JWTAuthentication
}

// Authenticate is a middleware function that is used to authenticate requests - it is not applied to all routes so be cognizant of that
func Authenticate(conf *Options) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// if skipper function returns true, skip this middleware
			if conf.Skipper(c) {
				return next(c)
			}

			// execute any before functions
			if conf.BeforeFunc != nil {
				conf.BeforeFunc(c)
			}

			validator, err := conf.Validator()
			if err != nil {
				return unauthorized(c, err)
			}

			// Create a reauthenticator function to handle refresh tokens if they are provided.
			reauthenticate := Reauthenticate(conf, validator)

			// Get access token from the request, if not available then attempt to refresh
			// using the refresh token cookie.
			bearerToken, err := auth.GetBearerToken(c)
			if err != nil {
				switch {
				case errors.Is(err, ErrNoAuthorization):
					if bearerToken, err = reauthenticate(c); err != nil {
						return unauthorized(c, err)
					}
				default:
					return unauthorized(c, err)
				}
			}

			var (
				au *auth.AuthenticatedUser
				id string
			)

			reqCtx := c.Request().Context()

			switch getTokenType(bearerToken) {
			case auth.PATAuthentication, auth.APITokenAuthentication:
				au, id, err = checkToken(reqCtx, conf, bearerToken)
				if err != nil {
					return unauthorized(c, err)
				}

			default:
				claims, err := validator.Verify(bearerToken)
				if err != nil {
					return unauthorized(c, err)
				}

				if strings.HasPrefix(claims.UserID, "anon:") {
					if !conf.AllowAnonymous {
						return unauthorized(c, fmt.Errorf("anonymous access not allowed"))
					}
					an, err := createAnonymousTrustCenterUserFromClaims(reqCtx, conf.DBClient, claims, auth.JWTAuthentication)
					if err != nil {
						return unauthorized(c, err)
					}

					auth.SetAnonymousTrustCenterUserContext(c, an)

					return next(c)
				} else {
					// Add claims to context for use in downstream processing and continue handlers
					au, err = createAuthenticatedUserFromClaims(reqCtx, conf.DBClient, claims, auth.JWTAuthentication)
					if err != nil {
						return unauthorized(c, err)
					}
				}

				auth.SetRefreshToken(c, bearerToken)
			}

			auth.SetAuthenticatedUserContext(c, au)

			// add the user and org ID to the logger context
			zerolog.Ctx(reqCtx).UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("user_id", au.SubjectID).
					Str("org_id", au.OrganizationID)
			})

			if err := updateLastUsed(c.Request().Context(), conf.DBClient, au, id); err != nil {
				return unauthorized(c, err)
			}

			return next(c)
		}
	}
}

// getTokenType returns the authentication type based on the bearer token
// this is used to determine if the token is a personal access token or an API token
// the default is JWT authentication
func getTokenType(bearerToken string) auth.AuthenticationType {
	switch {
	case strings.HasPrefix(bearerToken, "tola_"):
		return auth.APITokenAuthentication
	case strings.HasPrefix(bearerToken, "tolp_"):
		return auth.PATAuthentication
	}

	return auth.JWTAuthentication
}

// updateLastUsed updates the last used time for the token depending on the authentication type
func updateLastUsed(ctx context.Context, dbClient *ent.Client, au *auth.AuthenticatedUser, tokenID string) error {
	switch au.AuthenticationType {
	case auth.PATAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
		if err := dbClient.PersonalAccessToken.UpdateOneID(tokenID).SetLastUsedAt(time.Now()).Exec(allowCtx); err != nil {
			log.Error().Err(err).Msg("unable to update last used time for personal access token")

			return err
		}
	case auth.APITokenAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
		if err := dbClient.APIToken.UpdateOneID(tokenID).SetLastUsedAt(time.Now()).Exec(allowCtx); err != nil {
			log.Err(err).Msg("unable to update last used time for API token")

			return err
		}
	}

	return nil
}

// Reauthenticate is a middleware helper that can use refresh tokens in the echo context
// to obtain a new access token. If it is unable to obtain a new valid access token,
// then an error is returned and processing should stop.
func Reauthenticate(conf *Options, validator tokens.Validator) func(c echo.Context) (string, error) {
	// If no reauthenticator is available on the configuration, always return an error.
	if conf.reauth == nil {
		return func(_ echo.Context) (string, error) {
			return "", ErrRefreshDisabled
		}
	}

	// If the reauthenticator is available, return a function that utilizes it.
	return func(c echo.Context) (string, error) {
		// Get the refresh token from the cookies or the headers of the request.
		refreshToken, err := auth.GetRefreshToken(c)
		if err != nil {
			return "", err
		}

		// Check to ensure the refresh token is still valid.
		if _, err = validator.Verify(refreshToken); err != nil {
			return "", err
		}

		// Reauthenticate using the refresh token.
		req := &api.RefreshRequest{RefreshToken: refreshToken}

		reply, err := conf.reauth.Refresh(c.Request().Context(), req)
		if err != nil {
			return "", err
		}

		// Set the new access and refresh cookies
		auth.SetAuthCookies(c.Response().Writer, reply.AccessToken, reply.RefreshToken, *conf.CookieConfig)

		return reply.AccessToken, nil
	}
}

// createAuthenticatedUserFromClaims creates an authenticated user from the claims provided
func createAuthenticatedUserFromClaims(ctx context.Context, dbClient *ent.Client, claims *tokens.Claims, authType auth.AuthenticationType) (*auth.AuthenticatedUser, error) {
	// get the user ID from the claims
	user, err := dbClient.User.Get(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// all the query to get the organization, need to bypass the authz filter to get the org
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	org, err := dbClient.Organization.Get(ctx, claims.OrgID)
	if err != nil {
		return nil, err
	}

	return &auth.AuthenticatedUser{
		SubjectID:          user.ID,
		SubjectName:        getSubjectName(user),
		SubjectEmail:       user.Email,
		OrganizationID:     org.ID,
		OrganizationName:   org.Name,
		OrganizationIDs:    []string{org.ID},
		AuthenticationType: authType,
	}, nil
}

func createAnonymousTrustCenterUserFromClaims(ctx context.Context, dbClient *ent.Client, claims *tokens.Claims, authType auth.AuthenticationType) (*auth.AnonymousTrustCenterUser, error) {

	return &auth.AnonymousTrustCenterUser{
		SubjectID:          claims.UserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     claims.OrgID,
		AuthenticationType: authType,
		TrustCenterID:      claims.TrustCenterID,
	}, nil
}

// checkToken checks the bearer authorization token against the database to see if the provided
// token is an active personal access token or api token.
// If the token is valid, the authenticated user is returned
func checkToken(ctx context.Context, conf *Options, token string) (*auth.AuthenticatedUser, string, error) {
	// allow check to bypass privacy rules
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// check if the token is a personal access token
	au, id, err := isValidPersonalAccessToken(ctx, *conf.DBClient, token)
	if err == nil {
		return au, id, nil
	}

	// check if the token is an API token
	au, id, err = isValidAPIToken(ctx, *conf.DBClient, token)
	if err == nil {
		return au, id, nil
	}

	return nil, "", err
}

// isValidPersonalAccessToken checks if the provided token is a valid personal access token and returns the authenticated user
func isValidPersonalAccessToken(ctx context.Context, dbClient ent.Client, token string) (*auth.AuthenticatedUser, string, error) {
	// query for the token before the user is authorized
	// allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	pat, err := dbClient.PersonalAccessToken.Query().Where(personalaccesstoken.Token(token)).
		WithOwner().
		WithOrganizations().
		Only(ctx)
	if err != nil {
		return nil, "", err
	}

	// check if the token has expired
	if pat.ExpiresAt != nil && pat.ExpiresAt.Before(time.Now()) {
		return nil, "", rout.ErrExpiredCredentials
	}

	// gather the authorized organization IDs
	var orgIDs []string

	orgs := pat.Edges.Organizations

	if len(orgs) == 0 {
		// an access token must have at least one organization to be used
		return nil, "", rout.ErrInvalidCredentials
	}

	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}

	return &auth.AuthenticatedUser{
		SubjectID:          pat.OwnerID,
		SubjectName:        getSubjectName(pat.Edges.Owner),
		SubjectEmail:       pat.Edges.Owner.Email,
		OrganizationIDs:    orgIDs,
		AuthenticationType: auth.PATAuthentication,
	}, pat.ID, nil
}

func isValidAPIToken(ctx context.Context, dbClient ent.Client, token string) (*auth.AuthenticatedUser, string, error) {
	// query for the token before the user is authorized
	// allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	t, err := dbClient.APIToken.Query().Where(apitoken.Token(token)).
		Only(ctx)
	if err != nil {
		return nil, "", err
	}

	// check if the token has expired
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		return nil, "", rout.ErrExpiredCredentials
	}

	return &auth.AuthenticatedUser{
		SubjectID:          t.ID,
		SubjectName:        fmt.Sprintf("service: %s", t.Name),
		SubjectEmail:       "", // no email for service accounts
		OrganizationID:     t.OwnerID,
		OrganizationIDs:    []string{t.OwnerID},
		AuthenticationType: auth.APITokenAuthentication,
	}, t.ID, nil
}

// getSubjectName returns the subject name for the user
func getSubjectName(user *ent.User) string {
	if user == nil {
		return ""
	}

	subjectName := user.FirstName + " " + user.LastName
	if subjectName == "" {
		subjectName = user.DisplayName
	}

	return subjectName
}

// unauthorized returns a 401 Unauthorized response with the error message.
func unauthorized(c echo.Context, err error) error {
	if err := c.JSON(http.StatusUnauthorized, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}
