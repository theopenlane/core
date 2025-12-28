// Package workflows provides minimal registry types for workflow ent templates
package workflows

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// NOTE: This file contains stub placeholders for workflow functionality which will be added in the future

// Object captures the workflow target and its entity
type Object struct {
	ID   string
	Type enums.WorkflowObjectType
	Node any
}

// ObjectRefResolver maps a WorkflowObjectRef to a workflow Object
type ObjectRefResolver func(*generated.WorkflowObjectRef) (*Object, bool)

var objectRefResolvers []ObjectRefResolver

// RegisterObjectRefResolver registers a WorkflowObjectRef resolver
func RegisterObjectRefResolver(resolver ObjectRefResolver) {
	objectRefResolvers = append(objectRefResolvers, resolver)
}

// ObjectRefQueryBuilder builds a WorkflowObjectRef query for a workflow Object
type ObjectRefQueryBuilder func(*generated.WorkflowObjectRefQuery, *Object) (*generated.WorkflowObjectRefQuery, bool)

var objectRefQueryBuilders []ObjectRefQueryBuilder

// RegisterObjectRefQueryBuilder registers a WorkflowObjectRef query builder
func RegisterObjectRefQueryBuilder(builder ObjectRefQueryBuilder) {
	objectRefQueryBuilders = append(objectRefQueryBuilders, builder)
}

// CELContextBuilder builds CEL activation variables for a workflow Object
type CELContextBuilder func(obj *Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string) map[string]any

var celContextBuilders []CELContextBuilder

// RegisterCELContextBuilder registers a CEL context builder
func RegisterCELContextBuilder(builder CELContextBuilder) {
	celContextBuilders = append(celContextBuilders, builder)
}

// AssignmentContextBuilder builds runtime assignment context for CEL evaluation
type AssignmentContextBuilder func(ctx context.Context, client *generated.Client, instanceID string) (map[string]any, error)

var assignmentContextBuilder AssignmentContextBuilder

// RegisterAssignmentContextBuilder registers the assignment context builder
func RegisterAssignmentContextBuilder(builder AssignmentContextBuilder) {
	assignmentContextBuilder = builder
}

// GetAssignmentContextBuilder returns the registered assignment context builder, if any.
func GetAssignmentContextBuilder() AssignmentContextBuilder {
	return assignmentContextBuilder
}
