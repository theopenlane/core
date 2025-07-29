package handlers

import (
	"context"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
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
	"github.com/theopenlane/core/internal/httpserve/common"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/utils/contextx"
)

// SchemaRegistry interface for dynamic schema registration
type SchemaRegistry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}


// isRegistrationContext checks if the context is a registration context
func isRegistrationContext(ctx echo.Context) bool {
	if ctx == nil || ctx.Request() == nil {
		return false
	}
	// Check if the request context has the registration marker
	_, ok := contextx.From[common.RegistrationMarker](ctx.Request().Context())
	return ok
}

// OpenAPIContext holds the OpenAPI operation and schema registry for automatic registration
type OpenAPIContext struct {
	Operation *openapi3.Operation
	Registry  SchemaRegistry
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
	// IntegrationOauthProvider contains the configuration settings for integration Oauth2 providers
	IntegrationOauthProvider IntegrationOauthProviderConfig
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

// BindAndValidate binds the context payload into T and runs Validate if present.
func BindAndValidate[T any](ctx echo.Context) (*T, error) {
	var obj T
	reqType := reflect.TypeOf(obj).Name()

	ctx.Set(requestTypeKey, reqType)

	if err := ctx.Bind(&obj); err != nil {
		metrics.RequestValidations.WithLabelValues(reqType, "false").Inc()
		return nil, err
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


// BindAndValidateWithRequest registers the request with the OpenAPI specification
// and then binds and validates the payload.
// op may be nil when the handler does not require OpenAPI registration.
func BindAndValidateWithRequest[T any](ctx echo.Context, h *Handler, op *openapi3.Operation, example T) (*T, error) {
	if op != nil {
		AddRequest(h, example, op)
	}

	return BindAndValidate[T](ctx)
}

// BindAndValidateWithAutoRegistry registers the request with the OpenAPI specification using dynamic schema registration
// and then binds and validates the payload.
func BindAndValidateWithAutoRegistry[T any](ctx echo.Context, h *Handler, op *openapi3.Operation, example T, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) (*T, error) {
	if op != nil && registry != nil {
		if err := h.AddRequestBodyWithRegistry(example, op, registry); err != nil {
			return nil, err
		}
	}

	// If we're in OpenAPI registration mode, return nil
	if isRegistrationContext(ctx) {
		return nil, nil
	}

	return BindAndValidate[T](ctx)
}

// validator is an interface that matches objects that can validate themselves
type validator interface {
	Validate() error
}

const requestTypeKey = "requestType"

// AddRequest adds a request body to the OpenAPI schema using the type name of T
func AddRequest[T any](h *Handler, example T, op *openapi3.Operation) {
	var t T
	h.AddRequestBody(reflect.TypeOf(t).Name(), example, op)
}

// AddResponseFor adds a response definition to the OpenAPI schema using the type name of T
func AddResponseFor[T any](h *Handler, description string, example T, op *openapi3.Operation, status int) {
	var t T
	h.AddResponse(reflect.TypeOf(t).Name(), description, example, op, status)
}

// AddRequestWithRegistry adds a request body to the OpenAPI schema with automatic type registration
func AddRequestWithRegistry[T any](h *Handler, example T, op *openapi3.Operation, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) error {
	return h.AddRequestBodyWithRegistry(example, op, registry)
}

// AddResponseWithRegistry adds a response definition to the OpenAPI schema with automatic type registration
func AddResponseWithRegistry[T any](h *Handler, description string, example T, op *openapi3.Operation, status int, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) error {
	return h.AddResponseWithRegistry(description, example, op, status, registry)
}
