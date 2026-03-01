package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	api "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
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
	caller, ok := auth.CallerFromContext(c.Request().Context())
	return ok && caller != nil && caller.IsImpersonated()
}

// AuthenticateSkipperFuncForWebsockets determines whether Authenticate middleware should be skipped for websocket upgrades
var AuthenticateSkipperFuncForWebsockets = func(c echo.Context) bool {
	return websocket.IsWebSocketUpgrade(c.Request())
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
						return unauthorized(c, err, conf, validator)
					}
				default:
					return unauthorized(c, err, conf, validator)
				}
			}

			var (
				caller *auth.Caller
				id     string
			)

			reqCtx := logx.SeedContext(c.Request().Context())
			c.SetRequest(c.Request().WithContext(reqCtx))
			requestLogger := logx.FromContext(reqCtx)

			switch getTokenType(bearerToken) {
			case auth.PATAuthentication, auth.APITokenAuthentication:
				caller, id, err = checkToken(reqCtx, conf, bearerToken, auth.GetOrganizationContextHeader(c))
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				// Record authentication metric based on token type
				switch caller.AuthenticationType {
				case auth.PATAuthentication:
					metrics.RecordAuthentication(metrics.AuthTypePAT)
				case auth.APITokenAuthentication:
					metrics.RecordAuthentication(metrics.AuthTypeAPIToken)
				}

			default:
				claims, err := validator.VerifyWithContext(reqCtx, bearerToken)
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				if strings.HasPrefix(claims.UserID, "anon_") {
					if !conf.AllowAnonymous {
						return unauthorized(c, ErrAnonymousAccessNotAllowed, conf, validator)
					}

					switch strings.HasPrefix(claims.UserID, "anon_questionnaire") {
					case true:
						ctx := auth.WithCaller(c.Request().Context(), auth.NewQuestionnaireCaller(claims.OrgID, claims.UserID, "Anonymous User", claims.Email))
						ctx = auth.ActiveAssessmentIDKey.Set(ctx, claims.AssessmentID)
						c.SetRequest(c.Request().WithContext(ctx))
					default:
						ctx := auth.WithCaller(c.Request().Context(), auth.NewTrustCenterCaller(claims.OrgID, claims.UserID, "Anonymous User", claims.Email))
						ctx = auth.ActiveTrustCenterIDKey.Set(ctx, claims.TrustCenterID)
						c.SetRequest(c.Request().WithContext(ctx))
					}

					// Record anonymous JWT authentication
					metrics.RecordAuthentication(metrics.AuthTypeJWTAnonymous)

					return next(c)
				}

				// Add claims to context for use in downstream processing and continue handlers
				caller, err = createCallerFromClaims(reqCtx, conf.DBClient, claims, auth.JWTAuthentication)
				if err != nil {
					return unauthorized(c, err, conf, validator)
				}

				ctx := auth.WithRefreshToken(c.Request().Context(), bearerToken)
				c.SetRequest(c.Request().WithContext(ctx))

				// Record regular JWT authentication
				metrics.RecordAuthentication(metrics.AuthTypeJWT)
			}

			c.SetRequest(c.Request().WithContext(auth.WithCaller(c.Request().Context(), caller)))

			if conf.RedisClient != nil {
				permissioncache.SetCacheContext(c, conf.RedisClient)
			}

			// add the user and org ID to the logger context
			requestLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("user_id", caller.SubjectID).
					Strs("org_id", caller.OrganizationIDs)
			})

			if err := updateLastUsedFunc(c.Request().Context(), conf.DBClient, caller, id); err != nil {
				return unauthorized(c, err, conf, validator)
			}

			return next(c)
		}
	}
}

// AuthenticateTransport authenticates a websocket transport init payload and returns the authenticated caller
func AuthenticateTransport(ctx context.Context, initPayload transport.InitPayload, authOptions *Options) (*auth.Caller, error) {
	bearerToken, err := auth.GetBearerTokenFromWebsocketRequest(initPayload)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get bearer token from websocket init payload")

		return nil, ErrUnableToAuthenticateTransport
	}

	validator, err := authOptions.Validator()
	if err != nil {
		return nil, err
	}

	claims, err := validator.VerifyWithContext(ctx, bearerToken)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to verify token in websocket init payload")

		return nil, ErrUnableToAuthenticateTransport
	}

	return createCallerFromClaims(ctx, authOptions.DBClient, claims, auth.JWTAuthentication)
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

func withOrgFilterBypass(ctx context.Context) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	return auth.WithCaller(ctx, caller.WithCapabilities(auth.CapBypassOrgFilter))
}

// updateLastUsed updates the last used time for the token depending on the authentication type
func updateLastUsed(ctx context.Context, dbClient *ent.Client, caller *auth.Caller, tokenID string) error {
	logger := logx.FromContext(ctx)
	switch caller.AuthenticationType {
	case auth.PATAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := withOrgFilterBypass(privacy.DecisionContext(ctx, privacy.Allow))
		if err := dbClient.PersonalAccessToken.UpdateOneID(tokenID).SetLastUsedAt(time.Now()).Exec(allowCtx); err != nil {
			logger.Error().Err(err).Msg("unable to update last used time for personal access token")

			return err
		}
	case auth.APITokenAuthentication:
		// allow the request, we know the user has access to the token, no need to check
		allowCtx := withOrgFilterBypass(privacy.DecisionContext(ctx, privacy.Allow))
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

// createCallerFromClaims creates a Caller directly from the JWT claims provided
func createCallerFromClaims(ctx context.Context, dbClient *ent.Client, claims *tokens.Claims, authType auth.AuthenticationType) (*auth.Caller, error) {
	// get the user ID from the claims
	user, err := dbClient.User.Get(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	systemAdmin, err := isSystemAdminFunc(ctx, dbClient, user.ID, auth.UserSubjectType)
	if err != nil {
		return nil, err
	}

	caller := &auth.Caller{
		SubjectID:          user.ID,
		SubjectName:        getSubjectName(user),
		SubjectEmail:       user.Email,
		OrganizationID:     claims.OrgID,
		OrganizationIDs:    []string{claims.OrgID},
		AuthenticationType: authType,
		Capabilities:       capabilitiesFor(systemAdmin),
	}

	role := getOrgRoleFunc(ctx, dbClient, user.ID, claims.OrgID)
	if role != nil {
		caller.OrganizationRole = *role
	}

	return caller, nil
}

// checkToken checks the bearer authorization token against the database to see if the provided
// token is an active personal access token or api token.
// If the token is valid, the caller is returned
func checkToken(ctx context.Context, conf *Options, token, orgFromHeader string) (*auth.Caller, string, error) {
	// allow check to bypass privacy rules
	ctx = withOrgFilterBypass(privacy.DecisionContext(ctx, privacy.Allow))

	// check if the token is a personal access token
	caller, id, err := isValidPersonalAccessToken(ctx, conf.DBClient, token, orgFromHeader)
	if err == nil {
		return caller, id, nil
	}

	// check if the token is an API token
	caller, id, err = isValidAPIToken(ctx, conf.DBClient, token)
	if err == nil {
		return caller, id, nil
	}

	return nil, "", err
}

// isValidPersonalAccessToken checks if the provided token is a valid personal access token and returns the caller
func isValidPersonalAccessToken(ctx context.Context, dbClient *ent.Client,
	token, orgFromHeader string) (*auth.Caller, string, error) {
	pat, err := fetchPATFunc(ctx, dbClient, token)
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

	caller := &auth.Caller{
		SubjectID:          pat.OwnerID,
		SubjectName:        getSubjectName(pat.Edges.Owner),
		SubjectEmail:       pat.Edges.Owner.Email,
		OrganizationIDs:    orgIDs,
		AuthenticationType: auth.PATAuthentication,
		Capabilities:       capabilitiesFor(systemAdmin),
	}

	var role *auth.OrganizationRoleType
	if len(orgIDs) == 1 {
		role = getOrgRoleFunc(ctx, dbClient, pat.OwnerID, orgIDs[0])
	}

	if role != nil {
		caller.OrganizationRole = *role
	}

	if slices.Contains(orgIDs, orgFromHeader) {
		caller.OrganizationID = orgFromHeader
	}

	return caller, pat.ID, nil
}

// isValidAPIToken checks if the provided token is a valid API token and returns the caller
func isValidAPIToken(ctx context.Context, dbClient *ent.Client, token string) (*auth.Caller, string, error) {
	t, err := fetchAPITokenFunc(ctx, dbClient, token)
	if err != nil {
		return nil, "", err
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

	return &auth.Caller{
		SubjectID:          t.ID,
		SubjectName:        fmt.Sprintf("service: %s", t.Name),
		SubjectEmail:       "", // no email for service accounts
		OrganizationID:     t.OwnerID,
		OrganizationIDs:    []string{t.OwnerID},
		AuthenticationType: auth.APITokenAuthentication,
		Capabilities:       capabilitiesFor(systemAdmin),
	}, t.ID, nil
}

func capabilitiesFor(isAdmin bool) auth.Capability {
	if isAdmin {
		return auth.CapSystemAdmin
	}

	return 0
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

	ctx = models.SSOAuthorizationsContextKey.Set(ctx, &pat.SSOAuthorizations)

	auths, ok := models.SSOAuthorizationsContextKey.Get(ctx)
	if !ok || auths == nil {
		return false, nil
	}

	_, ok = (*auths)[orgID]

	return ok, nil
}

// getOrganizationRole retrieves the organization role for the user in the given organization ID and returns it; if the
// role cannot be determined, nil is returned
func getOrganizationRole(ctx context.Context, db *ent.Client, userID, orgID string) *auth.OrganizationRoleType {
	if orgID == "" || userID == "" {
		logx.FromContext(ctx).Debug().Msg("organization ID or user ID is empty, cannot determine organization role")

		return nil
	}

	role, err := getRole(ctx, db, userID, orgID)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("failed to get organization role, proceeding with other auth checks")
		return nil
	}

	orgRole, ok := auth.ToOrganizationRoleType(role.String())
	if !ok {
		logx.FromContext(ctx).Debug().Str("role", role.String()).Msg("invalid organization role, proceeding with other auth checks")
	}

	return &orgRole
}

func getRole(ctx context.Context, db *ent.Client, userID, orgID string) (enums.Role, error) {
	member, err := db.OrgMembership.Query().
		Where(orgmembership.UserID(userID), orgmembership.OrganizationID(orgID)).
		Select(orgmembership.FieldRole).
		Only(privacy.DecisionContext(ctx, privacy.Allow))
	if err != nil {
		return "", err
	}

	return member.Role, nil
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
		return db.PersonalAccessToken.Query().Where(personalaccesstoken.Token(token)).
			WithOwner().
			WithOrganizations().
			Only(ctx)
	}

	fetchAPITokenFunc = func(ctx context.Context, db *ent.Client, token string) (*ent.APIToken, error) {
		return db.APIToken.Query().Where(apitoken.Token(token)).Only(ctx)
	}

	isPATSSOAuthorizedFunc = isPATSSOAuthorized

	updateLastUsedFunc = updateLastUsed

	isSystemAdminFunc = isSystemAdmin

	getOrgRoleFunc = getOrganizationRole
)
