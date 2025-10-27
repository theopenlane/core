package scim

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"
	scimschema "github.com/elimity-com/scim/schema"
	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/entx"
)

// UserHandler implements scim.ResourceHandler for User resources.
type UserHandler struct{}

// NewUserHandler creates a new UserHandler.
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Create stores given attributes and returns a resource with the attributes that are stored and a unique identifier.
func (h *UserHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	userName, _ := attributes["userName"].(string)
	if userName == "" {
		return scim.Resource{}, fmt.Errorf("%w: userName is required", ErrInvalidAttributes)
	}

	email := userName

	var firstName, lastName, displayName string
	if nameMap, ok := attributes["name"].(map[string]interface{}); ok {
		firstName, _ = nameMap["givenName"].(string)
		lastName, _ = nameMap["familyName"].(string)
	}

	if dn, ok := attributes["displayName"].(string); ok {
		displayName = dn
	}

	if displayName == "" {
		displayName = strings.TrimSpace(firstName + " " + lastName)
	}

	if displayName == "" {
		displayName = email
	}

	active := true
	if activeVal, ok := attributes["active"].(bool); ok {
		active = activeVal
	}

	userSetting, err := client.UserSetting.Create().
		SetEmailConfirmed(true).
		Save(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to create user settings: %w", err)
	}

	lastLoginProvider := enums.AuthProviderOIDC
	authProvider := enums.AuthProviderOIDC
	userRole := enums.RoleUser
	input := generated.CreateUserInput{
		Email:             email,
		FirstName:         &firstName,
		LastName:          &lastName,
		DisplayName:       displayName,
		LastSeen:          lo.ToPtr(time.Now().UTC()),
		LastLoginProvider: &lastLoginProvider,
		AuthProvider:      &authProvider,
		Role:              &userRole,
		ScimUsername:      &email,
		ScimActive:        &active,
		SettingID:         userSetting.ID,
	}

	entUser, err := client.User.Create().
		SetInput(input).
		Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return scim.Resource{}, scimerrors.ScimError{
				ScimType: scimerrors.ScimTypeUniqueness,
				Detail:   fmt.Sprintf("User with email %s already exists", email),
				Status:   http.StatusConflict,
			}
		}

		if generated.IsValidationError(err) {
			return scim.Resource{}, scimerrors.ScimError{
				ScimType: scimerrors.ScimTypeInvalidValue,
				Detail:   fmt.Sprintf("Invalid user attributes: %v", err),
				Status:   http.StatusBadRequest,
			}
		}

		return scim.Resource{}, fmt.Errorf("failed to create user: %w", err)
	}

	if _, err := client.OrgMembership.Create().
		SetUserID(entUser.ID).
		SetOrganizationID(orgID).
		SetRole(enums.RoleMember).
		Save(ctx); err != nil {
		return scim.Resource{}, fmt.Errorf("failed to create org membership: %w", err)
	}

	if !active {
		updatedUser, err := client.User.UpdateOne(entUser).
			SetDeletedAt(time.Now()).
			SetScimActive(false).
			Save(ctx)
		if err != nil {
			return scim.Resource{}, fmt.Errorf("failed to set user inactive: %w", err)
		}

		entUser = updatedUser
	}

	return h.toSCIMResource(ctx, entUser, orgID)
}

// Get returns the resource corresponding with the given identifier.
func (h *UserHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	entUser, err := client.User.Query().
		Where(
			user.ID(id),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)),
		).
		WithGroups(func(gq *generated.GroupQuery) {
			gq.Where(group.HasOwnerWith(organization.ID(orgID)))
		}).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get user: %w", err)
	}

	return h.toSCIMResource(ctx, entUser, orgID)
}

// GetAll returns a paginated list of resources.
func (h *UserHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	query := client.User.Query().
		Where(user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to count users: %w", err)
	}

	offset := params.StartIndex - 1
	if offset < 0 {
		offset = 0
	}

	count := params.Count
	if count <= 0 {
		count = 100
	}

	users, err := query.
		Offset(offset).
		Limit(count).
		WithGroups(func(gq *generated.GroupQuery) {
			gq.Where(group.HasOwnerWith(organization.ID(orgID)))
		}).
		All(ctx)
	if err != nil {
		return scim.Page{}, fmt.Errorf("failed to list users: %w", err)
	}

	resources := make([]scim.Resource, 0, len(users))

	for _, u := range users {
		resource, err := h.toSCIMResource(ctx, u, orgID)
		if err != nil {
			return scim.Page{}, err
		}

		resources = append(resources, resource)
	}

	return scim.Page{
		TotalResults: total,
		Resources:    resources,
	}, nil
}

// Replace replaces ALL existing attributes of the resource with given identifier.
func (h *UserHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	entUser, err := client.User.Query().
		Where(
			user.ID(id),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)),
		).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get user: %w", err)
	}

	userName, _ := attributes["userName"].(string)
	if userName == "" {
		return scim.Resource{}, fmt.Errorf("%w: userName is required", ErrInvalidAttributes)
	}

	email := userName

	var firstName, lastName, displayName string
	if nameMap, ok := attributes["name"].(map[string]interface{}); ok {
		firstName, _ = nameMap["givenName"].(string)
		lastName, _ = nameMap["familyName"].(string)
	}

	if dn, ok := attributes["displayName"].(string); ok {
		displayName = dn
	}

	if displayName == "" {
		displayName = strings.TrimSpace(firstName + " " + lastName)
	}

	if displayName == "" {
		displayName = email
	}

	active := true
	if activeVal, ok := attributes["active"].(bool); ok {
		active = activeVal
	}

	authProvider := enums.AuthProviderOIDC
	update := client.User.UpdateOne(entUser).
		SetEmail(email).
		SetDisplayName(displayName).
		SetLastLoginProvider(enums.AuthProviderOIDC).
		SetAuthProvider(authProvider).
		SetScimUsername(email).
		SetScimActive(active)

	if firstName != "" {
		update.SetFirstName(firstName)
	} else {
		update.ClearFirstName()
	}

	if lastName != "" {
		update.SetLastName(lastName)
	} else {
		update.ClearLastName()
	}

	if !active {
		update.SetDeletedAt(time.Now()).
			SetScimActive(false)
	} else {
		update.ClearDeletedAt().
			SetScimActive(true)
	}

	updatedUser, err := update.Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return scim.Resource{}, scimerrors.ScimError{
				ScimType: scimerrors.ScimTypeUniqueness,
				Detail:   fmt.Sprintf("User with email %s already exists", email),
				Status:   http.StatusConflict,
			}
		}

		if generated.IsValidationError(err) {
			return scim.Resource{}, scimerrors.ScimError{
				ScimType: scimerrors.ScimTypeInvalidValue,
				Detail:   fmt.Sprintf("Invalid user attributes: %v", err),
				Status:   http.StatusBadRequest,
			}
		}

		return scim.Resource{}, fmt.Errorf("failed to update user: %w", err)
	}

	return h.toSCIMResource(ctx, updatedUser, orgID)
}

// Delete removes the resource with corresponding ID.
func (h *UserHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	entUser, err := client.User.Query().
		Where(
			user.ID(id),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)),
		).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scimerrors.ScimErrorResourceNotFound(id)
		}

		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := client.User.UpdateOne(entUser).SetDeletedAt(time.Now()).Exec(ctx); err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}

// Patch updates one or more attributes of a SCIM resource using a sequence of operations.
func (h *UserHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, entx.SoftDeleteSkipKey{}, true)
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("%w: %w", ErrOrgNotFound, err)
	}

	entUser, err := client.User.Query().
		Where(
			user.ID(id),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)),
		).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get user: %w", err)
	}

	update := client.User.UpdateOne(entUser)
	modified := false

	for _, op := range operations {
		switch strings.ToLower(op.Op) {
		case scim.PatchOperationReplace, scim.PatchOperationAdd:
			if err := h.applyPatchOperation(update, op, &modified); err != nil {
				return scim.Resource{}, err
			}
		case scim.PatchOperationRemove:
			if err := h.applyRemoveOperation(update, op, &modified); err != nil {
				return scim.Resource{}, err
			}
		}
	}

	if modified {
		entUser, err = update.Save(ctx)
		if err != nil {
			if generated.IsConstraintError(err) {
				return scim.Resource{}, scimerrors.ScimError{
					ScimType: scimerrors.ScimTypeUniqueness,
					Detail:   "Constraint violation during patch operation",
					Status:   http.StatusConflict,
				}
			}

			if generated.IsValidationError(err) {
				return scim.Resource{}, scimerrors.ScimError{
					ScimType: scimerrors.ScimTypeInvalidValue,
					Detail:   fmt.Sprintf("Invalid user attributes: %v", err),
					Status:   http.StatusBadRequest,
				}
			}

			return scim.Resource{}, fmt.Errorf("failed to patch user: %w", err)
		}
	}

	return h.toSCIMResource(ctx, entUser, orgID)
}

func (h *UserHandler) applyPatchOperation(update *generated.UserUpdateOne, op scim.PatchOperation, modified *bool) error {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	valueMap, isMap := op.Value.(map[string]interface{})
	if !isMap && pathStr == "" {
		return fmt.Errorf("%w: patch operation requires path or value map", ErrInvalidAttributes)
	}

	if isMap {
		if userName, ok := valueMap["userName"].(string); ok && userName != "" {
			update.SetEmail(userName)
			*modified = true
		}

		if displayName, ok := valueMap["displayName"].(string); ok {
			update.SetDisplayName(displayName)
			*modified = true
		}

		if name, ok := valueMap["name"].(map[string]interface{}); ok {
			if givenName, ok := name["givenName"].(string); ok {
				update.SetFirstName(givenName)
				*modified = true
			}

			if familyName, ok := name["familyName"].(string); ok {
				update.SetLastName(familyName)
				*modified = true
			}
		}

		if active, ok := valueMap["active"].(bool); ok {
			if !active {
				update.SetDeletedAt(time.Now()).
					SetScimActive(false)
			} else {
				update.ClearDeletedAt().
					SetScimActive(true)
			}

			*modified = true
		}
	} else {
		switch pathStr {
		case "username":
			if strVal, ok := op.Value.(string); ok {
				update.SetEmail(strVal)
				*modified = true
			}
		case "displayname":
			if strVal, ok := op.Value.(string); ok {
				update.SetDisplayName(strVal)
				*modified = true
			}
		case "name.givenname":
			if strVal, ok := op.Value.(string); ok {
				update.SetFirstName(strVal)
				*modified = true
			}
		case "name.familyname":
			if strVal, ok := op.Value.(string); ok {
				update.SetLastName(strVal)
				*modified = true
			}
		case "active":
			if boolVal, ok := op.Value.(bool); ok {
				if !boolVal {
					update.SetDeletedAt(time.Now()).
						SetScimActive(false)
				} else {
					update.ClearDeletedAt().
						SetScimActive(true)
				}

				*modified = true
			}
		}
	}

	return nil
}

func (h *UserHandler) applyRemoveOperation(update *generated.UserUpdateOne, op scim.PatchOperation, modified *bool) error {
	if op.Path == nil {
		return fmt.Errorf("%w: remove operation requires path", ErrInvalidAttributes)
	}

	pathStr := strings.ToLower(op.Path.String())

	switch pathStr {
	case "name.givenname":
		update.ClearFirstName()
		*modified = true
	case "name.familyname":
		update.ClearLastName()
		*modified = true
	}

	return nil
}

func (h *UserHandler) toSCIMResource(_ any, entUser *generated.User, _ string) (scim.Resource, error) {
	firstName := entUser.FirstName

	lastName := entUser.LastName

	active := entUser.DeletedAt.IsZero()

	groups := make([]map[string]interface{}, 0)
	if entUser.Edges.Groups != nil {
		groups = lo.Map(entUser.Edges.Groups, func(g *generated.Group, _ int) map[string]interface{} {
			return map[string]interface{}{
				"value":   g.ID,
				"display": g.DisplayName,
				"$ref":    fmt.Sprintf("/v1/scim/Groups/%s", g.ID),
			}
		})
	}

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: entUser.ID,
		"userName":                   entUser.Email,
		"name": map[string]interface{}{
			"givenName":  firstName,
			"familyName": lastName,
		},
		"displayName": entUser.DisplayName,
		"emails": []map[string]interface{}{
			{
				"value":   entUser.Email,
				"primary": true,
			},
		},
		"active": active,
		"groups": groups,
	}

	meta := scim.Meta{
		Created:      &entUser.CreatedAt,
		LastModified: &entUser.UpdatedAt,
		Version:      fmt.Sprintf("W/\"%d\"", entUser.UpdatedAt.Unix()),
	}

	return scim.Resource{
		ID:         entUser.ID,
		ExternalID: scimoptional.NewString(""),
		Attributes: attrs,
		Meta:       meta,
	}, nil
}
