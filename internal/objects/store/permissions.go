package store

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

var (
	// ErrMissingParent is returned when the parent id or type is missing for file uploads
	ErrMissingParent = errors.New("parent id or type is missing")
	// ErrMissingOrganizationID is returned when organization ID cannot be determined for file upload
	ErrMissingOrganizationID = errors.New("organization ID is required for file upload")
)

// AddFilePermissions writes authorization tuples for uploaded files and removes them from the request context.
func AddFilePermissions(ctx context.Context) (context.Context, error) {
	filesUpload, err := pkgobjects.FilesFromContext(ctx)
	if err != nil {
		return ctx, nil
	}

	for _, file := range filesUpload {
		for _, f := range file {
			if f.Parent.ID == "" || f.Parent.Type == "" {
				return ctx, ErrMissingParent
			}

			parentRelation := fgax.ParentRelation
			if f.Parent.Type == "trust_center_doc" {
				parentRelation = "tc_doc_parent"
			}

			req := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   f.Parent.ID,
				SubjectType: f.Parent.Type,
				ObjectID:    f.ID,
				ObjectType:  generated.TypeFile,
				Relation:    parentRelation,
			})

			tuples := []fgax.TupleKey{req}

			const avatarFileKey = "avatarFile"
			if f.FieldName == avatarFileKey {
				orgID, err := auth.GetOrganizationIDFromContext(ctx)
				if err != nil {
					return ctx, err
				}

				orgReq := fgax.GetTupleKey(fgax.TupleRequest{
					SubjectID:       orgID,
					SubjectType:     generated.TypeOrganization,
					SubjectRelation: fgax.MemberRelation,
					ObjectID:        f.ID,
					ObjectType:      generated.TypeFile,
					Relation:        fgax.CanView,
				})
				tuples = append(tuples, orgReq)
			}

			log.Debug().Interface("req", req).Msg("adding file permissions")

			if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, tuples, nil); err != nil {
				return ctx, err
			}

			log.Debug().Interface("req", req).Msg("added file permissions")

			ctx = pkgobjects.RemoveFileFromContext(ctx, f)

			ec, err := echocontext.EchoContextFromContext(ctx)
			if err == nil {
				ec.SetRequest(ec.Request().WithContext(ctx))
			}
		}
	}

	return ctx, nil
}
