package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerAccountFeaturesHandler registers the /account/features handler
func registerAccountFeaturesHandler(router *Router) (err error) {
	// add route without the path param
	path := "/account/features"
	method := http.MethodGet
	name := "AccountFeatures"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.AccountFeaturesHandler(c)
		},
	}

	featuresOrganizationOperation := router.Handler.BindAccountFeatures()

	if err := router.AddV1Route(route.Path, route.Method, featuresOrganizationOperation, route); err != nil {
		return err
	}

	// add an additional route with the path param
	route.Path = "/account/features/:{id}"
	route.Name = name + "ByID"

	rolesOrganizationOperationByID := router.Handler.BindAccountFeaturesByID()

	if err := router.AddV1Route(route.Path, route.Method, rolesOrganizationOperationByID, route); err != nil {
		return err
	}

	return nil
}
