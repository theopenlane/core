package route

import (
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

var (
	mw     = []echo.MiddlewareFunc{middleware.Recover()}
	authMW = []echo.MiddlewareFunc{}

	restrictedRateLimit   = &ratelimit.Config{RateLimit: 10, BurstLimit: 10, ExpiresIn: 15 * time.Minute} //nolint:mnd
	restrictedEndpointsMW = []echo.MiddlewareFunc{}
)

// Router is a struct that holds the echo router, the OpenAPI schema, and the handler - it's a way to group these components together
type Router struct {
	Echo          *echo.Echo
	OAS           *openapi3.T
	Handler       *handlers.Handler
	StartConfig   *echo.StartConfig
	LocalFilePath string
	Logger        *echo.Logger
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

	r.OAS.AddOperation(pattern, method, op)

	return nil
}

// AddV1Route is used to add a route to the echo router and OpenAPI schema at the same time ensuring consistency between the spec and the server for version 1 routes of the api
func (r *Router) AddV1Route(pattern, method string, op *openapi3.Operation, route echo.Routable) error {
	grp := r.VersionOne()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	r.OAS.AddOperation(pattern, method, op)

	return nil
}

// AddUnversionedRoute is used to add a versioned route to the echo router and OpenAPI schema at the same time ensuring consistency between the spec and the server
func (r *Router) AddUnversionedRoute(pattern, method string, op *openapi3.Operation, route echo.Routable) error {
	grp := r.Base()

	_, err := grp.AddRoute(route)
	if err != nil {
		return err
	}

	r.OAS.AddOperation(pattern, method, op)

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

// RegisterRoutes with the echo routers - Router is defined within openapi.go
func RegisterRoutes(router *Router) error {
	// add transaction middleware
	transactionConfig := transaction.Client{
		EntDBClient: router.Handler.DBClient,
	}

	mw = append(mw, transactionConfig.Middleware)

	// Middleware for restricted endpoints
	restrictedEndpointsMW = append(restrictedEndpointsMW, mw...)
	restrictedEndpointsMW = append(restrictedEndpointsMW, ratelimit.RateLimiterWithConfig(restrictedRateLimit)) // add restricted ratelimit middleware

	// Middleware for authenticated endpoints
	authMW = append(authMW, mw...)
	authMW = append(authMW, router.Handler.AuthMiddleware...)

	// routeHandlers that take the router and handler as input
	routeHandlers := []interface{}{
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
		registerAppleMerchantHandler,
		registerWebhookHandler,
		register2faHandler,
		registerExampleCSVHandler,
		registerWebAuthnWellKnownHandler,
		registerAcmeSolverHandler,
		registerTrustCenterAnonymousJWTHandler,

		// JOB Runners
		// TODO(adelowo): at some point in the future, maybe we should extract these into
		// it's own service/binary
		registerJobRunnerRegistrationHandler,
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
