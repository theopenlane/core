package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerCreateDomainHandler registers the /domain handler
func registerCustomDomainHandler(router *Router) (err error) {
	createRoute := echo.Route{
		Name:        "CreateCustomDomain",
		Method:      http.MethodPost,
		Path:        "/domain/custom",
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.CreateCustomDomainHandler(c)
		},
	}
	deleteRoute := echo.Route{
		Name:        "DeleteCustomDomainByID",
		Method:      http.MethodDelete,
		Path:        "/domain/custom/:{id}",
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.DeleteCustomDomainByIDHandler(c)
		},
	}
	getStatusRoute := echo.Route{
		Name:        "GetCustomDomainStatusByID",
		Method:      http.MethodGet,
		Path:        "/domain/custom/:{id}/status",
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.GetCustomDomainStatusByIDHandler(c)
		},
	}

	createCustomDomainOperation := router.Handler.BindCreateCustomDomain()
	deleteCustomDomainByIDOperation := router.Handler.BindDeleteCustomDomainByID()
	getCustomDomainStatusByIDOperation := router.Handler.BindGetCustomDomainStatusByID()

	if err := router.Addv1Route(createRoute.Path, createRoute.Method, createCustomDomainOperation, createRoute); err != nil {
		return err
	}

	if err := router.Addv1Route(deleteRoute.Path, deleteRoute.Method, deleteCustomDomainByIDOperation, deleteRoute); err != nil {
		return err
	}

	if err := router.Addv1Route(getStatusRoute.Path, getStatusRoute.Method, getCustomDomainStatusByIDOperation, getStatusRoute); err != nil {
		return err
	}

	return nil
}
