package gala

import (
	"context"
	"strings"

	"github.com/theopenlane/utils/contextx"
)

const (
	workflowMutationTopicPrefix     = "workflow.mutation."
	notificationMutationTopicPrefix = "notification.mutation."
)

// MutationConcern identifies the eventing concern namespace for mutation topics
type MutationConcern string

const (
	// MutationConcernDirect is the default concern for direct mutation listeners
	MutationConcernDirect MutationConcern = "direct"
	// MutationConcernWorkflow is the concern for workflow mutation listeners
	MutationConcernWorkflow MutationConcern = "workflow"
	// MutationConcernNotification is the concern for notification mutation listeners
	MutationConcernNotification MutationConcern = "notification"
)

const (
	// MutationPropertyEntityID is the standard mutation metadata key used for entity identifiers
	MutationPropertyEntityID = "ID"
	// MutationPropertyOperation is the mutation metadata key used for the operation type
	MutationPropertyOperation = "operation"
	// MutationPropertyMutationType is the mutation metadata key used for the ent schema type
	MutationPropertyMutationType = "mutation_type"
	// SoftDeleteOne is a synthetic operation used for soft-delete hooks
	SoftDeleteOne = "SoftDeleteOne"
)

// FlagSoftDeleteOperation marks a request context whose delete mutation is a soft delete;
// resolvers set it explicitly so mutation-event emission classifies the operation without
// inspecting transport-level request state
const FlagSoftDeleteOperation ContextFlag = "soft_delete_operation"

// MutationTopicName returns the mutation topic name for a concern + schema type pair
func MutationTopicName(concern MutationConcern, schemaType string) TopicName {
	schemaType = strings.TrimSpace(schemaType)
	if schemaType == "" {
		return ""
	}

	prefix := ""
	switch concern {
	case MutationConcernWorkflow:
		prefix = workflowMutationTopicPrefix
	case MutationConcernNotification:
		prefix = notificationMutationTopicPrefix
	}

	return TopicName(prefix + schemaType)
}

// skipEventEmissionFlag is a mutable flag allowing inner hooks to signal outward that
// mutation events should not be emitted for the current mutation
type skipEventEmissionFlag struct {
	skip bool
}

var skipEventEmissionFlagContextKey = contextx.NewKey[*skipEventEmissionFlag]()

// WithSkipEventEmission installs a mutable flag in the context so inner hooks can
// signal that mutation events should not be emitted via MarkSkipEventEmission
func WithSkipEventEmission(ctx context.Context) context.Context {
	if existing, ok := skipEventEmissionFlagContextKey.Get(ctx); ok && existing != nil {
		return ctx
	}

	return skipEventEmissionFlagContextKey.Set(ctx, &skipEventEmissionFlag{})
}

// MarkSkipEventEmission marks the context to skip emitting mutation events
func MarkSkipEventEmission(ctx context.Context) {
	if flag, ok := skipEventEmissionFlagContextKey.Get(ctx); ok && flag != nil {
		flag.skip = true
	}
}

// SkipEventEmission installs the mutable skip flag and immediately marks it, combining
// WithSkipEventEmission and MarkSkipEventEmission into a single call
func SkipEventEmission(ctx context.Context) context.Context {
	ctx = WithSkipEventEmission(ctx)
	MarkSkipEventEmission(ctx)

	return ctx
}

// ShouldSkipEventEmission reports whether the context has been marked to skip mutation events
func ShouldSkipEventEmission(ctx context.Context) bool {
	flag, ok := skipEventEmissionFlagContextKey.Get(ctx)

	return ok && flag != nil && flag.skip
}
