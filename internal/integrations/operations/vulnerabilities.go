package operations

import (
	"github.com/theopenlane/core/v2/pkg/gala"
)

// VulnerabilitiesIngestedEnvelope announces that an installation's run persisted vulnerabilities, so they can be enriched
type VulnerabilitiesIngestedEnvelope struct {
	// OwnerID is the organization owning the vulnerabilities
	OwnerID string `json:"owner_id"`
	// IntegrationID is the installation whose run persisted the vulnerabilities
	IntegrationID string `json:"integration_id"`
	// DefinitionID is the definition of the installation, so provider specific enrichment can pick its own ingests
	DefinitionID string `json:"definition_id"`
}

// vulnerabilitiesIngestedTopics namespaces the vulnerability enrichment envelopes
var vulnerabilitiesIngestedTopics = gala.IntegrationRun.At("vulnerabilities.ingested")

// VulnerabilitiesIngestedTopic carries one envelope per installation run that persisted vulnerabilities, keyed so
// repeat runs of one installation collapse while an enrichment is still pending
var VulnerabilitiesIngestedTopic = gala.NamespacedTopicFor(vulnerabilitiesIngestedTopics, gala.WithUniqueKey(vulnerabilitiesIngestedKey))

// vulnerabilitiesIngestedKey dedupes pending enrichment envelopes per installation
func vulnerabilitiesIngestedKey(envelope VulnerabilitiesIngestedEnvelope) string {
	return vulnerabilitiesIngestedTopics.Key(envelope.IntegrationID)
}
