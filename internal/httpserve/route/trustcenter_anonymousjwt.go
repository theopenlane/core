package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerTrustCenterAnonymousJWTHandler registers the trust center anonymous JWT handler and route
func registerTrustCenterAnonymousJWTHandler(router *Router) error {
	config := Config{
		Path:        "/trustcenter/auth/anonymous",
		Method:      http.MethodPost,
		Name:        "TrustCenterAnonymousJWT",
		Description: "Create anonymous JWT token for trust center access",
		Tags:        []string{"trustcenter", "authentication"},
		OperationID: "TrustCenterAnonymousJWT",
		Security:    handlers.BasicSecurity(),
		Middlewares: *unauthenticatedEndpoint,
		Handler:     router.Handler.CreateTrustCenterAnonymousJWT,
	}

	return router.AddUnversionedHandlerRoute(config)
}
