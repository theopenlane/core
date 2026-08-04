package taskrules

import "github.com/theopenlane/entx"

// Organization suggested-task rule IDs
const (
	RuleSecureOrganization   = "suggested-secure-organization"
	RuleInviteTeam           = "suggested-invite-team"
	RuleCreateGroups         = "suggested-create-groups"
	RuleCompleteRegistry     = "suggested-complete-registry"
	RuleConfigureTrustCenter = "suggested-configure-trust-center"
	RuleAddPaymentMethod     = "suggested-add-payment-method"
	RuleSetupIntegrations    = "suggested-setup-integrations"
)

// notPersonalOrg gates every OrganizationSuggestedRules entry so personal orgs never get these suggestions
const notPersonalOrg = "!has(value.personal_org) || value.personal_org == false"

// OrganizationSuggestedRules fire once when a non-personal organization is created
var OrganizationSuggestedRules = []entx.TaskRuleSpec{
	{
		RuleID:     RuleSecureOrganization,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleCreateGroups,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleInviteTeam,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleSetupIntegrations,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleAddPaymentMethod,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleCompleteRegistry,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleConfigureTrustCenter,
		Expression: notPersonalOrg,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}
