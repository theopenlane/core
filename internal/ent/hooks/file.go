package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
)

var (
	errInvalidStoragePath = errors.New("invalid path when deleting file from object storage")
)

// HookFileDelete makes sure to clean up the file from external storage once deleted
func HookFileDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(
			func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {

				var storagePath string
				if m.ObjectManager != nil && isDeleteOp(ctx, m) {

					id, ok := m.ID()
					if !ok {
						return nil, errInvalidStoragePath
					}

					file, err := m.Client().File.Get(ctx, id)
					if err != nil {
						return nil, err
					}

					storagePath = file.StoragePath
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				if storagePath != "" {
					if err := m.ObjectManager.Storage.Delete(ctx, storagePath); err != nil {
						return nil, err
					}
				}

				return v, err
			})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}

// HookFileCreate adds the organization id to the file if its not a user file
// we don't use the normal org owned hook because this is a special case (uses organization vs owner)
// as well as multiple parents, which is not supported by the org owned hook
// and we don't want to convolute that function more than needed
func HookFileCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.FileFunc(
			func(ctx context.Context, m *generated.FileMutation) (generated.Value, error) {
				if fileOrgSkipper(ctx) {
					return next.Mutate(ctx, m)
				}

				// add the organization id to the file if its not a user file
				orgID, err := getOrgIDForFile(ctx)
				if err != nil {
					return nil, err
				}

				if orgID == "" {
					return nil, fmt.Errorf("owner_id is required, %w", ErrFieldRequired)
				}

				m.AddOrganizationIDs(orgID)

				return next.Mutate(ctx, m)
			})
	}, ent.OpCreate)
}

// fileOrgSkipper skips the organization hook if the files are user owned
// most files should be linked back to an organization, but some like an avatar file for a user,
// should be available across all organizations the user has access to
func fileOrgSkipper(ctx context.Context) bool {
	if files, err := objects.FilesFromContext(ctx); err == nil && len(files) > 0 {
		for _, f := range files {
			for _, p := range f {
				if strings.EqualFold(p.Type, generated.TypeUser) {
					return true
				}
			}
		}
	}

	return false
}

// getOrgIDForFile gets the organization id for a file
// this first checks the context for the organization id
// if it's not found, it will attempt to get the organization id from the request context
// this is used when the owner id isn't yet set in the context because a filed
// is created by the middleware before the context could be set for personal
// access tokens; this is safe because the transaction will be rolled back later
// if the user has no access to the organization
func getOrgIDForFile(ctx context.Context) (string, error) {
	// add the organization id to the file if its not a user file
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		// check input instead
		input := graphutils.GetMapInputVariableByName(ctx, "input")
		if input != nil {
			i := *input
			if i["ownerID"] != nil {
				owner := i["ownerID"]
				if owner, ok := owner.(string); ok {
					orgID = owner
				}
			}
		}
	}

	if orgID == "" {
		return "", ErrOrganizationIDRequired
	}

	return orgID, nil
}
