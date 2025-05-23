package hooks

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/gravatar"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
)

// HookGroup runs on group mutations to set default values that are not provided
func HookGroup() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupFunc(func(ctx context.Context, m *generated.GroupMutation) (generated.Value, error) {
			if name, ok := m.Name(); ok {
				displayName, _ := m.DisplayName()

				if displayName == "" {
					m.SetDisplayName(name)
				}
			}

			if m.Op().Is(ent.OpCreate) {
				// if this is empty generate a default group setting schema
				settingID, _ := m.SettingID()
				if settingID == "" {
					// sets up default group settings using schema defaults
					groupSettingID, err := defaultGroupSettings(ctx, m)
					if err != nil {
						return nil, err
					}

					// add the group setting ID to the input
					m.SetSettingID(groupSettingID)
				}
			}

			if name, ok := m.Name(); ok {
				// trim trailing whitespace from the name
				m.SetName(strings.TrimSpace(name))

				url := gravatar.New(name, nil)
				m.SetGravatarLogoURL(url)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookManagedGroups runs on group mutations to prevent updates to managed groups
func HookManagedGroups() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupFunc(func(ctx context.Context, m *generated.GroupMutation) (ent.Value, error) {
			groupID, ok := m.ID()
			if !ok || groupID == "" {
				return next.Mutate(ctx, m)
			}

			group, err := m.Client().Group.Get(ctx, groupID)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get group")

				return nil, err
			}

			// allow general allow context to bypass managed group check
			_, allowCtx := privacy.DecisionFromContext(ctx)
			_, allowManagedCtx := contextx.From[ManagedContextKey](ctx)

			// before returning the error, we need to allow for edges to be updated
			// if they are permissions edges
			if group.IsManaged && (!allowManagedCtx && !allowCtx) {
				if err := checkOnlyDefaultFields(m); err != nil {
					return nil, ErrManagedGroup
				}

				if err := checkOnlyPermissionEdges(m); err != nil {
					return nil, err
				}
			}

			// if we got here, the only that that was updated was edges for permissions (Editor, Viewer, BlockedGroups)
			// and we can continue

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne)
}

// checkOnlyDefaultFields checks if the added or cleared fields are only default fields
// and returns an error if they are not
func checkOnlyDefaultFields(m *generated.GroupMutation) error {
	fields := m.Fields()
	numericFields := m.AddedFields()
	clearedFields := m.ClearedFields()

	// if nothing changed, return no error
	if len(fields) == 0 && len(numericFields) == 0 && len(clearedFields) == 0 {
		return nil
	}

	// default fields are updatedAt, updatedBy
	defaultFields := []string{
		"updated_at",
		"updated_by",
		// TODO: see why this is sent in the mutation, added a test to confirm it doesn't actually change
		"display_id",
	}

	// check if any of the fields are not default fields
	for _, field := range fields {
		if !slices.Contains(defaultFields, field) {
			return ErrManagedGroup
		}
	}

	return nil
}

// checkOnlyPermissionEdges checks if the added or cleared edges are only permission edges
// and returns an error if they are not
func checkOnlyPermissionEdges(m *generated.GroupMutation) error {
	addedEdges := m.AddedEdges()
	clearedEdges := m.ClearedEdges()

	if len(addedEdges) > 0 || len(clearedEdges) > 0 {
		for _, edge := range addedEdges {
			_, _, isPermissionEdge := isPermissionsEdge(edge)
			if !isPermissionEdge {
				return ErrManagedGroup
			}
		}

		for _, edge := range clearedEdges {
			_, _, isPermissionEdge := isPermissionsEdge(edge)
			if !isPermissionEdge {
				return ErrManagedGroup
			}
		}
	}

	return nil
}

// HookGroupAuthz runs on group mutations to setup or remove relationship tuples
func HookGroupAuthz() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return hook.GroupFunc(func(ctx context.Context, m *generated.GroupMutation) (ent.Value, error) {
			// do the mutation, and then create/delete the relationship
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				// if we error, do not attempt to create the relationships
				return retValue, err
			}

			if m.Op().Is(ent.OpCreate) {
				// create the group member admin and relationship tuple for parent org
				err = groupCreateHook(ctx, m)
			} else if m.Op().Is(ent.OpDelete|ent.OpDeleteOne) || entx.CheckIsSoftDelete(ctx) {
				// delete all relationship tuples on delete, or soft delete (Update Op)
				err = groupDeleteHook(ctx, m)
			}

			return retValue, err
		})
	}
}

func groupCreateHook(ctx context.Context, m *generated.GroupMutation) error {
	objID, exists := m.ID()
	if !exists {
		return nil
	}

	// create the admin group member if not using an API token (which is not associated with a user)
	if !auth.IsAPITokenAuthentication(ctx) {
		if err := createGroupMember(ctx, objID, m); err != nil {
			return err
		}
	} else {
		if err := addTokenEditPermissions(ctx, m, objID, GetObjectTypeFromEntMutation(m)); err != nil {
			return err
		}
	}

	// create the relationship tuple for the parent org
	org, orgExists := m.OwnerID()
	if !orgExists {
		// skip if the owner is not set
		return nil
	}

	// determine if the group is public
	publicGroup := true

	setting, ok := m.SettingID()
	if ok {
		// allow before tuples may be created
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

		groupSetting, err := m.Client().GroupSetting.Get(allowCtx, setting)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get group setting")

			return err
		}

		publicGroup = groupSetting.Visibility == enums.VisibilityPublic
	}

	groupTuple, err := createGroupParentTuple(org, objID, publicGroup)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to get tuple key")

		return err
	}

	if groupTuple == nil {
		return nil
	}

	if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{*groupTuple}, nil); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to create relationship tuple")

		return ErrInternalServerError
	}

	return nil
}

// createGroupParentTuple creates a relationship tuple for a group
func createGroupParentTuple(orgID, groupID string, isPublic bool) (*fgax.TupleKey, error) {
	const (
		conditionName = "public_group"
		contextKey    = "public"
	)

	req := fgax.TupleRequest{
		SubjectID:     orgID,
		SubjectType:   generated.TypeOrganization,
		ObjectID:      groupID,
		ObjectType:    generated.TypeGroup,
		ConditionName: conditionName,
		ConditionContext: &map[string]any{
			contextKey: isPublic,
		},
	}

	groupTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
	if err != nil {
		return nil, err
	}

	return &groupTuple, err
}

// createGroupMember creates a group membership for the authorized user who triggered the group creation
func createGroupMember(ctx context.Context, gID string, m *generated.GroupMutation) error {
	managed, _ := m.IsManaged()
	groupName, _ := m.Name()

	role := enums.RoleAdmin

	if managed {
		// do not add the owner to the Members group
		if groupName == ViewersGroup {
			return nil
		}

		// managed groups do not have owners, add them as a member
		role = enums.RoleMember
	}

	// get userID from context
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("unable to get user id from context, unable to add user to group")

		return err
	}

	// Add user as admin of group
	input := generated.CreateGroupMembershipInput{
		UserID:  userID,
		GroupID: gID,
		Role:    &role,
	}

	if err := m.Client().GroupMembership.Create().SetInput(input).Exec(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error creating group membership for admin")

		return err
	}

	return nil
}

// groupDeleteHook deletes all relationship tuples for a group on delete
// with the exception of the user, those are handled by the cascade delete of the group membership
func groupDeleteHook(ctx context.Context, m *generated.GroupMutation) error {
	// Add relationship tuples if authz is enabled
	objID, ok := m.ID()
	if !ok {
		// TODO (sfunk): ensure tuples get cascade deleted
		// continue for now
		return nil
	}

	objType := GetObjectTypeFromEntMutation(m)
	object := fmt.Sprintf("%s:%s", objType, objID)

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleting relationship tuples")

	// delete all relationship tuples except for the user, those are handled by the cascade delete of the group membership
	if err := m.Authz.DeleteAllObjectRelations(ctx, object, userRoles); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	zerolog.Ctx(ctx).Debug().Str("object", object).Msg("deleted relationship tuples")

	return nil
}

// defaultGroupSettings creates the default group settings for a new group
func defaultGroupSettings(ctx context.Context, m *generated.GroupMutation) (string, error) {
	input := generated.CreateGroupSettingInput{}

	groupSetting, err := m.Client().GroupSetting.Create().SetInput(input).Save(ctx)
	if err != nil {
		return "", err
	}

	return groupSetting.ID, nil
}
