package policy

import (
	"entgo.io/ent"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
)

// prePolicy is executed before privacy policy
var prePolicy = privacy.Policy{
	Query: privacy.QueryPolicy{
		// allow internal requests (used in tests) to proceed to query tables
		rule.AllowIfInternalRequest(),
		// allow history requests to proceed to query tables
		history.AllowIfHistoryRequest(),
	},
	Mutation: privacy.MutationPolicy{
		// allow internal requests (used in tests) to proceed to mutate tables
		rule.AllowIfInternalRequest(),
		rule.DenyIfMissingAllModules(),
	},
}

// postPolicy is executed after privacy policy
var postPolicy = privacy.Policy{
	Query: privacy.QueryPolicy{
		privacy.AlwaysAllowRule(),
	},
	Mutation: privacy.MutationPolicy{
		privacy.AlwaysDenyRule(),
	},
}

// Option configures policy creation.
type Option func(*policies)

// policies aggregate policy options.
type policies struct {
	query     privacy.QueryPolicy
	mutation  privacy.MutationPolicy
	pre, post privacy.Policy
}

// WithQueryRules adds query rules to policy.
func WithQueryRules(rules ...privacy.QueryRule) Option {
	return func(policies *policies) {
		policies.query = append(policies.query, rules...)
	}
}

// WithMutationRules adds mutation rules to policy.
func WithMutationRules(rules ...privacy.MutationRule) Option {
	return func(policies *policies) {
		policies.mutation = append(policies.mutation, rules...)
	}
}

// WithOnMutationRules adds mutation rules to policy for specific operations.
func WithOnMutationRules(op ent.Op, rules ...privacy.MutationRule) Option {
	opRules := []privacy.MutationRule{}

	for _, rule := range rules {
		r := privacy.OnMutationOperation(
			rule,
			op,
		)

		opRules = append(opRules, r)
	}

	return func(policies *policies) {
		policies.mutation = append(policies.mutation, opRules...)
	}
}

// WithPrePolicy overrides the pre-policy to be executed.
func WithPrePolicy(policy privacy.Policy) Option {
	return func(policies *policies) {
		policies.pre = policy
	}
}

// WithPostPolicy overrides the post-policy to be executed.
func WithPostPolicy(policy privacy.Policy) Option {
	return func(policies *policies) {
		policies.post = policy
	}
}

// NewPolicy creates a privacy policy.
func NewPolicy(opts ...Option) ent.Policy {
	policies := policies{
		pre:  prePolicy,
		post: postPolicy,
	}

	for _, opt := range opts {
		opt(&policies)
	}

	return privacy.Policy{
		Query:    policies.queryPolicy(),
		Mutation: policies.mutationPolicy(),
	}
}

func (p policies) queryPolicy() privacy.QueryPolicy {
	policy := append(privacy.QueryPolicy(nil), p.pre.Query...)
	policy = append(policy, p.query...)
	policy = append(policy, p.post.Query...)

	return policy
}

func (p policies) mutationPolicy() privacy.MutationPolicy {
	policy := append(privacy.MutationPolicy(nil), p.pre.Mutation...)
	policy = append(policy, p.mutation...)
	policy = append(policy, p.post.Mutation...)

	return policy
}
