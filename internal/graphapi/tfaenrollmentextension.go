package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/rout"
	"github.com/vektah/gqlparser/v2/ast"
)

// tfaEnrollmentRootFields lists the root fields an enrollment-only session may run
var tfaEnrollmentRootFields = map[ast.Operation]map[string]bool{
	ast.Query: {
		"tfaSetting":  true,
		"tfaSettings": true,
	},
	ast.Mutation: {
		"createTFASetting": true,
		"updateTFASetting": true,
	},
}

// tfaEnrollmentExtension denies callers holding an enrollment-only session every root field other than the ones
// needed to enroll in two factor authentication
type tfaEnrollmentExtension struct{}

// ExtensionName satisfies the extension interface
func (tfaEnrollmentExtension) ExtensionName() string {
	return "TFAEnrollment"
}

// Validate satisfies the extension interface
func (tfaEnrollmentExtension) Validate(graphql.ExecutableSchema) error {
	return nil
}

// InterceptRootField denies enrollment-only sessions root fields outside the enrollment allow list
func (tfaEnrollmentExtension) InterceptRootField(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
	if !auth.IsTFAEnrollmentRequiredFromContext(ctx) {
		return next(ctx)
	}

	field := graphql.GetRootFieldContext(ctx).Field

	if !tfaEnrollmentRootFields[graphql.GetOperationContext(ctx).Operation.Operation][field.Name] {
		logx.FromContext(ctx).Warn().Str("field", field.Name).Msg("enrollment-only session attempted a field outside the tfa enrollment allow list")

		graphql.AddError(ctx, rout.ErrPermissionDenied)

		return graphql.Null
	}

	return next(ctx)
}

// WithTFAEnrollmentBlock registers the tfa enrollment extension
func WithTFAEnrollmentBlock(srv *handler.Server) {
	srv.Use(tfaEnrollmentExtension{})
}
