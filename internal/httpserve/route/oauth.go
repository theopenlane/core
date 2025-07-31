package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/httpsling"
)

// registerOAuthRegisterHandler registers the oauth register handler used by the UI to register
// users logging in with an oauth provider
func registerOAuthRegisterHandler(router *Router) error {
	config := Config{
		Path:        "/oauth/register",
		Method:      http.MethodPost,
		Name:        "OAuthRegister",
		Description: "Register a user via OAuth provider authentication",
		Tags:        []string{"oauth", "authentication"},
		OperationID: "OAuthRegister",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		Handler:     router.Handler.OauthRegister,
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerUserInfoHandler registers the userinfo handler
func registerUserInfoHandler(router *Router) error {
	config := Config{
		Path:        "/oauth/userinfo",
		Method:      http.MethodGet,
		Name:        "UserInfo",
		Description: "Get user information for OAuth authenticated user",
		Tags:        []string{"oauth", "user"},
		OperationID: "UserInfo",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *AuthenticatedEndpoint,
		Handler: func(ctx echo.Context, openapi *handlers.OpenAPIContext) error {
			ctx.Response().Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
			return router.Handler.UserInfo(ctx, openapi)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerGithubLoginHandler registers the github login handler
func registerGithubLoginHandler(router *Router) error {
	config := Config{
		Path:        "/github/login",
		Method:      http.MethodGet,
		Name:        "GitHubLogin",
		Description: "Initiate GitHub OAuth login flow",
		Tags:        []string{"oauth", "github"},
		OperationID: "GitHubLogin",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return githubLogin(router)(ctx)
		},
	}

	return router.AddV1HandlerRoute(config)
}

// registerGithubCallbackHandler registers the github callback handler
func registerGithubCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/github/callback",
		Method:      http.MethodGet,
		Name:        "GitHubCallback",
		Description: "Handle GitHub OAuth callback",
		Tags:        []string{"oauth", "github"},
		OperationID: "GitHubCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return githubCallback(router)(ctx)
		},
	}

	return router.AddV1HandlerRoute(config)
}

// registerGoogleLoginHandler registers the google login handler
func registerGoogleLoginHandler(router *Router) error {
	config := Config{
		Path:        "/google/login",
		Method:      http.MethodGet,
		Name:        "GoogleLogin",
		Description: "Initiate Google OAuth login flow",
		Tags:        []string{"oauth", "google"},
		OperationID: "GoogleLogin",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return googleLogin(router)(ctx)
		},
	}

	return router.AddV1HandlerRoute(config)
}

// registerGoogleCallbackHandler registers the google callback handler
func registerGoogleCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/google/callback",
		Method:      http.MethodGet,
		Name:        "GoogleCallback",
		Description: "Handle Google OAuth callback",
		Tags:        []string{"oauth", "google"},
		OperationID: "GoogleCallback",
		Security:    handlers.PublicSecurity,
		Middlewares: *PublicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return googleCallback(router)(ctx)
		},
	}

	return router.AddV1HandlerRoute(config)
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
