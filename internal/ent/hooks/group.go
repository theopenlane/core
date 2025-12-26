package hooks

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/gravatar"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/logx"
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

				// if managed, the user's name ( and thus group name )
				// may include special characters. this makes sure to clean them
				// up as they will fail otherwise
				isManaged, _ := m.IsManaged()
				if isManaged {
					m.SetName(StripInvalidChars(name))
				}
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

			g, err := m.Client().Group.Query().
				Where(group.ID(groupID)).
				Select(group.FieldIsManaged).
				Only(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get group")

				return nil, err
			}

			// allow general allow context to bypass managed group check
			_, allowCtx := privacy.DecisionFromContext(ctx)
			_, allowManagedCtx := contextx.From[ManagedContextKey](ctx)

			// before returning the error, we need to allow for edges to be updated
			// if they are permissions edges
			if g.IsManaged && (!allowManagedCtx && !allowCtx) {
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

	if auth.IsAPITokenAuthentication(ctx) {
		if err := addTokenEditPermissions(ctx, m, objID, GetObjectTypeFromEntMutation(m)); err != nil {
			return err
		}
	}

	// determine if the group is public
	publicGroup := true

	setting, ok := m.SettingID()
	if ok {
		// allow before tuples may be created
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

		groupSetting, err := m.Client().GroupSetting.Get(allowCtx, setting)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get group setting")

			return err
		}

		publicGroup = groupSetting.Visibility == enums.VisibilityPublic
	}

	org, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	groupTuples, err := createGroupParentTuple(org, objID, publicGroup)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get tuple key")

		return err
	}

	if len(groupTuples) > 0 {
		if _, err := m.Authz.WriteTupleKeys(ctx, groupTuples, nil); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

// createGroupParentTuple creates a relationship tuple for a group
func createGroupParentTuple(orgID, groupID string, isPublic bool) ([]fgax.TupleKey, error) {
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

	tuples := []fgax.TupleKey{groupTuple}

	reqOwner := fgax.TupleRequest{
		SubjectID:       orgID,
		SubjectType:     generated.TypeOrganization,
		SubjectRelation: fgax.OwnerRelation,
		ObjectID:        groupID,
		ObjectType:      generated.TypeGroup,
		Relation:        "parent_admin",
	}

	tuples = append(tuples, fgax.GetTupleKey(reqOwner))

	return tuples, nil
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

// StripInvalidChars removes invalid characters from a string
func StripInvalidChars(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !strings.ContainsRune(validator.InvalidChars, r) {
			b.WriteRune(r)
		}
	}

	return b.String()
}
