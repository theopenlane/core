package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
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
	return func(next ent.Mutator) ent.Mutator {
		return hook.GroupFunc(func(ctx context.Context, m *generated.GroupMutation) (ent.Value, error) {
			if m.Op().Is(ent.OpCreate) {
				return next.Mutate(ctx, m)
			}

			groupID, ok := m.ID()
			if !ok || groupID == "" {
				return next.Mutate(ctx, m)
			}

			group, err := m.Client().Group.Get(ctx, groupID)
			if err != nil {
				log.Error().Err(err).Msg("failed to get group")

				return nil, err
			}

			// allow general allow context to bypass managed group check
			_, allowCtx := privacy.DecisionFromContext(ctx)
			_, allowManagedCtx := contextx.From[ManagedContextKey](ctx)

			if group.IsManaged && (!allowManagedCtx && !allowCtx) {
				return nil, ErrManagedGroup
			}

			return next.Mutate(ctx, m)
		})
	}
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

	if exists {
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
	}

	org, orgExists := m.OwnerID()

	if exists && orgExists {
		req := fgax.TupleRequest{
			SubjectID:   org,
			SubjectType: "organization",
			ObjectID:    objID,
			ObjectType:  GetObjectTypeFromEntMutation(m),
		}

		log.Debug().Interface("tuple", req).Msg("creating relationship tuple")

		orgTuple, err := getTupleKeyFromRole(req, fgax.ParentRelation)
		if err != nil {
			return err
		}

		if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{orgTuple}, nil); err != nil {
			log.Error().Err(err).Msg("failed to create relationship tuple")

			return ErrInternalServerError
		}
	}

	return nil
}

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
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user id from context, unable to add user to group")

		return err
	}

	// Add user as admin of group
	input := generated.CreateGroupMembershipInput{
		UserID:  userID,
		GroupID: gID,
		Role:    &role,
	}

	if _, err := m.Client().GroupMembership.Create().SetInput(input).Save(ctx); err != nil {
		log.Error().Err(err).Msg("error creating group membership for admin")

		return err
	}

	return nil
}

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

	log.Debug().Str("object", object).Msg("deleting relationship tuples")

	// delete all relationship tuples except for the user, those are handled by the cascade delete of the group membership
	if err := m.Authz.DeleteAllObjectRelations(ctx, object, userRoles); err != nil {
		log.Error().Err(err).Msg("failed to delete relationship tuples")

		return ErrInternalServerError
	}

	log.Debug().Str("object", object).Msg("deleted relationship tuples")

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
