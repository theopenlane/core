package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerMSFTIdentityWellKnownHandler serves up the msft identity well-known endpoint
// in order to register an application with your domain in Azure, you must host a well-known file at
// https://<domain>/.well-known/microsoft-identity-association.json
func registerMSFTIdentityWellKnownHandler(router *Router) (err error) {
	config := Config{
		Path:          "/.well-known/microsoft-identity-association.json",
		Method:        http.MethodGet,
		Name:          "MSFTIdentityWellKnown",
		Description:   "Microsoft Identity Association well-known configuration file",
		Tags:          []string{"well-known", "msft", "azure"},
		OperationID:   "MSFTIdentityWellKnown",
		Security:      handlers.PublicSecurity,
		Middlewares:   *publicEndpoint,
		SimpleHandler: router.Handler.MSFTIdentityWellKnownHandler,
	}

	return router.AddUnversionedHandlerRoute(config)
}
