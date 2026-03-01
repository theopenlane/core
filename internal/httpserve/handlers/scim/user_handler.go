package scim

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	scimoptional "github.com/elimity-com/scim/optional"
	scimschema "github.com/elimity-com/scim/schema"
	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/entx"
)

// UserHandler implements scim.ResourceHandler for User resources.
type UserHandler struct{}

// NewUserHandler creates a new UserHandler.
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// extractEmail extracts email address from SCIM attributes according to RFC 7643.
// Priority order:
// 1. Primary email from emails array
// 2. First email from emails array
// 3. userName if it's in valid email format
// Returns error if no valid email is found.
func extractEmail(attributes scim.ResourceAttributes) (string, error) {
	if emailsArray, ok := attributes["emails"].([]any); ok && len(emailsArray) > 0 {
		var firstEmail string

		for _, emailItem := range emailsArray {
			emailMap, ok := emailItem.(map[string]any)
			if !ok {
				continue
			}

			value, ok := emailMap["value"].(string)
			if !ok || value == "" {
				continue
			}

			if _, err := mail.ParseAddress(value); err != nil {
				continue
			}

			if firstEmail == "" {
				firstEmail = value
			}

			if primary, ok := emailMap["primary"].(bool); ok && primary {
				return value, nil
			}
		}

		if firstEmail != "" {
			return firstEmail, nil
		}
	}

	userName, _ := attributes["userName"].(string)
	if userName != "" {
		if _, err := mail.ParseAddress(userName); err == nil {
			return userName, nil
		}
	}

	return "", fmt.Errorf("%w: no valid email found in emails array or userName field", ErrInvalidAttributes)
}

// Create stores given attributes and returns a resource with the attributes that are stored and a unique identifier.
func (h *UserHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	ua, err := ExtractUserAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	userSetting, err := client.UserSetting.Create().SetEmailConfirmed(true).Save(ctx)
	if err != nil {
		return scim.Resource{}, fmt.Errorf("failed to create user settings: %w", err)
	}

	// SCIM users are provisioned by an IdP and will authenticate via SSO
	authProvider := enums.AuthProviderOIDC
	userRole := enums.RoleUser
	input := generated.CreateUserInput{
		Email:             ua.Email,
		DisplayName:       ua.DisplayName,
		AuthProvider:      &authProvider,
		LastLoginProvider: &authProvider,
		Role:              &userRole,
		ScimUsername:      &ua.UserName,
		ScimActive:        &ua.Active,
		SettingID:         userSetting.ID,
	}

	if ua.FirstName != "" {
		input.FirstName = &ua.FirstName
	}

	if ua.LastName != "" {
		input.LastName = &ua.LastName
	}

	if ua.ExternalID != "" {
		input.ScimExternalID = &ua.ExternalID
	}

	if ua.PreferredLanguage != "" {
		input.ScimPreferredLanguage = &ua.PreferredLanguage
	}

	if ua.Locale != "" {
		input.ScimLocale = &ua.Locale
	}

	if ua.ProfileURL != "" {
		input.AvatarRemoteURL = &ua.ProfileURL
	}

	entUser, err := client.User.Create().SetInput(input).Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to create user", fmt.Sprintf("User with email %s already exists", ua.Email))
	}

	if _, err := client.OrgMembership.Create().SetUserID(entUser.ID).SetOrganizationID(orgID).SetRole(enums.RoleMember).Save(ctx); err != nil {
		return scim.Resource{}, fmt.Errorf("failed to create org membership: %w", err)
	}

	if !ua.Active {
		updatedUser, err := client.User.UpdateOne(entUser).SetDeletedAt(time.Now()).SetScimActive(false).Save(ctx)
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
		return scim.Resource{}, ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	entUser, err := client.User.Query().Where(user.ID(id), user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID))).WithGroups(func(gq *generated.GroupQuery) { gq.Where(group.HasOwnerWith(organization.ID(orgID))) }).Only(ctx)
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
		return scim.Page{}, ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Page{}, err
	}

	query := client.User.Query().Where(user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)))

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

	users, err := query.Offset(offset).Limit(count).WithGroups(func(gq *generated.GroupQuery) { gq.Where(group.HasOwnerWith(organization.ID(orgID))) }).All(ctx)
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
		return scim.Resource{}, ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	entUser, err := client.User.Query().Where(user.ID(id), user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID))).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
		}

		return scim.Resource{}, fmt.Errorf("failed to get user: %w", err)
	}

	ua, err := ExtractUserAttributes(attributes)
	if err != nil {
		return scim.Resource{}, err
	}

	// SCIM users authenticate via SSO (typically OIDC or SAML), defaulting to OIDC
	authProvider := enums.AuthProviderOIDC
	update := client.User.UpdateOne(entUser).SetEmail(ua.Email).SetDisplayName(ua.DisplayName).SetAuthProvider(authProvider).SetScimUsername(ua.UserName).SetScimActive(ua.Active)

	if ua.ExternalID != "" {
		update.SetScimExternalID(ua.ExternalID)
	} else {
		update.ClearScimExternalID()
	}

	if ua.PreferredLanguage != "" {
		update.SetScimPreferredLanguage(ua.PreferredLanguage)
	} else {
		update.ClearScimPreferredLanguage()
	}

	if ua.Locale != "" {
		update.SetScimLocale(ua.Locale)
	} else {
		update.ClearScimLocale()
	}

	if ua.ProfileURL != "" {
		update.SetAvatarRemoteURL(ua.ProfileURL)
	} else {
		update.ClearAvatarRemoteURL()
	}

	if ua.FirstName != "" {
		update.SetFirstName(ua.FirstName)
	} else {
		update.ClearFirstName()
	}

	if ua.LastName != "" {
		update.SetLastName(ua.LastName)
	} else {
		update.ClearLastName()
	}

	if !ua.Active {
		update.SetDeletedAt(time.Now()).SetScimActive(false)
	} else {
		update.ClearDeletedAt().SetScimActive(true)
	}

	updatedUser, err := update.Save(ctx)
	if err != nil {
		return scim.Resource{}, HandleEntError(err, "failed to update user", fmt.Sprintf("User with email %s already exists", ua.Email))
	}

	return h.toSCIMResource(ctx, updatedUser, orgID)
}

// Delete removes the resource with corresponding ID.
func (h *UserHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return err
	}

	entUser, err := client.User.Query().Where(user.ID(id), user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID))).Only(ctx)
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
	ctx = entx.SkipSoftDelete(ctx)
	client := transaction.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return scim.Resource{}, ErrOrgNotFound
	}

	if err := ValidateSSOEnforced(ctx, orgID); err != nil {
		return scim.Resource{}, err
	}

	entUser, err := client.User.Query().Where(user.ID(id), user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID))).Only(ctx)
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
			return scim.Resource{}, HandleEntError(err, "failed to patch user", "Constraint violation during patch operation")
		}
	}

	return h.toSCIMResource(ctx, entUser, orgID)
}

func (h *UserHandler) applyPatchOperation(update *generated.UserUpdateOne, op scim.PatchOperation, modified *bool) error {
	pathStr := ""
	if op.Path != nil {
		pathStr = strings.ToLower(op.Path.String())
	}

	_, isMap := op.Value.(map[string]any)
	if !isMap && pathStr == "" {
		return fmt.Errorf("%w: patch operation requires path or value map", ErrInvalidAttributes)
	}

	if isMap {
		return h.applyMapPatchOperation(update, op, modified)
	}

	return h.applyPathPatchOperation(update, op, pathStr, modified)
}

// applyMapPatchOperation applies patch operations when the value is a map of attributes
func (h *UserHandler) applyMapPatchOperation(update *generated.UserUpdateOne, op scim.PatchOperation, modified *bool) error {
	patch, err := ExtractPatchUserAttribute(op)
	if err != nil {
		return err
	}

	applyStringField := func(value *string, setter func(string) *generated.UserUpdateOne) {
		if value != nil && *value != "" {
			setter(*value)
			*modified = true
		}
	}

	applyStringField(patch.ExternalID, update.SetScimExternalID)
	applyStringField(patch.PreferredLanguage, update.SetScimPreferredLanguage)
	applyStringField(patch.Locale, update.SetScimLocale)
	applyStringField(patch.ProfileURL, update.SetAvatarRemoteURL)
	applyStringField(patch.DisplayName, update.SetDisplayName)
	applyStringField(patch.FirstName, update.SetFirstName)
	applyStringField(patch.LastName, update.SetLastName)

	if patch.UserName != nil && patch.Email != nil {
		update.SetEmail(*patch.Email).SetScimUsername(*patch.UserName)
		*modified = true
	}

	if patch.Active != nil {
		h.setActiveStatus(update, *patch.Active)
		*modified = true
	}

	return nil
}

// applyPathPatchOperation applies patch operations when using a specific path attribute
func (h *UserHandler) applyPathPatchOperation(update *generated.UserUpdateOne, op scim.PatchOperation, pathStr string, modified *bool) error {
	applyStringPath := func(path string, setter func(string) *generated.UserUpdateOne) bool {
		if pathStr == path {
			if strVal, ok := op.Value.(string); ok && strVal != "" {
				setter(strVal)
				*modified = true
				return true
			}
		}
		return false
	}

	if applyStringPath("externalid", update.SetScimExternalID) ||
		applyStringPath("preferredlanguage", update.SetScimPreferredLanguage) ||
		applyStringPath("locale", update.SetScimLocale) ||
		applyStringPath("profileurl", update.SetAvatarRemoteURL) ||
		applyStringPath("displayname", update.SetDisplayName) ||
		applyStringPath("name.givenname", update.SetFirstName) ||
		applyStringPath("name.familyname", update.SetLastName) {
		return nil
	}

	switch pathStr {
	case "username":
		if strVal, ok := op.Value.(string); ok && strVal != "" {
			if _, err := mail.ParseAddress(strVal); err != nil {
				return fmt.Errorf("%w: userName must be a valid email address", ErrInvalidAttributes)
			}
			update.SetEmail(strVal).SetScimUsername(strVal)
			*modified = true
		}
	case "active":
		if boolVal, ok := op.Value.(bool); ok {
			h.setActiveStatus(update, boolVal)
			*modified = true
		}
	}

	return nil
}

// setActiveStatus sets the user active/inactive status using DeletedAt and ScimActive fields
func (h *UserHandler) setActiveStatus(update *generated.UserUpdateOne, active bool) {
	if !active {
		update.SetDeletedAt(time.Now()).SetScimActive(false)
	} else {
		update.ClearDeletedAt().SetScimActive(true)
	}
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

// toSCIMResource converts an ent User entity to a SCIM Resource representation
func (h *UserHandler) toSCIMResource(_ any, entUser *generated.User, _ string) (scim.Resource, error) {
	firstName := entUser.FirstName

	lastName := entUser.LastName

	active := entUser.DeletedAt.IsZero()

	groups := make([]map[string]any, 0)
	if entUser.Edges.Groups != nil {
		groups = lo.Map(entUser.Edges.Groups, func(g *generated.Group, _ int) map[string]any {
			return map[string]any{
				"value":   g.ID,
				"display": g.DisplayName,
				"$ref":    fmt.Sprintf("/v1/scim/Groups/%s", g.ID),
			}
		})
	}

	userName := entUser.Email
	if entUser.ScimUsername != nil && *entUser.ScimUsername != "" {
		userName = *entUser.ScimUsername
	}

	attrs := scim.ResourceAttributes{
		scimschema.CommonAttributeID: entUser.ID,
		"userName":                   userName,
		"name": map[string]any{
			"givenName":  firstName,
			"familyName": lastName,
		},
		"displayName": entUser.DisplayName,
		"emails": []map[string]any{
			{
				"value":   entUser.Email,
				"primary": true,
			},
		},
		"active": active,
		"groups": groups,
	}

	if entUser.ScimPreferredLanguage != nil && *entUser.ScimPreferredLanguage != "" {
		attrs["preferredLanguage"] = *entUser.ScimPreferredLanguage
	}

	if entUser.ScimLocale != nil && *entUser.ScimLocale != "" {
		attrs["locale"] = *entUser.ScimLocale
	}

	if entUser.AvatarRemoteURL != nil && *entUser.AvatarRemoteURL != "" {
		attrs["profileUrl"] = *entUser.AvatarRemoteURL
	}

	meta := scim.Meta{
		Created:      &entUser.CreatedAt,
		LastModified: &entUser.UpdatedAt,
		Version:      fmt.Sprintf("W/\"%d\"", entUser.UpdatedAt.Unix()),
	}

	externalID := scimoptional.NewString("")
	if entUser.ScimExternalID != nil && *entUser.ScimExternalID != "" {
		externalID = scimoptional.NewString(*entUser.ScimExternalID)
	}

	return scim.Resource{
		ID:         entUser.ID,
		ExternalID: externalID,
		Attributes: attrs,
		Meta:       meta,
	}, nil
}
