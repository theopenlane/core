package objects

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

var ErrMissingParent = fmt.Errorf("parent id or type is missing")

func AddFilePermissions(ctx context.Context) error {
	filesUpload, err := objects.FilesFromContext(ctx)
	if err != nil {
		// this is not an error, just means we are not uploading files
		return nil
	}

	for _, file := range filesUpload {
		for _, f := range file {
			if f.Parent.ID == "" || f.Parent.Type == "" {
				return ErrMissingParent
			}

			req := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   f.Parent.ID,
				SubjectType: f.Parent.Type,
				ObjectID:    f.ID,     // this is the object id (file id in this case) being created
				ObjectType:  "file",   // this is the object type (file in this case) being created
				Relation:    "parent", // this will always be parent in an object owned permission setup
			})

			if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, []fgax.TupleKey{req}, nil); err != nil {
				return err
			}

			log.Info().Interface("req", req).Msg("added file permissions")

			// remove file from context, we are done with it
			ctx = objects.RemoveFileFromContext(ctx, f)

			ec, err := echocontext.EchoContextFromContext(ctx)
			if err == nil {
				ec.SetRequest(ec.Request().WithContext(ctx))
			}
		}
	}

	return nil
}
