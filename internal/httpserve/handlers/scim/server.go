package scim

import (
	"net/http"

	"github.com/elimity-com/scim"
	opt "github.com/elimity-com/scim/optional"
	schema "github.com/elimity-com/scim/schema"

	"github.com/theopenlane/core/common/enums"
)

const (
	// maxSCIMResults is the maximum number of results returned in a SCIM list response
	maxSCIMResults = 200
)

// NewSCIMServer creates a new SCIM server with User and Group resource handlers
func NewSCIMServer() (scim.Server, error) {
	config := scim.ServiceProviderConfig{
		DocumentationURI: opt.NewString("https://docs.theopenlane.io/scim"),
		AuthenticationSchemes: []scim.AuthenticationScheme{
			{
				Type:        scim.AuthenticationTypeOauthBearerToken,
				Name:        "OAuth Bearer Token",
				Description: "Authentication scheme using the OAuth Bearer Token standard",
				SpecURI:     opt.NewString("https://www.rfc-editor.org/rfc/rfc6750.html"),
				Primary:     true,
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
		Handler: &ModeDispatchUserHandler{
			users:     NewUserHandler(),
			directory: NewDirectoryUserHandler(),
		},
	}

	groupResourceType := scim.ResourceType{
		ID:          opt.NewString("Group"),
		Name:        "Group",
		Description: opt.NewString("Group resource"),
		Endpoint:    "/Groups",
		Schema:      schema.CoreGroupSchema(),
		Handler: &ModeDispatchGroupHandler{
			groups:    NewGroupHandler(),
			directory: NewDirectoryGroupHandler(),
		},
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

// ModeDispatchUserHandler dispatches SCIM user operations based on provision mode
type ModeDispatchUserHandler struct {
	// users is the handler for USERS provision mode (creates real User entities)
	users *UserHandler
	// directory is the handler for DIRECTORY provision mode (creates DirectoryAccount records)
	directory *DirectoryUserHandler
}

// Create dispatches a SCIM user create to the appropriate handler(s) based on provision mode
func (d *ModeDispatchUserHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Create(r, attributes)
	case enums.SCIMProvisionModeBoth:
		res, err := d.users.Create(r, attributes)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Create(r, attributes); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.users.Create(r, attributes)
	}
}

// Get dispatches a SCIM user get to the appropriate handler based on provision mode
func (d *ModeDispatchUserHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	mode := ProvisionModeFromContext(r.Context())

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Get(r, id)
	default:
		return d.users.Get(r, id)
	}
}

// GetAll dispatches a SCIM user list to the appropriate handler based on provision mode
func (d *ModeDispatchUserHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	mode := ProvisionModeFromContext(r.Context())

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.GetAll(r, params)
	default:
		return d.users.GetAll(r, params)
	}
}

// Replace dispatches a SCIM user replace to the appropriate handler(s) based on provision mode
func (d *ModeDispatchUserHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Replace(r, id, attributes)
	case enums.SCIMProvisionModeBoth:
		res, err := d.users.Replace(r, id, attributes)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Replace(r, id, attributes); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.users.Replace(r, id, attributes)
	}
}

// Patch dispatches a SCIM user patch to the appropriate handler(s) based on provision mode
func (d *ModeDispatchUserHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Patch(r, id, operations)
	case enums.SCIMProvisionModeBoth:
		res, err := d.users.Patch(r, id, operations)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Patch(r, id, operations); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.users.Patch(r, id, operations)
	}
}

// Delete dispatches a SCIM user delete to the appropriate handler(s) based on provision mode
func (d *ModeDispatchUserHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Delete(r, id)
	case enums.SCIMProvisionModeBoth:
		if err := d.users.Delete(r, id); err != nil {
			return err
		}

		return d.directory.Delete(r, id)
	default:
		return d.users.Delete(r, id)
	}
}

// ModeDispatchGroupHandler dispatches SCIM group operations based on provision mode
type ModeDispatchGroupHandler struct {
	// groups is the handler for USERS provision mode (creates real Group entities)
	groups *GroupHandler
	// directory is the handler for DIRECTORY provision mode (creates DirectoryGroup records)
	directory *DirectoryGroupHandler
}

// Create dispatches a SCIM group create to the appropriate handler(s) based on provision mode
func (d *ModeDispatchGroupHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Create(r, attributes)
	case enums.SCIMProvisionModeBoth:
		res, err := d.groups.Create(r, attributes)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Create(r, attributes); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.groups.Create(r, attributes)
	}
}

// Get dispatches a SCIM group get to the appropriate handler based on provision mode
func (d *ModeDispatchGroupHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	mode := ProvisionModeFromContext(r.Context())

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Get(r, id)
	default:
		return d.groups.Get(r, id)
	}
}

// GetAll dispatches a SCIM group list to the appropriate handler based on provision mode
func (d *ModeDispatchGroupHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	mode := ProvisionModeFromContext(r.Context())

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.GetAll(r, params)
	default:
		return d.groups.GetAll(r, params)
	}
}

// Replace dispatches a SCIM group replace to the appropriate handler(s) based on provision mode
func (d *ModeDispatchGroupHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Replace(r, id, attributes)
	case enums.SCIMProvisionModeBoth:
		res, err := d.groups.Replace(r, id, attributes)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Replace(r, id, attributes); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.groups.Replace(r, id, attributes)
	}
}

// Patch dispatches a SCIM group patch to the appropriate handler(s) based on provision mode
func (d *ModeDispatchGroupHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Patch(r, id, operations)
	case enums.SCIMProvisionModeBoth:
		res, err := d.groups.Patch(r, id, operations)
		if err != nil {
			return scim.Resource{}, err
		}

		if _, err := d.directory.Patch(r, id, operations); err != nil {
			return scim.Resource{}, err
		}

		return res, nil
	default:
		return d.groups.Patch(r, id, operations)
	}
}

// Delete dispatches a SCIM group delete to the appropriate handler(s) based on provision mode
func (d *ModeDispatchGroupHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	mode := ProvisionModeFromContext(ctx)

	switch mode {
	case enums.SCIMProvisionModeDirectory:
		return d.directory.Delete(r, id)
	case enums.SCIMProvisionModeBoth:
		if err := d.groups.Delete(r, id); err != nil {
			return err
		}

		return d.directory.Delete(r, id)
	default:
		return d.groups.Delete(r, id)
	}
}

