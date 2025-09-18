package hooks

import (
	"context"

	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// AvatarMutation is an interface for setting the local file ID for an avatar
type AvatarMutation interface {
	SetAvatarLocalFileID(s string)
	ID() (id string, exists bool)
	Type() string
}

// checkAvatarFile checks if an avatar file is provided and sets the local file ID
// this can be used for any schema that has an avatar field
func checkAvatarFile[T AvatarMutation](ctx context.Context, m T) (context.Context, error) {
	key := "avatarFile"

	// get the file from the context, if it exists
	file, _ := pkgobjects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrTooManyAvatarFiles
	}

	// this should always be true, but check just in case
	if file[0].FieldName == key {
		m.SetAvatarLocalFileID(file[0].ID)

		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = m.Type()

		ctx = pkgobjects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
