package authmanager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/models"
)

const (
	bearerAuthType = "Bearer"
)

type Config struct {
	db *generated.Client
}

func New(db *generated.Client) *Config {
	return &Config{
		db: db,
	}
}

func (a *Config) GetSessionConfig() *sessions.SessionConfig {
	return a.db.SessionConfig
}

// generateNewAuthSession creates a new auth session for the user and their default organization id
func (a *Config) GenerateUserAuthSession(ctx echo.Context, user *generated.User) (*models.AuthData, error) {
	return a.GenerateUserAuthSessionWithOrg(ctx, user, "")
}

// generateUserAuthSessionWithOrg creates a new auth session for the user and the new target organization id
func (a *Config) GenerateUserAuthSessionWithOrg(ctx echo.Context, user *generated.User, targetOrgID string) (*models.AuthData, error) {
	auth, err := a.createTokenPair(ctx.Request().Context(), user, targetOrgID)
	if err != nil {
		return nil, err
	}

	auth.Session, err = a.generateUserSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	auth.TokenType = bearerAuthType

	return auth, nil
}

// GenerateOauthAuthSession creates a new auth session for the oauth user and their default organization id
func (a *Config) GenerateOauthAuthSession(ctx context.Context, w http.ResponseWriter, user *generated.User, oauthRequest models.OauthTokenRequest) (*models.AuthData, error) {
	auth, err := a.createTokenPair(ctx, user, "")
	if err != nil {
		return nil, err
	}

	auth.Session, err = a.generateOauthUserSession(ctx, w, user.ID, oauthRequest)
	if err != nil {
		return nil, err
	}

	auth.TokenType = bearerAuthType

	return auth, nil
}

// createClaims creates the claims for the JWT token using the id for the user and organization
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
func (a *Config) createTokenPair(ctx context.Context, user *generated.User, targetOrgID string) (*models.AuthData, error) {
	orgCheck := targetOrgID
	if orgCheck == "" {
		// get the default org for the user, to use as the target org when one is not provided
		orgCtx := privacy.DecisionContext(ctx, privacy.Allow)

		org, err := user.Edges.Setting.DefaultOrg(orgCtx)
		if err != nil {
			log.Error().Err(err).Msg("error obtaining default org")

			return nil, err
		}

		// add default org to user object
		orgCheck = org.ID
		user.Edges.Setting.Edges.DefaultOrg = org
	}

	newTarget, err := a.authCheck(ctx, user, orgCheck)
	if err != nil {
		if targetOrgID != "" {
			log.Error().Err(err).Msg("user attempting to switch into an org they cannot access, returning error")

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
func (a *Config) generateUserSession(ctx echo.Context, userID string) (string, error) {
	// set sessions in response
	if err := a.db.SessionConfig.CreateAndStoreSession(ctx, userID); err != nil {
		return "", err
	}

	// return the session value for the UI to use
	session, err := sessions.SessionToken(ctx.Request().Context())
	if err != nil {
		return "", err
	}

	return session, nil
}

// generateOauthUserSession creates a new session for the oauth user and stores it in the response
func (a *Config) generateOauthUserSession(ctx context.Context, w http.ResponseWriter, userID string, oauthRequest models.OauthTokenRequest) (string, error) {
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

func (a *Config) authCheck(ctx context.Context, user *generated.User, orgID string) (string, error) {
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return "", nil
	}

	au, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return "", nil
	}

	// if no org is provided, check with the authenticated org
	if orgID == "" {
		orgID = au.OrganizationID
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

	if allow {
		return "", nil
	}

	newOrgID, err := a.updateDefaultOrg(ctx, user, au)
	if err != nil {
		log.Error().Err(err).Msg("unable to update default org")

		return "", err
	}

	return newOrgID, generated.ErrPermissionDenied
}

func (a *Config) updateDefaultOrg(ctx context.Context, user *generated.User, au *auth.AuthenticatedUser) (string, error) {
	// see if we should update the default org to the personal org
	// if the user is not a member of the target org
	// if user.Edges.Setting != nil {
	// 	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	// 	defaultOrg, err := user.Edges.Setting.DefaultOrg(allowCtx)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("unable to get default org")

	// 		return "", err
	// 	}

	// 	if defaultOrg.ID != au.OrganizationID {
	// 		return defaultOrg.ID, err
	// 	}
	// }

	// update default org to personal org
	newOrgID, err := a.updateDefaultOrgToPersonal(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("error updating default org to personal org")

		return "", err
	}

	return newOrgID, nil
}

// updateDefaultOrgToPersonal updates the default org for the user to the personal org
func (a *Config) updateDefaultOrgToPersonal(ctx context.Context, user *generated.User) (string, error) {
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

func (a *Config) getUserDetailsByID(ctx context.Context, userID string) (*generated.User, error) {
	return a.db.User.Query().WithSetting().Where(user.ID(userID)).Only(ctx)
}

// getPersonalOrgID returns the personal org ID for the user
func (a *Config) getPersonalOrgID(ctx context.Context, user *generated.User) (*generated.Organization, error) {
	return a.db.User.QueryOrganizations(user).Where(organization.PersonalOrg(true)).Only(ctx)
}
