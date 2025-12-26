package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupsetting"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/common/enums"
)

// HookGroupSettingVisibility is a hook that updates the conditional tuples for group settings
// based on the visibility setting changing
// the initial tuple is set up on group creation
func HookGroupSettingVisibility() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.GroupSettingFunc(func(ctx context.Context, m *generated.GroupSettingMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			addTuplesPerSettingUpdate, err := getVisibilityTuples(ctx, m, retVal)
			if err != nil {
				return nil, err
			}

			// nothing to do, skip
			if addTuplesPerSettingUpdate == nil {
				return retVal, err
			}

			// update the visibility tuples
			for _, addTuples := range addTuplesPerSettingUpdate {
				if _, err := m.Authz.UpdateConditionalTupleKey(ctx, addTuples); err != nil {
					return nil, err
				}
			}

			return retVal, err
		})
	},
		hook.And(
			hook.HasFields("visibility"),
			hook.HasOp(ent.OpUpdate|ent.OpUpdateOne),
		),
	)
}

// getVisibilityTuples returns the visibility tuples based on the group setting visibility setting
// it will return nil if no visibility change is detected
func getVisibilityTuples(ctx context.Context, m *generated.GroupSettingMutation, retVal any) ([]fgax.TupleKey, error) {
	visibility, ok := m.Visibility()
	if !ok {
		return nil, nil
	}

	groupIDs, err := getGroupIDFromSettingMutation(ctx, m, retVal)
	if err != nil {
		return nil, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, nil
	}

	oldVisibility, err := m.OldVisibility(ctx)
	if err != nil {
		return nil, nil
	}

	if oldVisibility == visibility {
		return nil, nil
	}

	var writeReqs []fgax.TupleKey

	for _, groupID := range groupIDs {
		writeReq, err := createGroupParentTuple(orgID, groupID, visibility == enums.VisibilityPublic)
		if err != nil {
			return nil, err
		}

		writeReqs = append(writeReqs, writeReq...)
	}

	return writeReqs, nil
}

// getGroupIDFromSettingMutation returns the group ID(s) from the group setting mutation or return value
func getGroupIDFromSettingMutation(ctx context.Context, m *generated.GroupSettingMutation, retVal any) ([]string, error) {
	// if we have it just return it
	groupID, ok := m.GroupID()
	if ok && groupID != "" {
		return []string{groupID}, nil
	}

	groupIDs := m.GroupIDs()
	if len(groupIDs) > 0 {
		return groupIDs, nil
	}

	// otherwise get from the settings
	var (
		err        error
		settingIDs []string
	)

	settingIDs, err = GetObjectIDsFromMutation(ctx, m, retVal)
	if err != nil {
		return nil, err
	}

	return m.Client().Group.Query().
		Where(group.HasSettingWith(groupsetting.IDIn(settingIDs...))).
		IDs(ctx)
}
