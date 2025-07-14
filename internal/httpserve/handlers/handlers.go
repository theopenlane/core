package handlers

import (
	"context"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
)

// Handler contains configuration options for handlers
type Handler struct {
	// IsTest is a flag to determine if the application is running in test mode and will mock external calls
	IsTest bool
	// IsDev is a flag to determine if the application is running in development mode
	IsDev bool
	// DBClient to interact with the database
	DBClient *ent.Client
	// RedisClient to interact with redis
	RedisClient *redis.Client
	// AuthManager contains the required configuration for the auth session creation
	AuthManager *authmanager.Client
	// TokenManager contains the token manager in order to validate auth requests
	TokenManager *tokens.TokenManager
	// ReadyChecks is a set of checkFuncs to determine if the application is "ready" upon startup
	ReadyChecks Checks
	// JWTKeys contains the set of valid JWT authentication key
	JWTKeys jwk.Set
	// SessionConfig to handle sessions
	SessionConfig *sessions.SessionConfig
	// OauthProvider contains the configuration settings for all supported Oauth2 providers
	OauthProvider OauthProviderConfig
	// AuthMiddleware contains the middleware to be used for authenticated endpoints
	AuthMiddleware []echo.MiddlewareFunc
	// AdditionalMiddleware contains the additional middleware to be used for all endpoints
	// it is separate so it can be applied after any auth middleware if needed
	AdditionalMiddleware []echo.MiddlewareFunc
	// WebAuthn contains the configuration settings for the webauthn provider
	WebAuthn *webauthn.WebAuthn
	// OTPManager contains the configuration settings for the OTP provider
	OTPManager *totp.Client
	// Email contains email sending configuration for the server
	Emailer emailtemplates.Config
	// Entitlements contains the entitlements client
	Entitlements *entitlements.StripeClient
	// Summarizer contains the summarizing client
	Summarizer *summarizer.Client
	// Windmill contains the Windmill workflow automation client
	Windmill *windmill.Client
	// DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set
	DefaultTrustCenterDomain string
}

// setAuthenticatedContext is a wrapper that will set the minimal context for an authenticated user
// during a login or verification process
func setAuthenticatedContext(ctx context.Context, user *ent.User) context.Context {
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
		SubjectID:    user.ID,
		SubjectEmail: user.Email,
	})

	return ctx
}
