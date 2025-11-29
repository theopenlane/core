package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/models"
	api "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/permissioncache"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// These are stored here, instead of iam because they are specific
// to the openlane core service setup, other services could define different object
// type and ids in particular
const (
	// SystemObject type for FGA authorization
	SystemObject = "system"
	// SystemObjectID for FGA authorization
	SystemObjectID = "openlane_core"
)

// SessionSkipperFunc is the function that determines if the session check should be skipped
// due to the request being a PAT or API Token auth request
var SessionSkipperFunc = func(c echo.Context) bool {
	return auth.GetAuthTypeFromEchoContext(c) != auth.JWTAuthentication
}

// AuthenticateSkipperFuncForImpersonation determines whether Authenticate middleware should be skipped
// based on the presence of impersonation token
var AuthenticateSkipperFuncForImpersonation = func(c echo.Context) bool {
	_, ok := auth.ImpersonatedUserFromContext(c.Request().Context())
	return ok
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
				return unauthorized(c, err, conf, validator)
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
						if errors.Is(err, ErrRefreshDisabled) {
							return unauthorized(c, ErrNoAuthorization, conf, validator)
						}

						return unauthorized(c, err, conf, validator)
					}
				default:
					return unauthorized(c, err, conf, validator)
				}
			}

			var (
				au *auth.AuthenticatedUser
				id string
			)

			reqCtx := logx.SeedContext(c.Request().Context())
			c.SetRequest(c.Request().WithContext(reqCtx))
			requestLogger := logx.FromContext(reqCtx)

			switch getTokenType(bearerToken) {
			case auth.PATAuthentication, auth.APITokenAuthentication:
				au, id, err = checkToken(reqCtx, conf, bearerToken, auth.GetOrganizationContextHeader(c))
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				// Record authentication metric based on token type
				switch au.AuthenticationType {
				case auth.PATAuthentication:
					metrics.RecordAuthentication(metrics.AuthTypePAT)
				case auth.APITokenAuthentication:
					metrics.RecordAuthentication(metrics.AuthTypeAPIToken)
				}

			default:
				claims, err := validator.Verify(bearerToken)
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				if strings.HasPrefix(claims.UserID, "anon_") {
					if !conf.AllowAnonymous {
						return unauthorized(c, ErrAnonymousAccessNotAllowed, conf, validator)
					}

					switch strings.HasPrefix(claims.UserID, "anon_questionnaire") {

					case true:
						an, err := createAnonymousQuestionnaireFromClaims(claims, auth.JWTAuthentication)
						if err != nil {
							return unauthorized(c, err, conf, validator)
						}

						auth.SetAnonymousQuestionnaireUserContext(c, an)

					default:

						an, err := createAnonymousTrustCenterUserFromClaims(claims, auth.JWTAuthentication)
						if err != nil {
							return unauthorized(c, err, conf, validator)
						}

						auth.SetAnonymousTrustCenterUserContext(c, an)
					}

					// Record anonymous JWT authentication
					metrics.RecordAuthentication(metrics.AuthTypeJWTAnonymous)

					return next(c)
				}

				// Add claims to context for use in downstream processing and continue handlers
				au, err = createAuthenticatedUserFromClaims(reqCtx, conf.DBClient, claims, auth.JWTAuthentication)
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				auth.SetRefreshToken(c, bearerToken)

				// Record regular JWT authentication
				metrics.RecordAuthentication(metrics.AuthTypeJWT)
			}

			auth.SetAuthenticatedUserContext(c, au)

			if conf.RedisClient != nil {
				permissioncache.SetCacheContext(c, conf.RedisClient)
			}

			// add the user and org ID to the logger context
			requestLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("user_id", au.SubjectID).
					Strs("org_id", au.OrganizationIDs)
			})

			if err := updateLastUsedFunc(c.Request().Context(), conf.DBClient, au, id); err != nil {
				return unauthorized(c, err, conf, validator)
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
	logger := logx.FromContext(ctx)
	switch au.AuthenticationType {
	case auth.PATAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
		if err := dbClient.PersonalAccessToken.UpdateOneID(tokenID).SetLastUsedAt(time.Now()).Exec(allowCtx); err != nil {
			logger.Error().Err(err).Msg("unable to update last used time for personal access token")

			return err
		}
	case auth.APITokenAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
		if err := dbClient.APIToken.UpdateOneID(tokenID).SetLastUsedAt(time.Now()).Exec(allowCtx); err != nil {
			logger.Error().Err(err).Msg("unable to update last used time for API token")

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

	systemAdmin, err := isSystemAdminFunc(ctx, dbClient, user.ID, auth.UserSubjectType)
	if err != nil {
		return nil, err
	}

	return &auth.AuthenticatedUser{
		SubjectID:          user.ID,
		SubjectName:        getSubjectName(user),
		SubjectEmail:       user.Email,
		OrganizationID:     claims.OrgID,
		OrganizationIDs:    []string{claims.OrgID},
		AuthenticationType: authType,
		IsSystemAdmin:      systemAdmin,
	}, nil
}

func createAnonymousQuestionnaireFromClaims(claims *tokens.Claims, authType auth.AuthenticationType) (*auth.AnonymousQuestionnaireUser, error) {
	return &auth.AnonymousQuestionnaireUser{
		SubjectID:          claims.UserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     claims.OrgID,
		AuthenticationType: authType,
		SubjectEmail:       claims.Email,
		AssessmentID:       claims.AssessmentID,
	}, nil
}

func createAnonymousTrustCenterUserFromClaims(claims *tokens.Claims, authType auth.AuthenticationType) (*auth.AnonymousTrustCenterUser, error) {
	return &auth.AnonymousTrustCenterUser{
		SubjectID:          claims.UserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     claims.OrgID,
		AuthenticationType: authType,
		TrustCenterID:      claims.TrustCenterID,
		SubjectEmail:       claims.Email,
	}, nil
}

// checkToken checks the bearer authorization token against the database to see if the provided
// token is an active personal access token or api token.
// If the token is valid, the authenticated user is returned
func checkToken(ctx context.Context, conf *Options, token, orgFromHeader string) (*auth.AuthenticatedUser, string, error) {
	// allow check to bypass privacy rules
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// check if the token is a personal access token
	au, id, err := isValidPersonalAccessToken(ctx, conf.DBClient, token, orgFromHeader)
	if err == nil {
		return au, id, nil
	}

	// check if the token is an API token
	au, id, err = isValidAPIToken(ctx, conf.DBClient, token)
	if err == nil {
		return au, id, nil
	}

	return nil, "", err
}

// isValidPersonalAccessToken checks if the provided token is a valid personal access token and returns the authenticated user
func isValidPersonalAccessToken(ctx context.Context, dbClient *ent.Client,
	token, orgFromHeader string) (*auth.AuthenticatedUser, string, error) {
	// verify the token format and extract the public ID and secret
	publicID, secret, err := parseToken(token)
	if err != nil {
		return nil, "", err
	}

	pat, err := fetchPATFunc(ctx, dbClient, publicID)
	if err != nil {
		return nil, "", err
	}

	// verify the secret
	if pat.TokenSecret != secret {
		return nil, "", rout.ErrInvalidCredentials
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
		enforced, err := isSSOEnforcedFunc(ctx, dbClient, org.ID)
		if err != nil {
			return nil, "", err
		}

		if enforced {
			authorized, err := isPATSSOAuthorizedFunc(ctx, dbClient, pat.ID, org.ID)
			if err != nil {
				return nil, "", err
			}

			if !authorized {
				return nil, "", ErrTokenSSORequired
			}
		}

		orgIDs = append(orgIDs, org.ID)
	}

	systemAdmin, err := isSystemAdminFunc(ctx, dbClient, pat.OwnerID, auth.UserSubjectType)
	if err != nil {
		return nil, "", err
	}

	authUser := &auth.AuthenticatedUser{
		SubjectID:          pat.OwnerID,
		SubjectName:        getSubjectName(pat.Edges.Owner),
		SubjectEmail:       pat.Edges.Owner.Email,
		OrganizationIDs:    orgIDs,
		AuthenticationType: auth.PATAuthentication,
		IsSystemAdmin:      systemAdmin,
	}

	if slices.Contains(orgIDs, orgFromHeader) {
		authUser.OrganizationID = orgFromHeader
	}

	return authUser, pat.ID, nil
}

func isValidAPIToken(ctx context.Context, dbClient *ent.Client, token string) (*auth.AuthenticatedUser, string, error) {
	// verify the token format and extract the public ID and secret
	publicID, secret, err := parseToken(token)
	if err != nil {
		return nil, "", err
	}

	t, err := fetchAPITokenFunc(ctx, dbClient, publicID)
	if err != nil {
		return nil, "", err
	}

	// verify the secret
	if t.TokenSecret != secret {
		return nil, "", rout.ErrInvalidCredentials
	}

	// check if the token has expired
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		return nil, "", rout.ErrExpiredCredentials
	}

	enforced, err := isSSOEnforcedFunc(ctx, dbClient, t.OwnerID)
	if err != nil {
		return nil, "", err
	}

	if enforced {
		if t.SSOAuthorizations == nil || t.SSOAuthorizations[t.OwnerID].IsZero() {
			return nil, "", ErrTokenSSORequired
		}
	}

	systemAdmin, err := isSystemAdminFunc(ctx, dbClient, t.OwnerID, auth.ServiceSubjectType)
	if err != nil {
		return nil, "", err
	}

	return &auth.AuthenticatedUser{
		SubjectID:          t.ID,
		SubjectName:        fmt.Sprintf("service: %s", t.Name),
		SubjectEmail:       "", // no email for service accounts
		OrganizationID:     t.OwnerID,
		OrganizationIDs:    []string{t.OwnerID},
		AuthenticationType: auth.APITokenAuthentication,
		IsSystemAdmin:      systemAdmin,
	}, t.ID, nil
}

// isSystemAdmin checks fga to see if the user is a system admin
// it uses the authz client from the context to perform the check
// it returns true if the user is a system admin, false otherwise
// if the authz client is not available, it returns false
func isSystemAdmin(ctx context.Context, dbClient *ent.Client, subjectID, subjectType string) (bool, error) {
	ac := fgax.AccessCheck{
		ObjectType:  SystemObject,
		ObjectID:    SystemObjectID,
		Relation:    fgax.SystemAdminRelation,
		SubjectID:   subjectID,
		SubjectType: subjectType,
	}

	if dbClient == nil {
		return false, nil
	}

	return dbClient.Authz.CheckAccess(ctx, ac)
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

// unauthorized returns a 401 Unauthorized response with the error message
func unauthorized(c echo.Context, err error, conf *Options, v tokens.Validator) error {
	reqCtx := c.Request().Context()
	logger := logx.FromContext(reqCtx)

	if v != nil && conf != nil && conf.DBClient != nil {
		orgID := orgIDFromToken(c, v)
		if orgID != "" {
			userID := userIDFromToken(c, v)
			if userID != "" {
				enforced, dbErr := isSSOEnforcedFunc(reqCtx, conf.DBClient, orgID)

				if dbErr == nil && enforced {
					role, mErr := orgRoleFunc(reqCtx, conf.DBClient, userID, orgID)

					if mErr == nil && role != enums.RoleOwner {
						if redirErr := c.Redirect(http.StatusFound, sso.SSOLogin(c.Echo(), orgID)); redirErr != nil {
							return redirErr
						}

						return nil
					}
				}
			}
		}
	}

	if jsonErr := c.JSON(http.StatusUnauthorized, rout.ErrorResponse(err)); jsonErr != nil {
		logger.Error().Err(jsonErr).Msg("failed to write unauthorized JSON response")
		return jsonErr
	}

	return nil
}

// orgIDFromToken extracts the organization ID from the token in the request context
func orgIDFromToken(c echo.Context, v tokens.Validator) string {
	authHeader := c.Request().Header.Get("Authorization")
	logger := logx.FromContext(c.Request().Context())

	if authHeader != "" {
		token, err := auth.GetBearerToken(c)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get bearer token from context")
			return ""
		}

		if token != "" {
			if claims, parseErr := v.Parse(token); parseErr == nil {
				return claims.OrgID
			}
		}
	}

	cookie, err := c.Cookie("refresh_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		refresh := cookie.Value

		if claims, parseErr := v.Parse(refresh); parseErr == nil {
			return claims.OrgID
		}
	}

	return ""
}

// isSSOEnforced checks if SSO is enforced for the given organization ID
func isSSOEnforced(ctx context.Context, db *ent.Client, orgID string) (bool, error) {
	setting, err := db.OrganizationSetting.Query().Where(organizationsetting.OrganizationID(orgID)).
		Only(privacy.DecisionContext(ctx, privacy.Allow))
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return setting.IdentityProviderLoginEnforced, nil
}

// userIDFromToken extracts the user ID from the token in the request context
func userIDFromToken(c echo.Context, v tokens.Validator) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		token, _ := auth.GetBearerToken(c)
		if token != "" {
			if claims, parseErr := v.Parse(token); parseErr == nil {
				return claims.UserID
			}
		}
	}

	cookie, err := c.Cookie("refresh_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		refresh := cookie.Value
		if claims, parseErr := v.Parse(refresh); parseErr == nil {
			return claims.UserID
		}
	}

	return ""
}

// isPATSSOAuthorized checks if the personal access token is authorized for SSO with the given organization ID
func isPATSSOAuthorized(ctx context.Context, db *ent.Client, tokenID, orgID string) (bool, error) {
	pat, err := db.PersonalAccessToken.Get(ctx, tokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	ctx = models.WithSSOAuthorizations(ctx, &pat.SSOAuthorizations)

	auths, ok := models.SSOAuthorizationsFromContext(ctx)
	if !ok || auths == nil {
		return false, nil
	}

	_, ok = (*auths)[orgID]

	return ok, nil
}

// function variables allow tests to override SSO checks without a database
var (
	isSSOEnforcedFunc = isSSOEnforced
	orgRoleFunc       = func(ctx context.Context, db *ent.Client, userID, orgID string) (enums.Role, error) {
		member, err := db.OrgMembership.Query().
			Where(orgmembership.UserID(userID), orgmembership.OrganizationID(orgID)).
			Only(privacy.DecisionContext(ctx, privacy.Allow))
		if err != nil {
			return "", err
		}

		return member.Role, nil
	}

	fetchPATFunc = func(ctx context.Context, db *ent.Client, token string) (*ent.PersonalAccessToken, error) {
		return db.PersonalAccessToken.Query().Where(personalaccesstoken.TokenPublicID(token)).
			WithOwner().
			WithOrganizations().
			Only(ctx)
	}

	fetchAPITokenFunc = func(ctx context.Context, db *ent.Client, token string) (*ent.APIToken, error) {
		return db.APIToken.Query().Where(apitoken.TokenPublicID(token)).Only(ctx)
	}

	isPATSSOAuthorizedFunc = isPATSSOAuthorized

	updateLastUsedFunc = updateLastUsed

	isSystemAdminFunc = isSystemAdmin
)

// parseToken parses the token string and returns the public ID and secret
func parseToken(token string) (publicID, secret string, err error) {
	const PartsCount int = 3
	if strings.HasPrefix(token, "tola_") {
		parts := strings.Split(token, "_")
		if len(parts) != PartsCount {
			return "", "", rout.ErrInvalidCredentials
		}
		return parts[1], parts[2], nil
	} else if strings.HasPrefix(token, "tolp_") {
		parts := strings.Split(token, "_")
		if len(parts) != PartsCount {
			return "", "", rout.ErrInvalidCredentials
		}
		return parts[1], parts[2], nil
	}

	return "", "", rout.ErrInvalidCredentials
}
