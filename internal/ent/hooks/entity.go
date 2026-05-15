package hooks

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// HookEntityFiles runs on entity mutations to check for uploaded files
func HookEntityFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, "entityFiles")
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

var (
	approvedStatus = []enums.EntityStatus{enums.EntityStatusApproved, enums.EntityStatusActive}
)

// HookEntityApprovedForUse sets approved_for_use based on the entity status.
func HookEntityApprovedForUse() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			_, ok := m.ApprovedForUse()
			if ok {
				return next.Mutate(ctx, m)
			}

			status, _ := m.Status()

			m.SetApprovedForUse(slices.Contains(approvedStatus, status))

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasFields("status"),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne),
		),
	)
}

// HookEntityLogoFile runs on entity mutations to check for an uploaded logo file
func HookEntityLogoFile() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkEntityLogoFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkEntityLogoFile(ctx context.Context, m *generated.EntityMutation) (context.Context, error) {
	logoKey := "logoFile"

	logoFiles, _ := pkgobjects.FilesFromContextWithKey(ctx, logoKey)
	if len(logoFiles) == 0 {
		return ctx, nil
	}

	if len(logoFiles) > 1 {
		return ctx, ErrTooManyLogoFiles
	}

	m.SetLogoFileID(logoFiles[0].ID)

	return pkgobjects.ProcessFilesForMutation(ctx, m, logoKey)
}

// HookEntityCreate runs on entity mutations to set default values that are not provided
func HookEntityCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EntityFunc(func(ctx context.Context, m *generated.EntityMutation) (generated.Value, error) {
			// require either a display name or a name
			displayName, _ := m.DisplayName()
			name, _ := m.Name()

			// exit early if we have no name
			if displayName == "" && name == "" {
				return nil, ErrMissingRequiredName
			}

			// set display name based on name if it isn't set
			if displayName == "" {
				m.SetDisplayName(name)
			}

			// set unique name based on display name if it isn't set
			if name == "" {
				uniqueName := fmt.Sprintf("%s-%s", displayName, ulids.New().String())
				m.SetName(uniqueName)
			}

			// trim trailing whitespace from the name
			name, _ = m.Name() // re-fetch incase it was updated above
			m.SetName(strings.TrimSpace(name))

			if err := enrichLogoFromSubprocessor(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

func enrichLogoFromSubprocessor(ctx context.Context, m *generated.EntityMutation) error {
	names := getNameMatches(m)
	if len(names) == 0 {
		return nil
	}

	defaultPredicates := make([]predicate.Subprocessor, 0, len(names))
	for _, name := range names {
		defaultPredicates = append(defaultPredicates, subprocessor.NameEqualFold(name))
	}

	processors, err := m.Client().Subprocessor.Query().
		Where(
			subprocessor.SystemOwned(true),
			subprocessor.Or(defaultPredicates...),
		).
		Select(
			subprocessor.FieldDescription,
			subprocessor.FieldLogoRemoteURL,
			subprocessor.FieldLogoFileID,
		).
		First(ctx)
	if generated.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if description, ok := m.Description(); !ok || strings.TrimSpace(description) == "" {
		if strings.TrimSpace(processors.Description) != "" {
			m.SetDescription(processors.Description)
		}
	}

	if _, ok := m.LogoFileID(); !ok && processors.LogoFileID != nil && *processors.LogoFileID != "" {
		m.SetLogoFileID(*processors.LogoFileID)
	}

	if _, ok := m.LogoRemoteURL(); !ok && processors.LogoRemoteURL != nil && *processors.LogoRemoteURL != "" {
		m.SetLogoRemoteURL(*processors.LogoRemoteURL)
	}

	return nil
}

func getNameMatches(m *generated.EntityMutation) []string {
	var names []string

	fn := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}

		for _, v := range names {
			if strings.EqualFold(v, name) {
				return
			}
		}

		names = append(names, name)
	}

	if name, ok := m.DisplayName(); ok {
		fn(name)
	}

	if name, ok := m.Name(); ok {
		fn(name)
	}

	return names
}
