package authmanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/models"
)

const (
	bearerScheme = "Bearer"
)

// Client holds the necessary clients and configuration for the auth manager
type Client struct {
	db *generated.Client
}

// New creates a new auth manager client with the provided database client
func New(db *generated.Client) *Client {
	return &Client{
		db: db,
	}
}

// SetDBClient sets the database client in the config
func (a *Client) SetDBClient(db *generated.Client) {
	a.db = db
}

// GetDBClient returns the database client
func (a *Client) GetDBClient() *generated.Client {
	return a.db
}

// GetSessionConfig returns the session config
func (a *Client) GetSessionConfig() *sessions.SessionConfig {
	return a.db.SessionConfig
}

// GenerateUserAuthSessionWithOrg creates a new auth session for the user and the new target organization id
// this is used when the user is switching organizations, or when the user deletes their authorized organization
// and is automatically switched into another organization
// Before the sessions is issues, we check that the user still has access to the target organization
// if not, the user's default org (or personal org) is used
func (a *Client) GenerateUserAuthSessionWithOrg(ctx context.Context, w http.ResponseWriter, user *generated.User, targetOrgID string) (*models.AuthData, error) {
	auth, err := a.createTokenPair(ctx, user, targetOrgID)
	if err != nil {
		return nil, err
	}

	auth.Session, err = a.generateUserSession(ctx, w, user.ID)
	if err != nil {
		return nil, err
	}

	auth.TokenType = bearerScheme

	return auth, nil
}

// GenerateUserAuthSession creates a new auth session for the user and their default organization id
// this is used during the login process
func (a *Client) GenerateUserAuthSession(ctx context.Context, w http.ResponseWriter, user *generated.User) (*models.AuthData, error) {
	return a.GenerateUserAuthSessionWithOrg(ctx, w, user, "")
}

// GenerateOauthAuthSession creates a new auth session for the oauth user and their default organization id
func (a *Client) GenerateOauthAuthSession(ctx context.Context, w http.ResponseWriter, user *generated.User, oauthRequest models.OauthTokenRequest) (*models.AuthData, error) {
	auth, err := a.createTokenPair(ctx, user, "")
	if err != nil {
		return nil, err
	}

	auth.Session, err = a.generateOauthUserSession(ctx, w, user.ID, oauthRequest)
	if err != nil {
		return nil, err
	}

	auth.TokenType = bearerScheme

	return auth, nil
}

var (
	// ErrOrgSubscriptionNotActive is the error message when the organization subscription is not active
	ErrOrgSubscriptionNotActive = errors.New("organization subscription is not active")
)

func (a *Client) checkActiveSubscription(ctx context.Context, orgID string) (active bool, err error) {
	// if the entitlement manager is disabled, we can skip the check
	if a.db.EntitlementManager == nil {
		return true, nil
	}

	if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
		return true, nil
	}

	// allow to skip the org interceptor middleware before a user could potentially be authenticated
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	subscription, err := a.db.OrgSubscription.Query().Select("active").Where(orgsubscription.OwnerID(orgID)).Only(allowCtx)
	if err != nil {
		return false, err
	}

	if subscription == nil || !subscription.Active {
		return false, ErrOrgSubscriptionNotActive
	}

	return true, nil
}

// createClaims creates the claims for the JWT token using the id for the user and organization
// if not target org is provided, the user's default org is used
func createClaimsWithOrg(u *generated.User, targetOrgID string) *tokens.Claims {
	if targetOrgID == "" {
		if u.Edges.Setting.Edges.DefaultOrg != nil {
			targetOrgID = u.Edges.Setting.Edges.DefaultOrg.ID
		}
	}

	return &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: u.ID,
		},
		UserID: u.ID,
		OrgID:  targetOrgID,
	}
}

// createTokenPair creates a new token pair for the user and the target organization id (or default org if none provided)
func (a *Client) createTokenPair(ctx context.Context, user *generated.User, targetOrgID string) (*models.AuthData, error) {
	newTarget, err := a.authCheck(ctx, user, targetOrgID)
	if err != nil {
		if targetOrgID != "" {
			log.Error().Err(err).Msg("user attempting to switch into an org they cannot access")

			return nil, err
		}

		log.Warn().Err(err).Msg("user attempting to authenticate into an org they cannot access; switching to personal org")

		targetOrgID = newTarget
	}

	// create new claims for the user
	newClaims := createClaimsWithOrg(user, targetOrgID)

	// create a new token pair for the user
	access, refresh, err := a.db.TokenManager.CreateTokenPair(newClaims)
	if err != nil {
		return nil, err
	}

	return &models.AuthData{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// GenerateUserSession creates a new session for the user and stores it in the response
func (a *Client) generateUserSession(ctx context.Context, w http.ResponseWriter, userID string) (string, error) {
	var err error

	ctx, err = a.db.SessionConfig.CreateAndStoreSession(ctx, w, userID)
	if err != nil {
		return "", err
	}

	// return the session value for the UI to use
	session, err := sessions.SessionToken(ctx)
	if err != nil {
		return "", err
	}

	return session, nil
}

// generateOauthUserSession creates a new session for the oauth user and stores it in the response
func (a *Client) generateOauthUserSession(ctx context.Context, w http.ResponseWriter, userID string, oauthRequest models.OauthTokenRequest) (string, error) {
	setSessionMap := map[string]any{}
	setSessionMap[sessions.ExternalUserIDKey] = fmt.Sprintf("%v", oauthRequest.ExternalUserID)
	setSessionMap[sessions.UsernameKey] = oauthRequest.ExternalUserName
	setSessionMap[sessions.UserTypeKey] = oauthRequest.AuthProvider
	setSessionMap[sessions.EmailKey] = oauthRequest.Email
	setSessionMap[sessions.UserIDKey] = userID

	c, err := a.db.SessionConfig.SaveAndStoreSession(ctx, w, setSessionMap, userID)
	if err != nil {
		return "", err
	}

	session, err := sessions.SessionToken(c)
	if err != nil {
		return "", err
	}

	return session, nil
}

// authCheck checks if the user has access to the target organization before issuing a new session and claims
// if the user does not have access to the target organization, the user's default org is used (or falls back)
// to their personal org
func (a *Client) authCheck(ctx context.Context, user *generated.User, orgID string) (string, error) {
	if skip := skipOrgValidation(ctx); skip {
		return orgID, nil
	}

	if orgID == "" {
		// get the default org for the user to check access
		var err error

		orgID, err = getUserDefaultOrg(ctx, user)
		if err != nil {
			return "", err
		}
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get authenticated user context")

		return "", err
	}

	// if no org is provided, check with the authenticated org
	if orgID == "" {
		orgID = au.OrganizationID
	}

	active, err := a.checkActiveSubscription(ctx, orgID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find org subscription for organization")

		return "", err
	}

	// ensure user is already a member of the destination organization
	req := fgax.AccessCheck{
		SubjectID:   au.SubjectID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    orgID,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	allow, err := a.db.Authz.CheckOrgReadAccess(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("unable to check access")

		return "", err
	}

	// if the org is active and they are allowed, we can return the org
	if active && allow {
		return orgID, nil
	}

	// if the user was not allowed or the org is not active, we need to update the default org if we used their default org
	// to authenticate
	newOrgID, err := a.updateDefaultOrg(ctx, user, orgID)
	if err != nil {
		log.Error().Err(err).Msg("unable to update default org")

		return "", err
	}

	return newOrgID, generated.ErrPermissionDenied
}

// getUserDefaultOrg returns the default org for the user
func getUserDefaultOrg(ctx context.Context, user *generated.User) (string, error) {
	// check if we already have the default org
	if user.Edges.Setting.Edges.DefaultOrg != nil {
		return user.Edges.Setting.Edges.DefaultOrg.ID, nil
	}

	// otherwise, query the default org from the user setting and allow the request
	// incase the user is in the login process
	orgCtx := privacy.DecisionContext(ctx, privacy.Allow)

	org, err := user.Edges.Setting.DefaultOrg(orgCtx)
	if err != nil {
		log.Error().Err(err).Msg("error obtaining default org")

		return "", err
	}

	// add default org to user object to use later
	user.Edges.Setting.Edges.DefaultOrg = org

	return org.ID, nil
}

func (a *Client) updateDefaultOrg(ctx context.Context, user *generated.User, orgID string) (string, error) {
	defaultOrgID, err := getUserDefaultOrg(ctx, user)
	if err != nil {
		return "", err
	}

	// if the org we were checking was not the default org, we can return that org instead
	// of updating the default org
	if orgID != defaultOrgID {
		return defaultOrgID, nil
	}

	// update default org to personal org
	newOrgID, err := a.updateDefaultOrgToPersonal(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("error updating default org to personal org")

		return "", err
	}

	return newOrgID, nil
}

// updateDefaultOrgToPersonal updates the default org for the user to the personal org
func (a *Client) updateDefaultOrgToPersonal(ctx context.Context, user *generated.User) (string, error) {
	log.Debug().Str("user", user.ID).Msg("user no longer has access to their default org, switching the default to their personal org")

	personalOrg, err := a.getPersonalOrgID(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("error obtaining personal org")

		return "", err
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	err = a.db.UserSetting.Update().
		SetDefaultOrgID(personalOrg.ID).
		Where(usersetting.UserID(user.ID)).
		Exec(allowCtx)
	if err != nil {
		log.Error().Err(err).Msg("error updating default org")

		return "", err
	}

	// update the user object with the new default org
	if user.Edges.Setting == nil {
		user.Edges.Setting, err = a.db.UserSetting.Query().
			Where(usersetting.UserID(user.ID)).Only(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error obtaining user setting")

			return "", err
		}
	}

	user.Edges.Setting.Edges.DefaultOrg = personalOrg

	return personalOrg.ID, nil
}

// getPersonalOrgID returns the personal org ID for the user
func (a *Client) getPersonalOrgID(ctx context.Context, user *generated.User) (*generated.Organization, error) {
	return a.db.User.QueryOrganizations(user).Where(organization.PersonalOrg(true)).Only(ctx)
}

// skipOrgValidation checks if the org validation should be skipped based on the context
func skipOrgValidation(ctx context.Context) bool {
	// skip if explicitly allowed
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return true
	}

	// skip on org creation
	if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
		return true
	}

	// skip on privacy tokens
	skipTokenType := []token.PrivacyToken{
		&token.VerifyToken{},
		&token.SignUpToken{},
	}

	return rule.SkipTokenInContext(ctx, skipTokenType)
}
