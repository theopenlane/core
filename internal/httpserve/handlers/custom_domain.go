package handlers

import (
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
)

// CreateCustomDomainHandler creates custom (aka vanity) domains
func (h *Handler) CreateCustomDomainHandler(ctx echo.Context) error {
	return h.InternalServerError(ctx, errors.New("Unimplemented"))
}

// BindCreateCustomDomain
func (h *Handler) BindCreateCustomDomain() *openapi3.Operation {
	orgRoles := openapi3.NewOperation()
	orgRoles.Description = "Creates a custom domain"
	orgRoles.Tags = []string{"domain"}
	orgRoles.OperationID = "CreateCustomDomain"
	orgRoles.Security = AllSecurityRequirements()

	orgRoles.AddResponse(http.StatusInternalServerError, internalServerError())
	orgRoles.AddResponse(http.StatusBadRequest, badRequest())
	orgRoles.AddResponse(http.StatusBadRequest, invalidInput())

	return orgRoles
}

// DeleteCustomDomainByIDHandler creates custom (aka vanity) domains
func (h *Handler) DeleteCustomDomainByIDHandler(ctx echo.Context) error {
	return h.InternalServerError(ctx, errors.New("Unimplemented"))
}

// BindCreateCustomDomain
func (h *Handler) BindDeleteCustomDomainByID() *openapi3.Operation {
	orgRoles := openapi3.NewOperation()
	orgRoles.Description = "Deletes a custom domain by ID"
	orgRoles.Tags = []string{"domain"}
	orgRoles.OperationID = "DeleteCustomDomainByID"
	orgRoles.Security = AllSecurityRequirements()

	orgRoles.AddResponse(http.StatusInternalServerError, internalServerError())
	orgRoles.AddResponse(http.StatusBadRequest, badRequest())
	orgRoles.AddResponse(http.StatusBadRequest, invalidInput())

	return orgRoles
}

// DeleteCustomDomainByIDHandler creates custom (aka vanity) domains
func (h *Handler) GetCustomDomainStatusByIDHandler(ctx echo.Context) error {
	return h.InternalServerError(ctx, errors.New("Unimplemented"))
}

// BindCreateCustomDomain
func (h *Handler) BindGetCustomDomainStatusByID() *openapi3.Operation {
	orgRoles := openapi3.NewOperation()
	orgRoles.Description = "Gets a custom domain's status"
	orgRoles.Tags = []string{"domain"}
	orgRoles.OperationID = "GetCustomDomainStatusByID"
	orgRoles.Security = AllSecurityRequirements()

	orgRoles.AddResponse(http.StatusInternalServerError, internalServerError())
	orgRoles.AddResponse(http.StatusBadRequest, badRequest())
	orgRoles.AddResponse(http.StatusBadRequest, invalidInput())

	return orgRoles
}
