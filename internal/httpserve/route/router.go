package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/middleware/cors"
	"github.com/theopenlane/core/pkg/middleware/impersonation"
	"github.com/theopenlane/core/pkg/middleware/mime"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/pkg/middleware/transaction"
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

// pathParamNames returns the parameter names present in an Echo-style path template
func pathParamNames(path string) map[string]bool {
	names := make(map[string]bool)

	for part := range strings.SplitSeq(path, "/") {
		if strings.HasPrefix(part, ":") {
			names[part[1:]] = true
		}
	}

	return names
}

// addPathParametersFromPattern extracts path parameters from Echo-style path and adds them to OpenAPI operation
func (r *Router) addPathParametersFromPattern(path string, operation *openapi3.Operation) {
	// Extract parameter names from Echo-style path (e.g., :id, :name)
	parts := strings.SplitSeq(path, "/")
	for part := range parts {
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
	// Echo is the underlying Echo router instance.
	Echo *echo.Echo
	// OAS is the OpenAPI spec being assembled.
	OAS *openapi3.T
	// Handler provides the HTTP handlers wired into routes.
	Handler *handlers.Handler
	// StartConfig holds Echo start configuration.
	StartConfig *echo.StartConfig
	// LocalFilePath points to static file roots for local assets.
	LocalFilePath string
	// Logger is the Echo logger used by the router.
	Logger *echo.Logger
	// SchemaRegistry registers and resolves OpenAPI schemas.
	SchemaRegistry handlers.SchemaRegistry
	// SpecInstances maps qualified model type names to instances for schema reflection; populated
	// from the generated instances file, only in spec-build mode
	SpecInstances map[string]any
	// SpecTypes accumulates every model type name the analyzed handlers need, used by the instance
	// emitter and to detect missing instances after registration
	SpecTypes map[string]bool
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
	// Path is the route path pattern.
	Path string
	// Method is the HTTP method for the route.
	Method string
	// Name is the OpenAPI summary for the route.
	Name string
	// Description is the OpenAPI description for the route.
	Description string
	// Tags are the OpenAPI tags for grouping.
	Tags []string
	// OperationID is the OpenAPI operation ID.
	OperationID string
	// Security defines OpenAPI security requirements for the route.
	Security *openapi3.SecurityRequirements
	// Middlewares are applied before the handler.
	Middlewares []echo.MiddlewareFunc
	// RateLimit applies a dedicated per-route rate limiter ahead of the route middleware when set and enabled.
	RateLimit *ratelimit.Config
	// Handler is the handler function wired into the route.
	Handler func(echo.Context) error
	// IncludeInOAS publishes this route in the OpenAPI specification. The
	// published surface is opt-in: only routes explicitly intended for
	// external consumers are documented, everything else stays private.
	IncludeInOAS bool
	// PublicCORS serves the route with a wildcard, credential-less CORS policy, for fully public
	// token-authorized endpoints called cross-origin from arbitrary domains (e.g. trust centers);
	// only supported for static paths (no path parameters)
	PublicCORS bool
}

// rateLimitedMiddlewares prepends a dedicated per-route rate limiter to the configured middleware when the route
// declares one that is enabled or in dry-run mode, so the limiter runs ahead of authentication and handler work
func rateLimitedMiddlewares(config Config) []echo.MiddlewareFunc {
	if config.RateLimit == nil {
		return config.Middlewares
	}

	if !config.RateLimit.Enabled && !config.RateLimit.DryRun {
		return config.Middlewares
	}

	return append([]echo.MiddlewareFunc{ratelimit.RateLimiterWithConfig(config.RateLimit)}, config.Middlewares...)
}

// AddV1HandlerRoute adds a route to the v1 group and, in spec-build mode, its derived operation
func (r *Router) AddV1HandlerRoute(config Config) error {
	if err := r.buildOperation(config, "/v1"+config.Path); err != nil {
		return err
	}

	route := echo.Route{
		Name:        config.Name,
		Method:      config.Method,
		Path:        config.Path,
		Middlewares: rateLimitedMiddlewares(config),
		Handler:     config.Handler,
	}

	// Add route to echo router
	grp := r.VersionOne()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	if config.PublicCORS {
		cors.RegisterPublicPath("/v1" + config.Path)
	}

	return nil
}

// buildOperation derives the route's OpenAPI operation from the handler source and adds it to the
// spec; it only runs in spec-build mode (OAS set, which only spec generation does) for published
// routes — at runtime the served spec is the pre-generated embedded artifact
func (r *Router) buildOperation(config Config, specPath string) error {
	if r.OAS == nil || !config.IncludeInOAS {
		return nil
	}

	operation := openapi3.NewOperation()
	operation.Summary = config.Name
	operation.Description = config.Description
	operation.Tags = config.Tags
	operation.OperationID = config.OperationID

	if config.Security != nil {
		operation.Security = config.Security
	}

	analysis, err := handlers.AnalyzeHandler(config.Handler)
	if err != nil {
		return fmt.Errorf("%s: %w", config.OperationID, err)
	}

	if analysis.Request != "" {
		r.SpecTypes[analysis.Request] = true

		if instance, ok := r.SpecInstances[analysis.Request]; ok {
			handlers.RegisterParameters(operation, instance, pathParamNames(config.Path))

			if config.Method != http.MethodGet {
				if err := handlers.RegisterRequestBody(operation, r.SchemaRegistry, instance); err != nil {
					return fmt.Errorf("%s: %w", config.OperationID, err)
				}
			}
		}
	}

	if err := r.registerResponses(config, operation, analysis); err != nil {
		return err
	}

	// Add path parameters from the path pattern if not already derived from the request model
	r.addPathParametersFromPattern(config.Path, operation)

	// Add operation to OpenAPI schema (convert Echo path syntax to OpenAPI syntax)
	r.OAS.AddOperation(convertEchoPathToOpenAPI(specPath), config.Method, operation)

	return nil
}

// registerResponses adds the analyzed responses to the operation: success payloads with schemas,
// redirects, error statuses, and statuses produced by route middleware rather than handler code
// (401 for authenticated routes, 429 for rate limited routes); a 500 is always registered since
// any endpoint can fail internally
func (r *Router) registerResponses(config Config, operation *openapi3.Operation, analysis *handlers.HandlerAnalysis) error {
	statuses := analysis.Responses

	if config.Security != nil && len(*config.Security) > 0 {
		statuses[http.StatusUnauthorized] = handlers.ResponseShape{}
	}

	if config.RateLimit != nil && config.RateLimit.Enabled {
		statuses[http.StatusTooManyRequests] = handlers.ResponseShape{}
	}

	statuses[http.StatusInternalServerError] = handlers.ResponseShape{}

	for status, shape := range statuses {
		switch {
		case status >= http.StatusBadRequest:
			operation.AddResponse(status, handlers.NewErrorResponse(status))
		case status == http.StatusFound:
			operation.AddResponse(status, handlers.RedirectResponse())
		case shape.Type != "":
			r.SpecTypes[shape.Type] = true

			if instance, ok := r.SpecInstances[shape.Type]; ok {
				if err := handlers.RegisterSuccessResponse(operation, r.SchemaRegistry, status, instance); err != nil {
					return fmt.Errorf("%s: %w", config.OperationID, err)
				}
			}
		default:
			operation.AddResponse(status, contentResponse(status, shape.ContentType))
		}
	}

	return nil
}

// contentResponse builds a schema-less response for the given status and media type; payloads the
// analysis could not type get a description-only response
func contentResponse(status int, contentType string) *openapi3.Response {
	response := openapi3.NewResponse().WithDescription(http.StatusText(status))

	switch contentType {
	case "":
		return response
	case "application/json":
		return response.WithContent(openapi3.NewContentWithSchema(openapi3.NewObjectSchema(), []string{contentType}))
	default:
		return response.WithContent(openapi3.NewContentWithSchema(openapi3.NewStringSchema(), []string{contentType}))
	}
}

// AddUnversionedHandlerRoute adds a route to the base group and, in spec-build mode, its derived operation
func (r *Router) AddUnversionedHandlerRoute(config Config) error {
	if err := r.buildOperation(config, config.Path); err != nil {
		return err
	}

	route := echo.Route{
		Name:        config.Name,
		Method:      config.Method,
		Path:        config.Path,
		Middlewares: rateLimitedMiddlewares(config),
		Handler:     config.Handler,
	}

	// Add route to echo router
	grp := r.Base()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	if config.PublicCORS {
		cors.RegisterPublicPath(config.Path)
	}

	return nil
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
		registerUnsubscribeHandler,
		registerRefreshHandler,
		registerLogoutHandler,
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
		registerIntegrationAuthStartHandler,
		registerIntegrationAuthCallbackHandler,
		registerStaticWebhookRoutes,
		registerIntegrationProvidersHandler,
		registerIntegrationConfigHandler,
		registerIntegrationDisconnectHandler,
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
		registerAccountRolesMeHandler,
		registerAccountRolesOrganizationHandler,
		registerAccountFeaturesHandler,
		register2faHandler,
		registerExampleCSVHandler,
		registerWebAuthnWellKnownHandler,
		registerAcmeSolverHandler,
		registerCSRFHandler,
		registerWebfingerHandler,
		registerSSOLoginHandler,
		registerSSOInitiateHandler,
		registerSSOCallbackHandler,
		registerSSOTokenAuthorizeHandler,
		registerSSOTokenCallbackHandler,
		registerTrustCenterAnonymousJWTHandler,
		registerQuestionnaireHandler,
		registerQuestionnaireSubmitHandler,
		registerResendQuestionnaireHandler,
		registerStartImpersonationHandler,
		registerEndImpersonationHandler,
		registerSupportCallbackHandler,
		registerProductCatalogHandler,
		registerFileDownloadHandler,
		registerIntegrationWebhookHandler,
		registerSCIMRoutes,
		registerEmailTestSendHandler,
		registerScopesHandler,
		registerOrganizationRolesHandler,
		registerRolesHandler,
		registerMSFTIdentityWellKnownHandler,

		// JOB Runners
		// TODO(adelowo): at some point in the future, maybe we should extract these into
		// it's own service/binary
	}

	if router.Handler.CloudflareConfig.Enabled {
		routeHandlers = append(routeHandlers, registerCloudflareSnapshotHandler)
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
	mw := impersonation.New(router.Handler.TokenManager, router.Handler.SupportAccessConfig.SubjectID, router.Handler.SupportAccessConfig.DisplayName)
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
