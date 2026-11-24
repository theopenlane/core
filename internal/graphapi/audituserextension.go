package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
)

const canViewAuditRelation = "can_view_audit"

// auditUserEdgeExtension prevents list-edge model methods on a synthetic audit-resolved user
type auditUserEdgeExtension struct{}

// ExtensionName satisfies the extension interface
func (auditUserEdgeExtension) ExtensionName() string { return "AuditUserEdges" }

// Validate satisfies the extension interface
func (auditUserEdgeExtension) Validate(graphql.ExecutableSchema) error { return nil }

// InterceptField returns empty connections for list edges on a synthetic audit user
func (auditUserEdgeExtension) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || fc.Object != "User" || fc.Parent == nil {
		return next(ctx)
	}

	rootFieldCtx := graphql.GetRootFieldContext(ctx)
	if rootFieldCtx == nil || rootFieldCtx.Object != "user" {
		return next(ctx)
	}

	if auth.HasAnyInContextCaller(ctx, auth.CapOrgSupport|auth.CapInternalOperation|auth.CapInternalRead|auth.CapSystemAdmin) {
		return next(ctx)
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok {
		return next(ctx)
	}

	parentUser, ok := fc.Parent.Result.(*generated.User)
	if !ok || parentUser == nil || parentUser.ID == caller.SubjectID {
		return next(ctx)
	}

	if res, ok := defaultUserEdges(fc.Field.Name); ok {
		return res, nil
	}

	return next(ctx)
}

// WithAuditUserEdges registers the audit-user edge extension
func WithAuditUserEdges(srv *handler.Server) {
	srv.Use(auditUserEdgeExtension{})
}

func defineAuditUser(u *generated.User) *generated.User {
	return &generated.User{
		ID:          u.ID,
		DisplayID:   u.DisplayID,
		DisplayName: u.DisplayName,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		Edges: generated.UserEdges{
			Setting:    &generated.UserSetting{UserID: u.ID},
			AvatarFile: &generated.File{},
		},
	}
}

func (r *queryResolver) resolveAuditUser(ctx context.Context, id string) (*generated.User, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.SubjectID == "" {
		return nil, nil
	}

	isAllowed, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  fgax.Kind(strcase.SnakeCase(generated.TypeUser)),
		ObjectID:    id,
		Relation:    canViewAuditRelation,
		SubjectID:   caller.SubjectID,
		SubjectType: caller.SubjectType(),
	})
	if err != nil || !isAllowed {
		return nil, err
	}

	auditCtx := auth.WithInternalReadContext(entx.SkipSoftDelete(ctx))

	u, err := withTransactionalMutation(ctx).User.Query().
		Where(user.ID(id)).
		Select(
			user.FieldID,
			user.FieldDisplayID,
			user.FieldDisplayName,
			user.FieldFirstName,
			user.FieldLastName,
			user.FieldEmail,
		).
		Only(auditCtx)
	if err != nil {
		return nil, err
	}

	return defineAuditUser(u), nil
}
