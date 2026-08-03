package hooks

import (
	"github.com/theopenlane/core/pkg/gala"
)

// GalaRuntime selects which gala runtime a registration targets
type GalaRuntime int

const (
	// GalaRuntimeMain targets the durable main gala runtime
	GalaRuntimeMain GalaRuntime = iota
	// GalaRuntimeNotification targets the in-memory notification runtime
	GalaRuntimeNotification
)

// GalaRegistration pairs a listener registration func with its target runtime
type GalaRegistration struct {
	// Runtime is the gala runtime class this registration targets
	Runtime GalaRuntime
	// Register wires the listeners into the given gala runtime
	Register func(*gala.Gala) ([]gala.ListenerID, error)
}

// GalaRegistrations is the single source of truth for hook listener registration;
// every RegisterGalaXListeners func must appear here exactly once
var GalaRegistrations = []GalaRegistration{
	{Runtime: GalaRuntimeMain, Register: RegisterGalaOrganizationAvatarListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaTaskRuleListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaEntitlementListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaTrustCenterCacheListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaTrustCenterWatermarkListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaWorkflowListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaVendorScoringListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaIdentityResolutionListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaDocumentAssociationListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaQuestionnaireTransformListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaCampaignRecurringListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaSubscriberLinkListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaNDAAttestationListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaDomainScanSubmitListeners},
	{Runtime: GalaRuntimeMain, Register: RegisterGalaDomainScanUpdateListener},
	{Runtime: GalaRuntimeNotification, Register: RegisterGalaNotificationListeners},
}
