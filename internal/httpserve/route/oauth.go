package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/httpsling"
)

// registerOAuthRegisterHandler registers the oauth register handler used by the UI to register
// users logging in with an oauth provider
func registerOAuthRegisterHandler(router *Router) (err error) {
	path := "/oauth/register"
	method := http.MethodPost
	name := "OAuthRegister"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler: func(c echo.Context) error {
			return router.Handler.OauthRegister(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerUserInfoHandler registers the userinfo handler
func registerUserInfoHandler(router *Router) (err error) {
	path := "/oauth/userinfo"
	method := http.MethodGet
	name := "UserInfo"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			c.Response().Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			return router.Handler.UserInfo(c)
		},
	}

	if err := router.AddUnversionedRoute(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerGithubLoginHandler registers the github login handler
func registerGithubLoginHandler(router *Router) (err error) {
	path := "/github/login"
	method := http.MethodGet
	name := "GitHubLogin"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     githubLogin(router),
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerGithubCallbackHandler registers the github callback handler
func registerGithubCallbackHandler(router *Router) (err error) {
	path := "/github/callback"
	method := http.MethodGet
	name := "GitHubCallback"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     githubCallback(router),
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerGoogleLoginHandler registers the google login handler
func registerGoogleLoginHandler(router *Router) (err error) {
	path := "/google/login"
	method := http.MethodGet
	name := "GoogleLogin"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     googleLogin(router),
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// registerGoogleCallbackHandler registers the google callback handler
func registerGoogleCallbackHandler(router *Router) (err error) {
	path := "/google/callback"
	method := http.MethodGet
	name := "GoogleCallback"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: mw,
		Handler:     googleCallback(router),
	}

	if err := router.AddV1Route(path, method, nil, route); err != nil {
		return err
	}

	return nil
}

// githubLogin wraps getloginhandlers
func githubLogin(h *Router) echo.HandlerFunc {
	login, _ := h.Handler.GetGitHubLoginHandlers()

	return echo.WrapHandler(login)
}

// googleLogin wraps getloginhandlers
func googleLogin(h *Router) echo.HandlerFunc {
	login, _ := h.Handler.GetGoogleLoginHandlers()

	return echo.WrapHandler(login)
}

// githubCallback wraps getloginhandlers
func githubCallback(h *Router) echo.HandlerFunc {
	_, callback := h.Handler.GetGitHubLoginHandlers()

	return echo.WrapHandler(callback)
}

// googleCallback wraps getloginhandlers
func googleCallback(h *Router) echo.HandlerFunc {
	_, callback := h.Handler.GetGoogleLoginHandlers()

	return echo.WrapHandler(callback)
}
