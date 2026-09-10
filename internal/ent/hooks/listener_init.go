package hooks

import "github.com/theopenlane/core/v2/pkg/gala"

func init() {
	registerListeners(
		CampaignRecurringListeners,
		DocumentAssociationListeners,
		EntitlementListeners,
		FileBackupListeners,
		IdentityResolutionListeners,
		IntegrationCleanupListeners,
		NDAAttestationListeners,
		OnboardingProgramListeners,
		OrganizationCleanupListeners,
		QuestionnaireTransformListeners,
		DomainScanListeners,
		SubscriberLinkListeners,
		TaskRuleListeners,
		TrustCenterCacheListeners,
		TrustCenterWatermarkListeners,
		VendorScoringListeners,
		WorkflowListeners,
		func() []gala.Registration { return OrganizationAvatarListeners() },
	)
}
