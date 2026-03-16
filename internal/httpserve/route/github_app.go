package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerGitHubAppInstallHandler registers the GitHub App installation start handler.
func registerGitHubAppInstallHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/github/app/install",
		Method:      http.MethodPost,
		Name:        "GitHubAppInstall",
		Description: "Start GitHub App installation flow",
		Tags:        []string{"integrations"},
		OperationID: "GitHubAppInstall",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.StartGitHubAppInstallation,
	}

	return router.AddV1HandlerRoute(config)
}

// registerGitHubAppCallbackHandler registers the GitHub App installation callback handler.
func registerGitHubAppCallbackHandler(router *Router) error {
	config := Config{
		Path:        "/integrations/github/app/callback",
		Method:      http.MethodGet,
		Name:        "GitHubAppInstallCallback",
		Description: "Handle GitHub App installation callback",
		Tags:        []string{"integrations"},
		OperationID: "GitHubAppInstallCallback",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.GitHubAppInstallCallback,
	}

	return router.AddV1HandlerRoute(config)
}
