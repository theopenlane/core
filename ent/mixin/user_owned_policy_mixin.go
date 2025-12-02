package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/utils"
)

type (
	// UserOwnedMutationPolicyMixin is a mixin for user owned mutation policy
	UserOwnedMutationPolicyMixin struct {
		mixin.Schema
		AllowAdminMutation bool
	}

	// UserOwnedQueryPolicyMixin is a mixin for user owned query policy
	UserOwnedQueryPolicyMixin struct {
		mixin.Schema
	}
)

// UserOwnedMutationPolicyMixin sets the policy for updating owned fields by the user
func (mixin UserOwnedMutationPolicyMixin) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.OnMutationOperation(
				utils.NewMutationPolicyWithoutNil(privacy.MutationPolicy{
					rule.AllowMutationAfterApplyingOwnerFilter(),
					privacy.AlwaysDenyRule(),
				}),
				ent.OpCreate,
			),
			privacy.OnMutationOperation(
				utils.NewMutationPolicyWithoutNil(privacy.MutationPolicy{
					rule.AllowMutationAfterApplyingOwnerFilter(),
					privacy.AlwaysDenyRule(),
				}),
				ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete,
			),
		},
	}
}
