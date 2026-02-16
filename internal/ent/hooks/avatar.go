package hooks

import (
	"context"

	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// AvatarMutation is an interface for setting the local file ID for an avatar
type AvatarMutation interface {
	pkgobjects.Mutation

	SetAvatarLocalFileID(s string)
}

// checkAvatarFile checks if an avatar file is provided and sets the local file ID
// this can be used for any schema that has an avatar field
func checkAvatarFile[T AvatarMutation](ctx context.Context, m T) (context.Context, error) {
	key := "avatarFile"

	files, _ := pkgobjects.FilesFromContextWithKey(ctx, key)
	if len(files) == 0 {
		return ctx, nil
	}

	if len(files) > 1 {
		return ctx, ErrTooManyAvatarFiles
	}

	m.SetAvatarLocalFileID(files[0].ID)

	return pkgobjects.ProcessFilesForMutation(ctx, m, key)
}
