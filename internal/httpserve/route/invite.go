package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerInviteHandler registers the invite handler
func registerInviteHandler(router *Router) error {
	config := Config{
		Path:        "/invite",
		Method:      http.MethodGet,
		Name:        "OrganizationInviteAccept",
		Description: "Accept an organization invitation",
		Tags:        []string{"organization"},
		OperationID: "OrganizationInviteAccept",
		Security:    handlers.AllSecurityRequirements(),
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.OrganizationInviteAccept,
	}

	return router.AddV1HandlerRoute(config)
}
