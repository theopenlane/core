package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerAccountFeaturesHandler registers the /account/features handler
func registerAccountFeaturesHandler(router *Router) error {
	// add route without the path param
	config := Config{
		Path:        "/account/features",
		Method:      http.MethodGet,
		Name:        "AccountFeatures",
		Description: "List features a subject has in relation to the authenticated organization",
		Tags:        []string{"account"},
		OperationID: "AccountFeatures",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.AccountFeaturesHandler,
	}

	if err := router.AddV1HandlerRoute(config); err != nil {
		return err
	}

	// add an additional route with the path param
	configByID := Config{
		Path:        "/account/features/:id",
		Method:      http.MethodGet,
		Name:        "AccountFeaturesByID",
		Description: "List the features a subject has in relation to the organization ID provided",
		Tags:        []string{"account"},
		OperationID: "AccountFeaturesByID",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.AccountFeaturesHandler,
	}

	return router.AddV1HandlerRoute(configByID)
}
