package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
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

// CheckIsSystemAdmin checks if the user is a system admin based on the authz service
func CheckIsSystemAdmin(ctx context.Context, m ent.Mutation) (bool, error) {
	au, err := auth.GetAuthenticatedUserContext(ctx)
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

// getSubjectType gets the subject type based on the authentication type
func getSubjectType(au *auth.AuthenticatedUser) string {
	// determine the subject type
	if au.AuthenticationType == auth.APITokenAuthentication {
		return auth.ServiceSubjectType
	}

	return auth.UserSubjectType
}
