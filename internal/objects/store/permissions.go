package store

import (
	"context"
	"errors"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
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
				SubjectType: strcase.SnakeCase(f.Parent.Type),
				ObjectID:    f.ID,
				ObjectType:  generated.TypeFile,
				Relation:    parentRelation,
			})

			tuples := []fgax.TupleKey{req}

			const avatarFileKey = "avatarFile"
			if f.FieldName == avatarFileKey {
				au, err := auth.GetAuthenticatedUserFromContext(ctx)
				if err != nil {
					return ctx, err
				}

				orgID := au.OrganizationID
				if orgID == "" {
					if len(au.OrganizationIDs) == 1 {
						orgID = au.OrganizationIDs[0]
					} else {
						return ctx, ErrMissingOrganizationID
					}
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

			logx.FromContext(ctx).Debug().Interface("req", req).Msg("adding file permissions")

			if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, tuples, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to write tuple keys")

				return ctx, err
			}

			logx.FromContext(ctx).Debug().Interface("req", req).Msg("added file permissions")

			ctx = pkgobjects.RemoveFileFromContext(ctx, f)

			ec, err := echocontext.EchoContextFromContext(ctx)
			if err == nil {
				ec.SetRequest(ec.Request().WithContext(ctx))
			}
		}
	}

	return ctx, nil
}
