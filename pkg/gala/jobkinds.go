package gala

import "strings"

// River job kinds for every durable envelope family; the single source of truth for
// kind names across the codebase
const (
	// JobKindMutation carries mutation event envelopes
	JobKindMutation = "gala.mutation"
	// JobKindWorkflow carries workflow command envelopes
	JobKindWorkflow = "gala.workflow"
	// JobKindIntegrationRun carries one-shot integration operation envelopes
	JobKindIntegrationRun = "gala.integration.run"
	// JobKindIntegrationReconcile carries recurring reconcile and scheduled cycle envelopes
	JobKindIntegrationReconcile = "gala.integration.reconcile"
	// JobKindIntegrationIngest carries ingest persistence envelopes
	JobKindIntegrationIngest = "gala.integration.ingest"
	// JobKindIntegrationWebhook carries inbound webhook envelopes
	JobKindIntegrationWebhook = "gala.integration.webhook"
	// JobKindSystem carries startup and maintenance envelopes
	JobKindSystem = "gala.system"
)

// JobKinds returns every job kind a durable runtime registers by default
func JobKinds() []string {
	return []string{
		JobKindMutation,
		JobKindWorkflow,
		JobKindIntegrationRun,
		JobKindIntegrationReconcile,
		JobKindIntegrationIngest,
		JobKindIntegrationWebhook,
		JobKindSystem,
	}
}

// topic namespace vocabulary: every durable topic family carries one of these
// designating prefixes, making job kind resolution total over topic names
const (
	// TopicPrefixMutation designates direct mutation event topics
	TopicPrefixMutation = "mutation."
	// TopicPrefixWorkflowMutation designates workflow-concern mutation topics
	TopicPrefixWorkflowMutation = "workflow.mutation."
	// TopicPrefixNotificationMutation designates notification-concern mutation topics
	TopicPrefixNotificationMutation = "notification.mutation."
	// TopicPrefixWorkflowCommand designates workflow command topics
	TopicPrefixWorkflowCommand = "workflow.command."
	// TopicNamespaceIntegrationReconcile designates recurring reconcile and scheduled cycle topics
	TopicNamespaceIntegrationReconcile = "integration.reconcile."
	// TopicPrefixIntegration designates integration operation and webhook topics
	TopicPrefixIntegration = "integration."
	// TopicInfixIntegrationWebhook designates integration webhook event topics
	TopicInfixIntegrationWebhook = ".webhook."
	// TopicInfixIngest designates ingest persistence topics
	TopicInfixIngest = ".ingest."
	// TopicNamespaceDomainScan designates domain scan saga topics
	TopicNamespaceDomainScan = "domainscan.poll."
	// TopicPrefixSystem designates gala system maintenance topics
	TopicPrefixSystem = "system."
	// TopicSystemBackfill is the startup backfill topic
	TopicSystemBackfill = "startup.backfill"
)

// SystemTopics is the namespace for gala system maintenance topics
var SystemTopics = NewTopicNamespace(TopicPrefixSystem, JobKindSystem)

// ResolveTopicKind returns the job kind for a topic by historical topology rules; the
// runtime registry is the authority for registered topics
func ResolveTopicKind(topic TopicName) (string, bool) {
	name := string(topic)

	switch {
	case strings.HasPrefix(name, TopicPrefixMutation),
		strings.HasPrefix(name, TopicPrefixWorkflowMutation),
		strings.HasPrefix(name, TopicPrefixNotificationMutation):
		return JobKindMutation, true
	case strings.HasPrefix(name, TopicPrefixWorkflowCommand):
		return JobKindWorkflow, true
	case name == TopicSystemBackfill, strings.HasPrefix(name, TopicPrefixSystem):
		return JobKindSystem, true
	case strings.HasPrefix(name, TopicNamespaceDomainScan):
		return JobKindIntegrationRun, true
	case strings.HasPrefix(name, TopicNamespaceIntegrationReconcile):
		return JobKindIntegrationReconcile, true
	case strings.Contains(name, TopicInfixIngest):
		return JobKindIntegrationIngest, true
	case strings.HasPrefix(name, TopicPrefixIntegration) && strings.Contains(name, TopicInfixIntegrationWebhook):
		return JobKindIntegrationWebhook, true
	case strings.HasPrefix(name, TopicPrefixIntegration):
		return JobKindIntegrationRun, true
	}

	return "", false
}
