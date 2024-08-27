package auth

import (
	"context"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/middleware/echocontext"
	"github.com/theopenlane/utils/ulids"
)

type AuthenticationType string

const (
	// JWTAuthentication is the authentication type for JWT tokens
	JWTAuthentication AuthenticationType = "jwt"
	// PATAuthentication is the authentication type for personal access tokens
	PATAuthentication AuthenticationType = "pat"
	// APITokenAuthentication is the authentication type for API tokens
	APITokenAuthentication AuthenticationType = "api_token"
)

// ContextAuthenticatedUser is the context key for the user claims
var ContextAuthenticatedUser = &ContextKey{"authenticated_user"}

// ContextAccessToken is the context key for the access token
var ContextAccessToken = &ContextKey{"access_token"}

// ContextAccessToken is the context key for the access token
var ContextRefreshToken = &ContextKey{"refresh_token"}

// ContextRequestID is the context key for the request ID
var ContextRequestID = &ContextKey{"request_id"}

// ContextKey is the key name for the additional context
type ContextKey struct {
	name string
}

// AuthenticatedUser contains the user and organization ID for the authenticated user
type AuthenticatedUser struct {
	// SubjectID is the user ID of the authenticated user or the api token ID if the user is an API token
	SubjectID string
	// SubjectName is the name of the authenticated user
	SubjectName string
	// OrganizationID is the organization ID of the authenticated user
	OrganizationID string
	// OrganizationName is the name of the organization the user is authenticated to
	OrganizationName string
	// OrganizationIDs is the list of organization IDs the user is authorized to access
	OrganizationIDs []string
	// AuthenticationType is the type of authentication used to authenticate the user (JWT, PAT, API Token)
	AuthenticationType AuthenticationType
}

// GetContextName returns the name of the context key
func GetContextName(key *ContextKey) string {
	return key.name
}

// SetAuthenticatedUserContext sets the authenticated user context in the echo context
func SetAuthenticatedUserContext(c echo.Context, user *AuthenticatedUser) {
	c.Set(ContextAuthenticatedUser.name, user)
}

// GetAuthenticatedUserContext gets the authenticated user context
func GetAuthenticatedUserContext(c context.Context) (*AuthenticatedUser, error) {
	ec, err := echocontext.EchoContextFromContext(c)
	if err != nil {
		return nil, err
	}

	result := ec.Get(ContextAuthenticatedUser.name)

	au, ok := result.(*AuthenticatedUser)
	if !ok {
		return nil, ErrNoAuthUser
	}

	return au, nil
}

// AddAuthenticatedUserContext adds the authenticated user context and returns the context
func AddAuthenticatedUserContext(c echo.Context, user *AuthenticatedUser) context.Context {
	c.Set(ContextAuthenticatedUser.name, user)

	ctx := context.WithValue(c.Request().Context(), echocontext.EchoContextKey, c)

	c.SetRequest(c.Request().WithContext(ctx))

	return ctx
}

// GetAuthTypeFromEchoContext retrieves the authentication type from the echo context
func GetAuthTypeFromEchoContext(c echo.Context) AuthenticationType {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if ok {
			return a.AuthenticationType
		}
	}

	return ""
}

// getSubjectIDFromEchoContext retrieves the subject ID from the echo context
func getSubjectIDFromEchoContext(c echo.Context) (string, error) {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return "", ErrNoAuthUser
		}

		uid, err := ulids.Parse(a.SubjectID)
		if err != nil {
			return "", err
		}

		if ulids.IsZero(uid) {
			return "", ErrNoAuthUser
		}

		return a.SubjectID, nil
	}

	return "", ErrNoAuthUser
}

// getSubjectNameFromEchoContext retrieves the subject name from the echo context
func getSubjectNameFromEchoContext(c echo.Context) (string, error) {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return "", ErrNoAuthUser
		}

		return a.SubjectName, nil
	}

	return "", ErrNoAuthUser
}

// getOrganizationIDFromEchoContext returns the organization ID from the echo context
func getOrganizationIDFromEchoContext(c echo.Context) (string, error) {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return "", ErrNoAuthUser
		}

		oID, err := ulids.Parse(a.OrganizationID)
		if err != nil {
			return "", err
		}

		if ulids.IsZero(oID) {
			return "", ErrNoAuthUser
		}

		return a.OrganizationID, nil
	}

	return "", ErrNoAuthUser
}

// getOrganizationIDsFromEchoContext returns the list of organization IDs from the echo context
func getOrganizationIDsFromEchoContext(c echo.Context) ([]string, error) {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return []string{}, ErrNoAuthUser
		}

		return a.OrganizationIDs, nil
	}

	return []string{}, ErrNoAuthUser
}

// getOrganizationNameFromEchoContext returns the organization name from the echo context
func getOrganizationNameFromEchoContext(c echo.Context) (string, error) {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return "", ErrNoAuthUser
		}

		return a.OrganizationName, nil
	}

	return "", ErrNoAuthUser
}

// GetOrganizationIDFromContext returns the organization ID from context
func GetOrganizationIDFromContext(ctx context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	return getOrganizationIDFromEchoContext(ec)
}

// GetOrganizationNameFromContext returns the organization name from context
func GetOrganizationNameFromContext(ctx context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	return getOrganizationNameFromEchoContext(ec)
}

// GetOrganizationIDsFromContext returns the list of organization IDs from context
func GetOrganizationIDsFromContext(ctx context.Context) ([]string, error) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return []string{}, err
	}

	return getOrganizationIDsFromEchoContext(ec)
}

// GetUserIDFromContext returns the actor subject from the context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	return getSubjectIDFromEchoContext(ec)
}

// GetUserNameFromContext returns the actor name from the context
func GetUserNameFromContext(ctx context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	return getSubjectNameFromEchoContext(ec)
}

// GetAuthTypeFromEchoContext retrieves the authentication type from the context
func GetAuthTypeFromContext(ctx context.Context) AuthenticationType {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return ""
	}

	return GetAuthTypeFromEchoContext(ec)
}

func IsAPITokenAuthentication(ctx context.Context) bool {
	return GetAuthTypeFromContext(ctx) == APITokenAuthentication
}

const (
	// UserSubjectType is the subject type for user accounts
	UserSubjectType = "user"
	// ServiceSubjectType is the subject type for service accounts
	ServiceSubjectType = "service"
)

// GetAuthzSubjectType returns the subject type based on the authentication type
func GetAuthzSubjectType(ctx context.Context) string {
	subjectType := UserSubjectType

	authType := GetAuthTypeFromContext(ctx)
	if authType == APITokenAuthentication {
		subjectType = ServiceSubjectType
	}

	return subjectType
}

// SetOrganizationIDInAuthContext sets the organization ID in the auth context
// this should only be used when creating a new organization and subsequent updates
// need to happen in the context of the new organization
func SetOrganizationIDInAuthContext(ctx context.Context, orgID string) error {
	au, err := GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	au.OrganizationID = orgID

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	SetAuthenticatedUserContext(ec, au)

	return nil
}

// AddOrganizationIDToContext appends an authorized organization ID to the context.
// This generally should not be used, as the authorized organization should be
// determined by the claims or the token. This is only used in cases where the
// a user is newly authorized to an organization and the organization ID is not
// in the token claims
func AddOrganizationIDToContext(ctx context.Context, orgID string) error {
	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	return addOrganizationIDsToEchoContext(ec, orgID)
}

// getOrganizationIDsFromEchoContext appends an authorized organization ID to the echo context
func addOrganizationIDsToEchoContext(c echo.Context, orgID string) error {
	if v := c.Get(ContextAuthenticatedUser.name); v != nil {
		a, ok := v.(*AuthenticatedUser)
		if !ok {
			return ErrNoAuthUser
		}

		// append the organization ID to the list of organization IDs
		a.OrganizationIDs = append(a.OrganizationIDs, orgID)

		return nil
	}

	return ErrNoAuthUser
}

// SetRefreshTokenContext sets the refresh token context in the echo context
func SetRefreshTokenContext(c echo.Context, token string) {
	c.Set(ContextRefreshToken.name, token)
}

// SetAccessTokenContext sets the access token context in the echo context
func SetAccessTokenContext(c echo.Context, token string) {
	c.Set(ContextAccessToken.name, token)
}

// GetAccessTokenContext gets the authenticated user context
func GetAccessTokenContext(c context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(c)
	if err != nil {
		return "", err
	}

	result := ec.Get(ContextAccessToken.name)

	token, ok := result.(string)
	if !ok {
		return "", ErrNoAuthUser
	}

	return token, nil
}

// GetRefreshTokenContext gets the authenticated user context
func GetRefreshTokenContext(c context.Context) (string, error) {
	ec, err := echocontext.EchoContextFromContext(c)
	if err != nil {
		return "", err
	}

	result := ec.Get(ContextRefreshToken.name)

	token, ok := result.(string)
	if !ok {
		return "", ErrNoAuthUser
	}

	return token, nil
}
