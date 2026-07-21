package handlers

import (
	"context"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/integrations/definitions/catalog"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/shortlinks"
	"github.com/theopenlane/core/pkg/summarizer"
)

// SchemaRegistry interface for dynamic schema registration
type SchemaRegistry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}

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
	// OauthProvider contains the configuration settings for all supported Oauth2 providers (for social login)
	OauthProvider OauthProviderConfig
	// ConsoleURL is the full base frontend URL (e.g. https://console.example.com) used for browser redirects after auth flows
	ConsoleURL string
	// AuthMiddleware contains the middleware to be used for authenticated endpoints
	AuthMiddleware []echo.MiddlewareFunc
	// AdditionalMiddleware contains the additional middleware to be used for all endpoints
	// it is separate so it can be applied after any auth middleware if needed
	AdditionalMiddleware []echo.MiddlewareFunc
	// WebAuthn contains the configuration settings for the webauthn provider
	WebAuthn *webauthn.WebAuthn
	// OTPManager contains the configuration settings for the OTP provider
	OTPManager *totp.Client
	// Entitlements contains the entitlements client
	Entitlements *entitlements.StripeClient
	// Summarizer contains the summarizing client
	Summarizer *summarizer.Client
	// DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set
	DefaultTrustCenterDomain string
	// ObjectStore handles file storage operations
	ObjectStore *objects.Service
	// IntegrationsRuntime holds the integration runtime components.
	IntegrationsRuntime *integrationsruntime.Runtime
	// IntegrationsConfig contains environment-backed operator configuration for built-in integrations.
	IntegrationsConfig catalog.Config
	// Gala is the shared event runtime for asynchronous dispatch.
	Gala *gala.Gala
	// WorkflowEngine orchestrates workflow execution.
	WorkflowEngine *engine.WorkflowEngine
	// CloudflareConfig contains the configuration for Cloudflare integration
	CloudflareConfig CloudflareConfig
	// ShortlinksClient provides URL shortening functionality
	ShortlinksClient *shortlinks.Client
	// SupportAccessConfig contains the configuration for the Openlane support access flow
	SupportAccessConfig SupportAccessConfig
}

// SupportAccessConfig contains configuration for the Openlane support access flow. The support
// identity is virtual and authenticated entirely from these values, never from the database. This is
// the single place that holds the support identity, its shared password, and the second factor
// identity provider configuration, since both authentications must occur together
type SupportAccessConfig struct {
	// Enabled toggles the support access endpoints
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Email is the email of the virtual support identity, used as the first factor username
	Email string `json:"email" koanf:"email" default:"support@theopenlane.io"`
	// DisplayName is the display name of the virtual support identity, used for record attribution
	DisplayName string `json:"displayname" koanf:"displayname" default:"Openlane Support"`
	// SubjectID is the stable subject id of the virtual support identity used for created_by/updated_by
	// attribution. It must be a valid ULID and is consistent without a backing user row. Default should match anon.SupportSubjectID
	SubjectID string `json:"subjectid" koanf:"subjectid" default:"01JSPPRT000000000000000000"`
	// Password is the shared password for the virtual support identity, validated against this value
	Password string `json:"password" koanf:"password" default:"" sensitive:"true"`
	// ClientID is the client ID for the second factor identity provider
	ClientID string `json:"clientid" koanf:"clientid" default:"" sensitive:"true"`
	// ClientSecret is the client secret for the second factor identity provider
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" default:"" sensitive:"true"`
	// IssuerURL is the issuer URL of the second factor identity provider
	IssuerURL string `json:"issuerurl" koanf:"issuerurl" default:""`
	// DiscoveryEndpoint is the optional OIDC discovery endpoint of the second factor identity provider
	DiscoveryEndpoint string `json:"discoveryendpoint" koanf:"discoveryendpoint" default:""`
	// RedirectURL is the callback URL registered with the second factor identity provider
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:""`
	// AllowedDomain restricts which email domain may complete the second factor (e.g. theopenlane.io)
	AllowedDomain string `json:"alloweddomain" koanf:"alloweddomain" default:""`
}

// CloudflareConfig contains configuration for Cloudflare integration.
type CloudflareConfig struct {
	// Enabled toggles the Cloudflare snapshot handler
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// APIToken is the API token used for Cloudflare client initialization
	APIToken string `json:"apitoken" koanf:"apitoken" default:"" sensitive:"true"`
	// AccountID is the Cloudflare account ID to use for snapshot operations
	AccountID string `json:"accountid" koanf:"accountid" default:"" sensitive:"true"`
	// ClientID is the Cloudflare Access client ID for shortlink API requests
	ClientID string `json:"clientid" koanf:"clientid" default:"" sensitive:"true"`
	// ClientSecret is the Cloudflare Access client secret for shortlink API requests
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" default:"" sensitive:"true"`
}

// setAuthenticatedContext is a wrapper that will set the minimal context for an authenticated user
// during a login or verification process
func setAuthenticatedContext(ctx context.Context, user *ent.User) context.Context {
	ctx = auth.WithCaller(ctx, &auth.Caller{
		SubjectID:    user.ID,
		SubjectEmail: user.Email,
	})

	return ctx
}

// queryParamBinder opts a request model into query-param binding on any HTTP method; echo's binder
// only binds query params on GET/DELETE/HEAD
type queryParamBinder interface {
	BindsQueryParams() bool
}

// BindAndValidate binds the context payload into T and runs Validate if present.
func BindAndValidate[T any](ctx echo.Context) (*T, error) {
	var obj T

	reqType := reflect.TypeOf(obj).Name()

	ctx.Set(requestTypeKey, reqType)

	if err := ctx.Bind(&obj); err != nil {
		metrics.RequestValidations.WithLabelValues(reqType, "false").Inc()
		return nil, err
	}

	if qb, ok := any(&obj).(queryParamBinder); ok && qb.BindsQueryParams() {
		if err := echo.BindQueryParams(ctx, &obj); err != nil {
			metrics.RequestValidations.WithLabelValues(reqType, "false").Inc()
			return nil, err
		}
	}

	if v, ok := any(&obj).(validator); ok {
		if err := v.Validate(); err != nil {
			metrics.RequestValidations.WithLabelValues(reqType, "false").Inc()
			return nil, err
		}
	}

	metrics.RequestValidations.WithLabelValues(reqType, "true").Inc()

	return &obj, nil
}

// validator is an interface that matches objects that can validate themselves
type validator interface {
	Validate() error
}

const requestTypeKey = "requestType"

// ProcessAuthenticatedRequest provides a generic pattern for authenticated requests with automatic caller context injection
func ProcessAuthenticatedRequest[TReq, TResp any](ctx echo.Context, h *Handler, processor func(context.Context, *TReq, *auth.Caller) (*TResp, error)) error {
	// Bind and validate the request
	req, err := BindAndValidate[TReq](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	// Get caller from context
	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting caller from context")
		return h.InternalServerError(ctx, auth.ErrNoAuthUser)
	}

	// Process the request with caller context
	resp, err := processor(reqCtx, req, caller)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// Return successful response
	return h.Success(ctx, resp)
}
