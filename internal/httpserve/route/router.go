package route

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/internal/httpserve/common"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/middleware/impersonation"
	"github.com/theopenlane/core/pkg/middleware/mime"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/utils/contextx"
)

// convertEchoPathToOpenAPI converts Echo's :param syntax to OpenAPI's {param} syntax
func convertEchoPathToOpenAPI(echoPath string) string {
	// Split the path into parts and convert :param to {param}
	parts := strings.Split(echoPath, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			// Convert :param to {param}
			parts[i] = "{" + part[1:] + "}"
		}
	}

	return strings.Join(parts, "/")
}

// addPathParametersFromPattern extracts path parameters from Echo-style path and adds them to OpenAPI operation
func (r *Router) addPathParametersFromPattern(path string, operation *openapi3.Operation) {
	// Extract parameter names from Echo-style path (e.g., :id, :name)
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:] // Remove the : prefix

			// Check if parameter already exists (e.g., from struct tags)
			exists := false

			if operation.Parameters != nil {
				for _, p := range operation.Parameters {
					if p.Value != nil && p.Value.Name == paramName && p.Value.In == "path" {
						exists = true
						break
					}
				}
			}

			// Only add if it doesn't already exist
			if !exists {
				// Add path parameter to the operation
				param := openapi3.NewPathParameter(paramName).
					WithSchema(openapi3.NewStringSchema()).
					WithDescription(fmt.Sprintf("Path parameter: %s", paramName))
				operation.AddParameter(param)
			}
		}
	}
}

// registerSuccessResponseSchemas registers success response schemas dynamically
// This completely eliminates static mappings by allowing handlers to register their own response types
func (r *Router) registerSuccessResponseSchemas(config Config, openAPIContext *handlers.OpenAPIContext) {
	if openAPIContext.Registry == nil {
		return
	}

	// For special non-JSON responses, handle them explicitly
	r.handleSpecialResponses(config.OperationID, openAPIContext)
}

// handleSpecialResponses handles non-JSON response cases only
func (r *Router) handleSpecialResponses(operationID string, openAPIContext *handlers.OpenAPIContext) {
	switch operationID {
	// Non-JSON responses only - these don't use h.Success() so need explicit registration
	case "AcmeSolver":
		r.addPlainTextResponse(openAPIContext, "ACME challenge response")
	case "AppleMerchant", "SecurityTxt", "WebAuthnWellKnown", "Robots", "Favicon":
		r.addFileResponse(openAPIContext, "Static file content")
	case "JWKS":
		r.addJSONResponse(openAPIContext, "JSON Web Key Set", "application/json")
	case "APIDocs":
		r.addJSONResponse(openAPIContext, "OpenAPI specification", "application/json")
	case "CSRF":
		r.addJSONResponse(openAPIContext, "CSRF token", "application/json")
	case "ExampleCSV":
		r.addFileResponse(openAPIContext, "CSV file content")
	case "Files":
		r.addFileResponse(openAPIContext, "File content")
	case "GitHubCallback", "GitHubLogin", "GoogleCallback", "GoogleLogin":
		r.addRedirectResponse(openAPIContext, "OAuth redirect")
	case "UserInfo":
		r.addJSONResponse(openAPIContext, "User information", "application/json")
	case "Livez", "Ready":
		r.addJSONResponse(openAPIContext, "Health check status", "application/json")
	case "StripeWebhook", "ResendWebhook":
		r.addPlainTextResponse(openAPIContext, "Webhook acknowledgment")
	default:
		// All other endpoints register their responses via h.Success() calls during registration
	}
}

// addPlainTextResponse adds a plain text success response to the operation
func (r *Router) addPlainTextResponse(openAPIContext *handlers.OpenAPIContext, description string) {
	response := openapi3.NewResponse().
		WithDescription(description).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"text/plain"}))
	openAPIContext.Operation.AddResponse(http.StatusOK, response)
}

// addJSONResponse adds a JSON success response to the operation
func (r *Router) addJSONResponse(openAPIContext *handlers.OpenAPIContext, description, contentType string) {
	response := openapi3.NewResponse().
		WithDescription(description).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewObjectSchema(), []string{contentType}))
	openAPIContext.Operation.AddResponse(http.StatusOK, response)
}

// addFileResponse adds a file content success response to the operation
func (r *Router) addFileResponse(openAPIContext *handlers.OpenAPIContext, description string) {
	response := openapi3.NewResponse().
		WithDescription(description).
		WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{"application/octet-stream"}))
	openAPIContext.Operation.AddResponse(http.StatusOK, response)
}

// addRedirectResponse adds a redirect success response to the operation
func (r *Router) addRedirectResponse(openAPIContext *handlers.OpenAPIContext, description string) {
	response := openapi3.NewResponse().WithDescription(description)
	openAPIContext.Operation.AddResponse(http.StatusFound, response)
}

var (
	// baseMW includes the basic middleware, which includes the transaction middleware and recovery middleware, most endpoints will not use this, but use `mw` instead
	baseMW = []echo.MiddlewareFunc{}
	// mw is the default middleware that is applied to all routes, it includes the transaction middleware and any additional middleware (including csrf)
	// this is used for most routes that are not authenticated or restricted
	mw = []echo.MiddlewareFunc{}
	// authMW is the middleware that is used on authenticated routes, it includes the transaction middleware, the auth middleware, and any additional middleware after the auth middleware
	authMW = []echo.MiddlewareFunc{}
)

// Middleware Semantic Names for better readability
var (
	// authenticatedEndpoint for endpoints requiring authentication
	authenticatedEndpoint = &authMW
	// publicEndpoint for standard public endpoints
	publicEndpoint = &mw
	// unauthenticatedEndpoint for basic endpoints with minimal middleware
	unauthenticatedEndpoint = &baseMW
)

// Router is a struct that holds the echo router, the OpenAPI schema, and the handler - it's a way to group these components together
type Router struct {
	Echo           *echo.Echo
	OAS            *openapi3.T
	Handler        *handlers.Handler
	StartConfig    *echo.StartConfig
	LocalFilePath  string
	Logger         *echo.Logger
	SchemaRegistry SchemaRegistry
}

// SchemaRegistry interface for dynamic schema registration
type SchemaRegistry interface {
	RegisterType(v any) (*openapi3.SchemaRef, error)
	GetOrRegister(v any) (*openapi3.SchemaRef, error)
}

// RouterOption is an option function that can be used to configure the router
type RouterOption func(*Router)

// WithLogger is a RouterOption that allows the logger to be set on the router
func WithLogger(logger *echo.Logger) RouterOption {
	return func(r *Router) {
		r.Logger = logger
	}
}

// WithHandler is a RouterOption that allows the handler to be set on the router
func WithHandler(h *handlers.Handler) RouterOption {
	return func(r *Router) {
		r.Handler = h
	}
}

// WithEcho is a RouterOption that allows the echo router to be set on the router
func WithEcho(e *echo.Echo) RouterOption {
	return func(r *Router) {
		r.Echo = e
	}
}

// WithLocalFiles is a RouterOption that allows the local files to be set on the router
func WithLocalFiles(lf string) RouterOption {
	return func(r *Router) {
		r.LocalFilePath = lf
	}
}

// WithOpenAPI is a RouterOption that allows the OpenAPI schema to be set on the router
func WithOpenAPI(oas *openapi3.T) RouterOption {
	return func(r *Router) {
		r.OAS = oas
	}
}

// WithOptions is a RouterOption that allows multiple options to be set on the router
func WithOptions(opts ...RouterOption) RouterOption {
	return func(r *Router) {
		for _, opt := range opts {
			opt(r)
		}
	}
}

// WithHideBanner is a RouterOption that allows the banner to be hidden on the echo server
func WithHideBanner() RouterOption {
	return func(r *Router) {
		r.StartConfig = &echo.StartConfig{
			HideBanner: true,
		}
	}
}

// AddRoute is used to add a route to the echo router and OpenAPI schema at the same time ensuring consistency between the spec and the server
func (r *Router) AddRoute(pattern, method string, op *openapi3.Operation, route echo.Routable) error {
	_, err := r.Echo.AddRoute(route)
	if err != nil {
		return err
	}

	// Convert Echo path syntax to OpenAPI syntax
	openAPIPath := convertEchoPathToOpenAPI(pattern)
	r.OAS.AddOperation(openAPIPath, method, op)

	return nil
}

// AddV1Route is used to add a route to the echo router and OpenAPI schema at the same time ensuring consistency between the spec and the server for version 1 routes of the api
func (r *Router) AddV1Route(pattern, method string, op *openapi3.Operation, route echo.Routable) error {
	grp := r.VersionOne()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	// Convert Echo path syntax to OpenAPI syntax
	openAPIPath := convertEchoPathToOpenAPI(pattern)
	r.OAS.AddOperation(openAPIPath, method, op)

	return nil
}

// AddUnversionedRoute is used to add a versioned route to the echo router and OpenAPI schema at the same time ensuring consistency between the spec and the server
func (r *Router) AddUnversionedRoute(pattern, method string, op *openapi3.Operation, route echo.Routable) error {
	grp := r.Base()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	// Convert Echo path syntax to OpenAPI syntax
	openAPIPath := convertEchoPathToOpenAPI(pattern)
	r.OAS.AddOperation(openAPIPath, method, op)

	return nil
}

// AddEchoOnlyRoute is used to add a route to the echo router without adding it to the OpenAPI schema
func (r *Router) AddEchoOnlyRoute(route echo.Routable) error {
	grp := r.Base()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	return nil
}

// VersionOne returns a new echo group for version 1 of the API
func (r *Router) VersionOne() *echo.Group {
	return r.Echo.Group("v1")
}

// VersionTwo returns a new echo group for version 2 of the API - lets anticipate the future
func (r *Router) VersionTwo() *echo.Group {
	return r.Echo.Group("v2")
}

// Base returns the base echo group - no "version" prefix for the router group
func (r *Router) Base() *echo.Group {
	return r.Echo.Group("")
}

// Config holds the configuration for a route with automatic OpenAPI registration
type Config struct {
	Path          string
	Method        string
	Name          string
	Description   string
	Tags          []string
	OperationID   string
	Security      *openapi3.SecurityRequirements
	Middlewares   []echo.MiddlewareFunc
	Handler       func(echo.Context, *handlers.OpenAPIContext) error
	SimpleHandler func(echo.Context) error // For handlers that don't need OpenAPI context
}

// registrationContext is a special echo.Context implementation used during OpenAPI registration
type registrationContext struct {
	echo.Context
	ctx    context.Context
	method string
}

// newRegistrationContext creates a new registration context with the HTTP method
func newRegistrationContext(method string) *registrationContext {
	// Create a base context with registration marker
	baseCtx := contextx.With(context.Background(), common.RegistrationMarker{})

	return &registrationContext{
		Context: echo.New().NewContext(nil, nil),
		ctx:     baseCtx,
		method:  method,
	}
}

// Request returns a minimal request that won't panic when accessed
func (rc *registrationContext) Request() *http.Request {
	req, _ := http.NewRequestWithContext(rc.ctx, rc.method, "/", nil)

	return req
}

// AddV1HandlerRoute adds a route with automatic OpenAPI context injection
func (r *Router) AddV1HandlerRoute(config Config) error {
	operation := openapi3.NewOperation()
	operation.Summary = config.Name
	operation.Description = config.Description
	operation.Tags = config.Tags
	operation.OperationID = config.OperationID

	if config.Security != nil {
		operation.Security = config.Security
	}

	// Create OpenAPI context
	openAPIContext := &handlers.OpenAPIContext{
		Operation: operation,
		Registry:  r.SchemaRegistry,
	}

	// Call the handler with a registration context to trigger OpenAPI registration
	// This allows handlers to register their request/response schemas at startup
	regCtx := newRegistrationContext(config.Method)

	// Try to call the handler - if it returns an error or panics, that's OK
	// during registration. The important thing is that the handler had a chance
	// to register its schemas via BindAndValidateWithAutoRegistry and response methods
	func() {
		defer func() {
			// During registration, handlers might panic when accessing nil request fields
			// This is expected and OK - the schemas should still be registered
			_ = recover()
		}()

		if config.Handler != nil {
			_ = config.Handler(regCtx, openAPIContext)
		} else if config.SimpleHandler != nil {
			_ = config.SimpleHandler(regCtx)
		}
	}()

	// Ensure common error responses are registered for all endpoints
	if openAPIContext.Operation != nil {
		// Add standard error responses that all endpoints should have
		handlers.AddStandardResponses(openAPIContext.Operation)

		// Register success response schemas based on operation ID patterns
		r.registerSuccessResponseSchemas(config, openAPIContext)

		// Add path parameters from the path pattern if not already added by BindAndValidateWithAutoRegistry
		r.addPathParametersFromPattern(config.Path, operation)
	}

	// Create echo route with automatic OpenAPI context injection
	var routeHandler func(echo.Context) error
	if config.Handler != nil {
		routeHandler = func(c echo.Context) error {
			return config.Handler(c, openAPIContext)
		}
	} else if config.SimpleHandler != nil {
		routeHandler = config.SimpleHandler
	}

	route := echo.Route{
		Name:        config.Name,
		Method:      config.Method,
		Path:        config.Path,
		Middlewares: config.Middlewares,
		Handler:     routeHandler,
	}

	// Add route to echo router
	grp := r.VersionOne()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	// Add operation to OpenAPI schema (convert Echo path syntax to OpenAPI syntax)
	openAPIPath := convertEchoPathToOpenAPI("/v1" + config.Path)
	r.OAS.AddOperation(openAPIPath, config.Method, operation)

	return nil
}

// AddUnversionedHandlerRoute adds an unversioned route with automatic OpenAPI context injection
func (r *Router) AddUnversionedHandlerRoute(config Config) error {
	operation := openapi3.NewOperation()
	operation.Summary = config.Name
	operation.Description = config.Description
	operation.Tags = config.Tags
	operation.OperationID = config.OperationID

	if config.Security != nil {
		operation.Security = config.Security
	}

	// Create OpenAPI context
	openAPIContext := &handlers.OpenAPIContext{
		Operation: operation,
		Registry:  r.SchemaRegistry,
	}

	// Call the handler with a registration context to trigger OpenAPI registration
	regCtx := newRegistrationContext(config.Method)

	func() {
		defer func() {
			// During registration, handlers might panic when accessing nil request fields
			_ = recover()
		}()

		if config.Handler != nil {
			_ = config.Handler(regCtx, openAPIContext)
		} else if config.SimpleHandler != nil {
			_ = config.SimpleHandler(regCtx)
		}
	}()

	// Ensure common error responses are registered for all endpoints
	if openAPIContext.Operation != nil {
		r.registerSuccessResponseSchemas(config, openAPIContext)

		// Add path parameters from the path pattern if not already added by BindAndValidateWithAutoRegistry
		r.addPathParametersFromPattern(config.Path, operation)
	}

	// Create echo route with automatic OpenAPI context injection
	var routeHandler func(echo.Context) error
	if config.Handler != nil {
		routeHandler = func(c echo.Context) error {
			return config.Handler(c, openAPIContext)
		}
	} else if config.SimpleHandler != nil {
		routeHandler = config.SimpleHandler
	}

	route := echo.Route{
		Name:        config.Name,
		Method:      config.Method,
		Path:        config.Path,
		Middlewares: config.Middlewares,
		Handler:     routeHandler,
	}

	// Add route to echo router
	grp := r.Base()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	// Add operation to OpenAPI schema (convert Echo path syntax to OpenAPI syntax)
	openAPIPath := convertEchoPathToOpenAPI(config.Path)
	r.OAS.AddOperation(openAPIPath, config.Method, operation)

	return nil
}

// AddGraphQLToOpenAPI adds the GraphQL endpoint to the OpenAPI specification
func (r *Router) AddGraphQLToOpenAPI() {
	// Create GraphQL request schema
	queryProp := openapi3.NewStringSchema()
	queryProp.Description = "The GraphQL query string"

	variablesProp := openapi3.NewObjectSchema()
	variablesProp.Description = "A JSON object containing variables for the query"

	operationNameProp := openapi3.NewStringSchema()
	operationNameProp.Description = "The name of the operation to execute (optional)"

	requestSchema := openapi3.NewObjectSchema()
	requestSchema.WithProperty("query", queryProp)
	requestSchema.WithProperty("variables", variablesProp)
	requestSchema.WithProperty("operationName", operationNameProp)
	requestSchema.Example = map[string]any{
		"query":     "query GetBooks {\n  books {\n    id\n    title\n  }\n}",
		"variables": map[string]any{},
	}

	// Create GraphQL response schema
	dataProp := openapi3.NewObjectSchema()
	dataProp.Description = "The data returned by the GraphQL operation"

	errorItem := openapi3.NewObjectSchema()
	errorItem.Description = "An array of error objects if the operation failed"

	errorsProp := openapi3.NewArraySchema()
	errorsProp.WithItems(errorItem)

	responseSchema := openapi3.NewObjectSchema()
	responseSchema.WithProperty("data", dataProp)
	responseSchema.WithProperty("errors", errorsProp)

	// Create the GraphQL operation
	operation := openapi3.NewOperation()
	operation.OperationID = "GraphQLQuery"
	operation.Summary = "GraphQL Endpoint"
	operation.Description = "Handles all GraphQL queries, mutations, and subscriptions"
	operation.Tags = []string{"graphql"}

	// Create the GraphQLHistory operation
	operationHistory := openapi3.NewOperation()
	operationHistory.OperationID = "GraphQLQueryHistory"
	operationHistory.Summary = "GraphQL History Endpoint"
	operationHistory.Description = "Handles all GraphQL queries for historical data"
	operationHistory.Tags = []string{"graphql", "history", "audit logs"}

	// Add request body
	requestBody := openapi3.NewRequestBody()
	requestBody.Required = true
	requestBody.Description = "GraphQL query request"
	requestBody.WithJSONSchema(requestSchema)
	operation.RequestBody = &openapi3.RequestBodyRef{Value: requestBody}
	operationHistory.RequestBody = &openapi3.RequestBodyRef{Value: requestBody}

	// Add response
	response := openapi3.NewResponse()
	response.WithDescription("Successful GraphQL response")
	response.WithJSONSchema(responseSchema)
	operation.AddResponse(200, response)        //nolint:mnd
	operationHistory.AddResponse(200, response) //nolint:mnd

	// Add the operation to the OpenAPI spec
	r.OAS.AddOperation("/query", "POST", operation)
	r.OAS.AddOperation("/history/query", "POST", operationHistory)
}

// RegisterRoutes with the echo routers - Router is defined within openapi.go
func RegisterRoutes(router *Router) error {
	// base middleware for all routes that does not included additional middleware
	baseMW = baseMiddleware(router)
	// Middleware for authenticated endpoints
	authMW = authMiddleware(router)
	// Default middleware for other routes which includes additional middleware
	mw = defaultMiddleware(router)

	// routeHandlers that take the router and handler as input
	routeHandlers := []any{
		registerReadinessHandler,
		registerForgotPasswordHandler,
		registerVerifyHandler,
		registerResetPasswordHandler,
		registerResendEmailHandler,
		registerRegisterHandler,
		registerVerifySubscribeHandler,
		registerRefreshHandler,
		registerJwksWellKnownHandler,
		registerInviteHandler,
		registerGithubLoginHandler,
		registerGithubCallbackHandler,
		registerGoogleLoginHandler,
		registerGoogleCallbackHandler,
		registerWebauthnRegistrationHandler,
		registerWebauthnVerificationsHandler,
		registerWebauthnAuthenticationHandler,
		registerWebauthnAuthVerificationHandler,
		registerUserInfoHandler,
		registerOAuthRegisterHandler,
		registerIntegrationOAuthStartHandler,
		registerIntegrationOAuthCallbackHandler,
		registerRefreshIntegrationTokenHandler,
		registerIntegrationProvidersHandler,
		registerIntegrationConfigHandler,
		registerIntegrationOperationHandler,
		registerSwitchRoute,
		registerLivenessHandler,
		registerSecurityTxtHandler,
		registerRobotsHandler,
		registerFaviconHandler,
		registerOpenAPIHandler,
		registerLoginHandler,
		registerAccountAccessHandler,
		registerAccountRolesHandler,
		registerAccountRolesOrganizationHandler,
		registerAccountFeaturesHandler,
		register2faHandler,
		registerExampleCSVHandler,
		registerWebAuthnWellKnownHandler,
		registerAcmeSolverHandler,
		registerCSRFHandler,
		registerWebfingerHandler,
		registerSSOLoginHandler,
		registerSSOCallbackHandler,
		registerSSOTokenAuthorizeHandler,
		registerSSOTokenCallbackHandler,
		registerTrustCenterAnonymousJWTHandler,
		registerQuestionnaireHandler,
		registerQuestionnaireSubmitHandler,
		registerStartImpersonationHandler,
		registerEndImpersonationHandler,
		registerProductCatalogHandler,
		registerFileDownloadHandler,
		registerSCIMRoutes,
		registerResendWebhookHandler,

		// JOB Runners
		// TODO(adelowo): at some point in the future, maybe we should extract these into
		// it's own service/binary
		registerJobRunnerRegistrationHandler,
	}
	// Register the Stripe webhook endpoint only when the entitlements
	// client has been configured. This ensures the server can run without
	// requiring Stripe credentials or webhook support
	if router.Handler != nil && router.Handler.Entitlements != nil {
		routeHandlers = append(routeHandlers, registerWebhookHandler)
	}

	if router.LocalFilePath != "" {
		routeHandlers = append(routeHandlers, registerUploadsHandler)
	}

	for _, route := range routeHandlers {
		if err := route.(func(*Router) error)(router); err != nil {
			return err
		}
	}

	// Add GraphQL endpoint to OpenAPI specification
	router.AddGraphQLToOpenAPI()

	return nil
}

// baseMiddleware returns the base middleware for the router, which includes the transaction middleware
// this isn't used directly in the router register, instead its combined with other middleware functions below
// to include the additional middleware
func baseMiddleware(router *Router) []echo.MiddlewareFunc {
	mw := []echo.MiddlewareFunc{}

	// add transaction middleware
	transactionConfig := transaction.Client{
		EntDBClient: router.Handler.DBClient,
	}

	mimeMiddleware := mime.NewWithConfig(mime.Config{DefaultContentType: httpsling.ContentTypeJSONUTF8})

	return append(mw, mimeMiddleware, transactionConfig.Middleware)
}

// authMiddleware returns the middleware for the router that is used on authenticated routes
// it includes the transaction middleware, the auth middleware, and any additional middleware
// after the auth middleware
func authMiddleware(router *Router) []echo.MiddlewareFunc {
	mw := baseMW
	// add the impersonation middleware to identify and validate impersonated requests
	mw = append(mw, impersonationMiddleware(router))
	// add the auth middleware
	mw = append(mw, router.Handler.AuthMiddleware...)
	// add system admin user context middleware (after auth, so we know if user is system admin)
	mw = append(mw, impersonation.SystemAdminUserContextMiddleware())
	// append any additional middleware after the auth middleware (includes csrf)
	return append(mw, router.Handler.AdditionalMiddleware...)
}

// impersonationMiddleware returns the middleware for the router that is used on
// authenticated routes. it identifies impersonated requests and the user impersonating
// and checks access
func impersonationMiddleware(router *Router) echo.MiddlewareFunc {
	mw := impersonation.New(router.Handler.TokenManager)
	return mw.Process
}

// defaultMiddleware returns the default middleware for the router to be used
// on all unauthenticated + unrestricted routes
func defaultMiddleware(router *Router) []echo.MiddlewareFunc {
	mw := baseMW
	// this is the default middleware that is applied to all routes
	// it includes the transaction middleware and any additional middleware (includes csrf)
	return append(mw, router.Handler.AdditionalMiddleware...)
}
