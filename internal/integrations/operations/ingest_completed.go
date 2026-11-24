package operations

import (
	"github.com/theopenlane/core/v2/pkg/gala"
)

// IngestCompletedEnvelope announces that an installation's run finished ingesting, so listeners can act on the schemas it wrote
type IngestCompletedEnvelope struct {
	// OwnerID is the organization owning the ingested records
	OwnerID string `json:"owner_id"`
	// IntegrationID is the installation whose run ingested the records
	IntegrationID string `json:"integration_id"`
	// DefinitionID is the definition of the installation
	DefinitionID string `json:"definition_id"`
	// Operation is the operation that ran
	Operation string `json:"operation"`
	// Schemas are the schemas the operation ingests
	Schemas []string `json:"schemas"`
	// Persisted counts the records the run created, changed or left unchanged
	Persisted int `json:"persisted"`
	// Changed counts the records the run created or modified
	Changed int `json:"changed"`
}

// ingestCompletedTopics namespaces the ingest completed envelopes
var ingestCompletedTopics = gala.IntegrationRun.At("ingest.completed")

// IngestCompletedTopic carries one envelope per installation run that persisted records, keyed so repeat runs of one
// operation collapse while a listener is still pending
var IngestCompletedTopic = gala.NamespacedTopicFor(ingestCompletedTopics, gala.WithUniqueKey(ingestCompletedKey))

// ingestCompletedKey dedupes pending ingest completed envelopes per installation operation
func ingestCompletedKey(envelope IngestCompletedEnvelope) string {
	return ingestCompletedTopics.Key(envelope.IntegrationID, envelope.Operation)
}
