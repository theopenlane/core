package route

import (
	"fmt"
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerTrustCenterAnonymousJWTHandler registers the trust center anonymous JWT handler and route
func registerTrustCenterAnonymousJWTHandler(router *Router) (err error) {
	path := "/trustcenter/auth/anonymous"
	method := http.MethodPost
	name := "TrustCenterAnonymousJWT"

	route := echo.Route{
		Name:   name,
		Method: method,
		Path:   path,
		Handler: func(c echo.Context) error {
			fmt.Println("WTF")
			return router.Handler.CreateTrustCenterAnonymousJWT(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}
