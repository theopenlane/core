package handlers

import (
	"context"
	"net/http"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/common"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
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

// CheckRegistrationModeWithResponse checks if we're in registration mode and returns early with nil
// This should be called at the beginning of handlers to skip execution during OpenAPI generation
func CheckRegistrationModeWithResponse(ctx echo.Context) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	return nil
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

// BindAndValidateWithAutoRegistry registers the request and response with the OpenAPI specification using dynamic schema registration
// and then binds and validates the payload. This automatically detects response examples using the ExampleProvider interface.
func BindAndValidateWithAutoRegistry[T any, R any](ctx echo.Context, _ *Handler, op *openapi3.Operation, requestExample T, responseExample R, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) (*T, error) {
	if op != nil && registry != nil {
		// Auto-detect and register path parameters from struct tags
		registerPathParameters(requestExample, op)

		// Only add request body for non-GET methods
		if ctx.Request().Method != http.MethodGet {
			// Register request body schema dynamically
			schemaRef, err := registry.GetOrRegister(requestExample)
			if err != nil {
				return nil, err
			}

			request := openapi3.NewRequestBody().WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			op.RequestBody = &openapi3.RequestBodyRef{Value: request}

			request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(requestExample)}
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
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(exampleObject)}
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

// AddResponseFor adds a response definition to the OpenAPI schema using the type name of T
func AddResponseFor[T any](h *Handler, description string, example T, op *openapi3.Operation, status int) {
	var t T
	h.AddResponse(reflect.TypeOf(t).Name(), description, example, op, status)
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

// ProcessAuthenticatedRequest provides a generic pattern for authenticated requests with automatic user context injection
func ProcessAuthenticatedRequest[TReq, TResp any](ctx echo.Context, h *Handler, openapi *OpenAPIContext, requestExample TReq, responseExample TResp, processor func(context.Context, *TReq, *auth.AuthenticatedUser) (*TResp, error)) error {
	// Bind and validate the request
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, requestExample, responseExample, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Get authenticated user from context
	reqCtx := ctx.Request().Context()
	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")
		return h.InternalServerError(ctx, err, openapi)
	}

	// Process the request with authenticated user context
	resp, err := processor(reqCtx, req, au)
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
	if op != nil && registry != nil {
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
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(exampleObject)}
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
