package policy

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
)

// getPrePolicies returns the pre-policies which are executed before privacy policy
func getPrePolicies(skipDenyOrgRule bool, beforeScope privacy.MutationPolicy) privacy.Policy {
	// prePolicy is executed before privacy policy
	base := privacy.Policy{
		Query: privacy.QueryPolicy{
			// allow internal operations and reads (CapInternalOperation or CapInternalRead) to proceed to query tables
			rule.AllowIfInternalReadRequest(),
			// allow history requests to proceed to query tables
			history.AllowIfHistoryRequest(),
		},
		Mutation: privacy.MutationPolicy{
			// allow internal operations (system code paths with CapInternalOperation) to proceed to mutate tables
			rule.AllowIfInternalRequest(),
			// deny mutation if missing all modules
			rule.DenyIfMissingAllModules(),
		},
	}

	if !skipDenyOrgRule {
		base.Mutation = append(base.Mutation,
			// deny if the user doesn't have access to the organization
			rule.DenyIfNotInOrganization(),
		)
	}

	base.Mutation = append(base.Mutation, beforeScope...)

	base.Mutation = append(base.Mutation,
		// allow mutation if the api token has the appropriate mutation scope
		rule.AllowIfTokenHasMutationScope(),
	)

	return base
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
	query        privacy.QueryPolicy
	mutation     privacy.MutationPolicy
	beforeScope  privacy.MutationPolicy
	pre, post    privacy.Policy
	skipDenyRule bool
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
	opRules := onMutationOperation(op, rules)

	return func(policies *policies) {
		policies.mutation = append(policies.mutation, opRules...)
	}
}

// WithMutationRulesBeforeScope adds mutation rules that run before the mutation scope allow rule
func WithMutationRulesBeforeScope(rules ...privacy.MutationRule) Option {
	return func(policies *policies) {
		policies.beforeScope = append(policies.beforeScope, rules...)
	}
}

// WithOnMutationRulesBeforeScope adds mutation rules for specific operations that run before the mutation scope allow rule
func WithOnMutationRulesBeforeScope(op ent.Op, rules ...privacy.MutationRule) Option {
	opRules := onMutationOperation(op, rules)

	return func(policies *policies) {
		policies.beforeScope = append(policies.beforeScope, opRules...)
	}
}

// onMutationOperation wraps each rule so it only runs for the given operation
func onMutationOperation(op ent.Op, rules []privacy.MutationRule) []privacy.MutationRule {
	opRules := []privacy.MutationRule{}

	for _, rule := range rules {
		r := privacy.OnMutationOperation(
			rule,
			op,
		)

		opRules = append(opRules, r)
	}

	return opRules
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

// WithSkipDenyOrganizationRule will skip the default Deny organization filter, this should only be set on non-org owned schemas or schemas external users do not have access to
func WithSkipDenyOrganizationRule() Option {
	return func(policies *policies) {
		policies.skipDenyRule = true
	}
}

// NewPolicy creates a privacy policy.
func NewPolicy(opts ...Option) ent.Policy {
	policies := policies{
		post: postPolicy,
	}

	for _, opt := range opts {
		opt(&policies)
	}

	policies.pre = getPrePolicies(policies.skipDenyRule, policies.beforeScope)

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
