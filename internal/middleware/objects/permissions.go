package objects

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

// ErrMissingParent is returned when the parent id or type is missing for file uploads
var ErrMissingParent = fmt.Errorf("parent id or type is missing")

// AddFilePermissions adds file permissions to the object store
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

			tuples := []fgax.TupleKey{req}

			// if the parent of the file is a user, explicitly add view permissions for org members
			if f.Parent.Type == "User" {
				orgID, err := auth.GetOrganizationIDFromContext(ctx)
				if err != nil {
					return err
				}

				orgReq := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:   orgID + "#member",
					SubjectType: "organization",
					ObjectID:    f.ID,
					ObjectType:  "file",
					Relation:    "can_view",
				})
				tuples = append(tuples, orgReq)
			}

			if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, tuples, nil); err != nil {
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
