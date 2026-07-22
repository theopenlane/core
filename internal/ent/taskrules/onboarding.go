package taskrules

import "github.com/theopenlane/entx"

// Onboarding compliance rule IDs
const (
	RuleFramework                  = "framework"
	RuleFrameworkGeneric           = "framework-generic"
	RuleImportExistingControls     = "import-existing-controls"
	RuleImportTemplateControls     = "import-template-controls"
	RuleImportExistingPolicies     = "import-existing-policies"
	RuleImportPolicyTemplates      = "import-policy-templates"
	RuleHasAuditorAtOnboarding     = "has-auditor-at-onboarding"
	RuleWantsAuditorRecommendation = "wants-auditor-recommendation"
	RuleWantsPartnerRecommendation = "wants-partner-recommendation"
	RuleDemoRequested              = "demo-requested"
)

// OnboardingComplianceRules generate suggested tasks from the onboarding compliance answers
var OnboardingComplianceRules = []entx.TaskRuleSpec{
	{
		RuleID:      RuleFramework,
		EachElement: "value.frameworks",
		Trigger:     entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleFrameworkGeneric,
		Expression: "!(has(value.frameworks) && size(value.frameworks) > 0)",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportExistingControls,
		Expression: "value.has_existing_controls == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportTemplateControls,
		Expression: "!(has(value.has_existing_controls) && value.has_existing_controls == true)",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportExistingPolicies,
		Expression: "value.has_existing_policies == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportPolicyTemplates,
		Expression: "!(has(value.has_existing_policies) && value.has_existing_policies == true)",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleHasAuditorAtOnboarding,
		Expression: "value.auditor_status == 'yes'",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleWantsAuditorRecommendation,
		Expression: "value.auditor_status == 'recommendations'",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleWantsPartnerRecommendation,
		Expression: "value.vciso_preference == 'connect_vciso_partner'",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID: RuleDemoRequested,
		Expression: "value.demo_requested == true || (" +
			"!(has(value.auditor_status) && (value.auditor_status == 'yes' || value.auditor_status == 'recommendations')) && " +
			"!(has(value.vciso_preference) && value.vciso_preference == 'connect_vciso_partner'))",
		Trigger: entx.TaskRuleOnCreateOnly,
	},
}
