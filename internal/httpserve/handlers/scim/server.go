package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	opt "github.com/elimity-com/scim/optional"
	schema "github.com/elimity-com/scim/schema"
)

const (
	// maxSCIMResults is the maximum number of results returned in a SCIM list response
	maxSCIMResults = 200
)

// NewSCIMServer creates a new SCIM server with directory-backed User and Group resource handlers.
func NewSCIMServer() (scim.Server, error) {
	config := scim.ServiceProviderConfig{
		DocumentationURI: opt.NewString("https://docs.theopenlane.io/docs/platform/integrations/scim"),
		AuthenticationSchemes: []scim.AuthenticationScheme{
			{
				Type:        scim.AuthenticationTypeOauthBearerToken,
				Name:        "Bearer API Token",
				Description: "Authenticate with an Openlane API token using the Authorization: Bearer <token> header",
			},
		},
		SupportPatch:     true,
		SupportFiltering: true,
		MaxResults:       maxSCIMResults,
	}

	userResourceType := scim.ResourceType{
		ID:          opt.NewString("User"),
		Name:        "User",
		Description: opt.NewString("User account"),
		Endpoint:    "/Users",
		Schema:      schema.CoreUserSchema(),
		Handler:     NewDirectoryUserHandler(),
	}

	groupResourceType := scim.ResourceType{
		ID:          opt.NewString("Group"),
		Name:        "Group",
		Description: opt.NewString("Group resource"),
		Endpoint:    "/Groups",
		Schema:      schema.CoreGroupSchema(),
		Handler:     NewDirectoryGroupHandler(),
	}

	server, err := scim.NewServer(
		&scim.ServerArgs{
			ServiceProviderConfig: &config,
			ResourceTypes: []scim.ResourceType{
				userResourceType,
				groupResourceType,
			},
		},
	)
	if err != nil {
		return scim.Server{}, err
	}

	return server, nil
}

// WrapSCIMServerHTTPHandler wraps the SCIM server's HTTP handler with context preservation
// This ensures that request context (auth, transaction, etc.) flows through to handlers
func WrapSCIMServerHTTPHandler(server scim.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	}
}
