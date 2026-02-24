package handlers

import (
	"context"
	"net/http"
	"reflect"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/redis/go-redis/v9"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/common"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/shortlinks"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/utils/rout"
)

// ProviderRegistry exposes provider runtimes and metadata to handlers.
type ProviderRegistry interface {
	Provider(provider types.ProviderType) (types.Provider, bool)
	Config(provider types.ProviderType) (config.ProviderSpec, bool)
	ProviderMetadataCatalog() map[types.ProviderType]types.ProviderConfig
	OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor
}

// SchemaRegistry interface for dynamic schema registration
type SchemaRegistry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}

// isRegistrationContext checks if the context is a registration context
func isRegistrationContext(ctx echo.Context) bool {
	if ctx == nil || ctx.Request() == nil {
		return false
	}

	return common.IsRegistrationContext(ctx.Request().Context())
}

// OpenAPIContext holds the OpenAPI operation and schema registry for automatic registration
type OpenAPIContext struct {
	// Operation is the OpenAPI operation metadata for the handler.
	Operation *openapi3.Operation
	// Registry provides schema registration and lookup for OpenAPI models.
	Registry SchemaRegistry
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
	// IntegrationGitHubApp contains the configuration settings for GitHub App integrations
	IntegrationGitHubApp IntegrationGitHubAppConfig
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
	// DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set
	DefaultTrustCenterDomain string
	// ObjectStore handles file storage operations
	ObjectStore *objects.Service
	// IntegrationRegistry contains the declarative provider runtimes for third-party integrations
	IntegrationRegistry ProviderRegistry
	// IntegrationStore handles persistence for integration metadata and secrets
	IntegrationStore *keystore.Store
	// IntegrationBroker exchanges credentials for provider access tokens
	IntegrationBroker *keystore.Broker
	// IntegrationClients reuses provider SDK clients via pooled builders
	IntegrationClients *keystore.ClientPoolManager
	// IntegrationOperations standardizes executing provider operations
	IntegrationOperations *keystore.OperationManager
	// Gala is the shared event runtime for asynchronous dispatch.
	Gala *gala.Gala
	// IntegrationActivation orchestrates integration activation flows
	IntegrationActivation *activation.Service
	// WorkflowEngine orchestrates workflow execution.
	WorkflowEngine *engine.WorkflowEngine
	// CampaignWebhook contains the configuration for campaign-related email webhooks
	CampaignWebhook CampaignWebhookConfig
	// CloudflareConfig contains the configuration for Cloudflare integration
	CloudflareConfig CloudflareConfig
	// ShortlinksClient provides URL shortening functionality
	ShortlinksClient *shortlinks.Client
}

// CampaignWebhookConfig contains webhook configuration for campaign-related email providers.
type CampaignWebhookConfig struct {
	// Enabled toggles the campaign webhook handler
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// ResendAPIKey is the API key used for Resend client initialization
	ResendAPIKey string `json:"resendapikey" koanf:"resendapikey" default:"" sensitive:"true"`
	// ResendSecret is the signing secret used to verify Resend webhook payloads
	ResendSecret string `json:"resendsecret" koanf:"resendsecret" default:"" sensitive:"true"`
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

// opMutex protects OpenAPI registration operations
var opMutex sync.Mutex

// BindAndValidateWithAutoRegistry registers the request and response with the OpenAPI specification using dynamic schema registration
// and then binds and validates the payload. This automatically detects response examples using the ExampleProvider interface.
func BindAndValidateWithAutoRegistry[T any, R any](ctx echo.Context, _ *Handler, op *openapi3.Operation, requestExample T, responseExample R, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) (*T, error) {
	if isRegistrationContext(ctx) && op != nil && registry != nil {
		// lock to prevent concurrent map writes
		opMutex.Lock()
		defer opMutex.Unlock()

		// Auto-detect and register path parameters from struct tags
		registerPathParameters(requestExample, op)

		// Only add request body for non-GET methods
		if ctx.Request().Method != http.MethodGet {
			// Register request body schema dynamically
			schemaRef, err := registry.GetOrRegister(requestExample)
			if err != nil {
				return nil, err
			}

			request := openapi3.NewRequestBody().
				WithDescription("Request body").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			op.RequestBody = &openapi3.RequestBodyRef{Value: request}

			request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{
				Value: openapi3.NewExample(normalizeExampleValue(requestExample)),
			}
		}

		// Register success response schema dynamically
		var exampleObject any = responseExample

		// If the response implements ExampleProvider, use its example for the OpenAPI spec
		if provider, ok := any(&responseExample).(models.ExampleProvider); ok {
			exampleObject = provider.ExampleResponse()
		}

		if schemaRef, err := registry.GetOrRegister(responseExample); err == nil {
			response := openapi3.NewResponse().
				WithDescription("Success").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			op.AddResponse(http.StatusOK, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{
				Value: openapi3.NewExample(normalizeExampleValue(exampleObject)),
			}
		}
	}

	// If we're in OpenAPI registration mode, return the example object
	if isRegistrationContext(ctx) {
		return &requestExample, nil
	}

	return BindAndValidate[T](ctx)
}

// registerPathParameters automatically detects path parameters from struct tags and registers them with the OpenAPI operation
func registerPathParameters[T any](example T, op *openapi3.Operation) {
	t := reflect.TypeOf(example)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process struct types
	if t.Kind() != reflect.Struct {
		return
	}

	// Iterate through struct fields looking for param tags
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		paramTag := field.Tag.Get("param")

		if paramTag != "" {
			// Create path parameter
			param := openapi3.NewPathParameter(paramTag).WithSchema(openapi3.NewStringSchema())

			// Add description if available
			if description := field.Tag.Get("description"); description != "" {
				param.Description = description
			}

			// Add example if available
			if example := field.Tag.Get("example"); example != "" {
				param.Example = example
			}

			// Add parameter to operation
			op.AddParameter(param)
		}
	}
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

// ProcessRequest provides a generic pattern for handling requests with automatic binding, validation, and response handling
func ProcessRequest[TReq, TResp any](ctx echo.Context, h *Handler, openapi *OpenAPIContext, requestExample TReq, responseExample TResp, processor func(context.Context, *TReq) (*TResp, error)) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, requestExample, responseExample, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Process the request using the provided processor function
	resp, err := processor(ctx.Request().Context(), req)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	// Return successful response
	return h.Success(ctx, resp, openapi)
}

// ProcessAuthenticatedRequest provides a generic pattern for authenticated requests with automatic caller context injection
func ProcessAuthenticatedRequest[TReq, TResp any](ctx echo.Context, h *Handler, openapi *OpenAPIContext, requestExample TReq, responseExample TResp, processor func(context.Context, *TReq, *auth.Caller) (*TResp, error)) error {
	// Bind and validate the request
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, requestExample, responseExample, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	// Get caller from context
	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting caller from context")
		return h.InternalServerError(ctx, auth.ErrNoAuthUser, openapi)
	}

	// Process the request with caller context
	resp, err := processor(reqCtx, req, caller)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	// Return successful response
	return h.Success(ctx, resp, openapi)
}

// BindAndValidateQueryParams binds and validates query parameters for GET requests and registers them in the OpenAPI schema
// For backwards compatibility - use BindAndValidateQueryParamsWithResponse for new code
func BindAndValidateQueryParams[T any](ctx echo.Context, op *openapi3.Operation, example T, registry SchemaRegistry) (*T, error) {
	return BindAndValidateQueryParamsWithResponse(ctx, op, example, rout.Reply{}, registry)
}

// BindAndValidateQueryParamsWithResponse binds and validates query parameters and registers both request and response schemas
func BindAndValidateQueryParamsWithResponse[T any, R any](ctx echo.Context, op *openapi3.Operation, requestExample T, responseExample R, registry SchemaRegistry) (*T, error) {
	// Register query parameters and response schema in OpenAPI schema during registration
	if isRegistrationContext(ctx) && op != nil && registry != nil {
		// lock to prevent concurrent map writes
		opMutex.Lock()
		defer opMutex.Unlock()

		// Use reflection to extract query parameter information
		typ := reflect.TypeOf(requestExample)

		// Handle pointer types
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		// Add query parameters to OpenAPI operation
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			queryTag := field.Tag.Get("query")
			jsonTag := field.Tag.Get("json")
			description := field.Tag.Get("description")
			exampleTag := field.Tag.Get("example")

			if queryTag != "" {
				param := &openapi3.Parameter{
					Name:        queryTag,
					In:          "query",
					Description: description,
					Schema: &openapi3.SchemaRef{
						Value: openapi3.NewStringSchema(),
					},
				}

				if exampleTag != "" {
					param.Example = exampleTag
				}

				// Add parameter to operation
				if op.Parameters == nil {
					op.Parameters = make([]*openapi3.ParameterRef, 0)
				}

				op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: param})
			}

			if jsonTag != "" {
				// Handle path parameters (when json tag exists but query tag doesn't, it might be a path param)
				paramTag := field.Tag.Get("param")
				if paramTag != "" {
					param := &openapi3.Parameter{
						Name:        paramTag,
						In:          "path",
						Required:    true,
						Description: description,
						Schema: &openapi3.SchemaRef{
							Value: openapi3.NewStringSchema(),
						},
					}

					if exampleTag != "" {
						param.Example = exampleTag
					}

					// Add parameter to operation
					if op.Parameters == nil {
						op.Parameters = make([]*openapi3.ParameterRef, 0)
					}

					op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: param})
				}
			}
		}

		// Register success response schema dynamically
		var exampleObject any = responseExample

		// If the response implements ExampleProvider, use its example for the OpenAPI spec
		if provider, ok := any(&responseExample).(models.ExampleProvider); ok {
			exampleObject = provider.ExampleResponse()
		}

		if schemaRef, err := registry.GetOrRegister(responseExample); err == nil {
			response := openapi3.NewResponse().
				WithDescription("Success").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			op.AddResponse(http.StatusOK, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(exampleObject))}
		}
	}

	// If we're in OpenAPI registration mode, return the example object
	if isRegistrationContext(ctx) {
		return &requestExample, nil
	}

	// Bind query parameters to the struct
	var req T

	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}

	// Validate the bound request if it implements the validator interface
	if v, ok := any(&req).(validator); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	return &req, nil
}
