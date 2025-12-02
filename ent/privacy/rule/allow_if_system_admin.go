package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/shared/logx"
)

// AllowMutationIfSystemAdmin determines whether a mutation operation should be allowed based on whether the user is a system admin
func AllowMutationIfSystemAdmin() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking if user is system admin")
		return systemAdminCheck(ctx)
	})
}

// AllowQueryIfSystemAdmin determines whether a query operation should be allowed based on whether the user is a system admin
func AllowQueryIfSystemAdmin() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, _ ent.Query) error {
		return systemAdminCheck(ctx)
	})
}

// systemAdminCheck checks if the user is a system admin and returns an error if not
// it uses the context, instead of checking the authz client directly
// this value will be set my the auth middleware
func systemAdminCheck(ctx context.Context) error {
	allow, err := CheckIsSystemAdminWithContext(ctx)
	if err != nil {
		return err
	}

	if allow {
		return privacy.Allow
	}

	// this covers the impersonation path where the
	// auth context is no longer an admin, but
	// the original user was an admin
	admin, ok := auth.SystemAdminFromContext(ctx)
	if ok && admin.IsSystemAdmin {
		return privacy.Allow
	}

	// if not a system admin, skip to the next rule
	return privacy.Skip
}

// CheckIsSystemAdminWithContext checks if the user is a system admin based on the authz service
// using the authz client from the context
func CheckIsSystemAdminWithContext(ctx context.Context) (bool, error) {
	return auth.IsSystemAdminFromContext(ctx), nil
}
