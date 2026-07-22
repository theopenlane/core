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
		Expression: "value.existing_controls == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportTemplateControls,
		Expression: "!(has(value.existing_controls) && value.existing_controls == true)",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportExistingPolicies,
		Expression: "value.existing_policies_procedures == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleImportPolicyTemplates,
		Expression: "!(has(value.existing_policies_procedures) && value.existing_policies_procedures == true)",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleHasAuditorAtOnboarding,
		Expression: "value.has_auditor == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleWantsAuditorRecommendation,
		Expression: "value.recommend_auditors == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:     RuleWantsPartnerRecommendation,
		Expression: "value.recommend_vciso_partner == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}

// OnboardingDemoRequestedRule fires directly off the sibling demo_requested field --
// it lives outside OnboardingComplianceRules because "compliance" and "demo_requested"
// are separate ent fields
var OnboardingDemoRequestedRule = []entx.TaskRuleSpec{
	{
		RuleID:     RuleDemoRequested,
		Expression: "value == true",
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}
