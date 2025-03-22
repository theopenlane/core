package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

const (
	// system object type for FGA authorization
	SystemObject = "system"
	// system object id for FGA authorization
	SystemObjectID = "openlane_core"
)

// AllowMutationIfSystemAdmin determines whether a mutation operation should be allowed based on whether the user is a system admin
func AllowMutationIfSystemAdmin() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		allow, err := CheckIsSystemAdmin(ctx, m)
		if err != nil {
			return err
		}

		if allow {
			return privacy.Allow
		}

		// if not a system admin, skip to the next rule
		return privacy.Skip
	})
}

// AllowMutationIfSystemAdmin determines whether a query operation should be allowed based on whether the user is a system admin
func AllowQueryIfSystemAdmin() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		allow, err := CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if allow {
			return privacy.Allow
		}

		// if not a system admin, skip to the next rule
		return privacy.Skip
	})
}

// CheckIsSystemAdmin checks if the user is a system admin based on the authz service
func CheckIsSystemAdmin(ctx context.Context, m ent.Mutation) (bool, error) {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	ac := fgax.AccessCheck{
		ObjectType:  SystemObject,
		ObjectID:    SystemObjectID,
		Relation:    fgax.SystemAdminRelation,
		SubjectID:   au.SubjectID,
		SubjectType: getSubjectType(au),
	}

	return utils.AuthzClient(ctx, m).CheckAccess(ctx, ac)
}

// CheckIsSystemAdminWithContext checks if the user is a system admin based on the authz service
// using the authz client from the context
func CheckIsSystemAdminWithContext(ctx context.Context) (bool, error) {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return false, err
	}

	ac := fgax.AccessCheck{
		ObjectType:  SystemObject,
		ObjectID:    SystemObjectID,
		Relation:    fgax.SystemAdminRelation,
		SubjectID:   au.SubjectID,
		SubjectType: getSubjectType(au),
	}

	return utils.AuthzClientFromContext(ctx).CheckAccess(ctx, ac)
}

// getSubjectType gets the subject type based on the authentication type
func getSubjectType(au *auth.AuthenticatedUser) string {
	// determine the subject type
	if au.AuthenticationType == auth.APITokenAuthentication {
		return auth.ServiceSubjectType
	}

	return auth.UserSubjectType
}
