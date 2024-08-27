package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

type (
	UserOwnedMutationPolicyMixin struct {
		mixin.Schema
		AllowAdminMutation bool
	}

	UserOwnedQueryPolicyMixin struct {
		mixin.Schema
	}
)

// Policy sets the policy for updating owned fields by the user
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

// Policy sets the policy for querying owned fields by the user
func (mixin UserOwnedQueryPolicyMixin) Policy() ent.Policy {
	return privacy.Policy{
		Query: privacy.QueryPolicy{
			privacy.AlwaysAllowRule(),
		},
	}
}
