package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerTrustCenterAnonymousJWTHandler registers the trust center anonymous JWT handler and route
func registerTrustCenterAnonymousJWTHandler(router *Router) (err error) {
	path := "/trustcenter/anonymous-jwt"
	method := http.MethodPost
	name := "TrustCenterAnonymousJWT"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.CreateTrustCenterAnonymousJWT(c)
		},
	}

	trustCenterJWTOperation := router.Handler.BindCreateTrustCenterAnonymousJWT()

	if err := router.AddV1Route(path, method, trustCenterJWTOperation, route); err != nil {
		return err
	}

	return nil
}
