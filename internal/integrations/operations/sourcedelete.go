package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// DeleteSource selects the rows the dispatching definition wrote into the organization under one source key, from one
// source instance, or both
type DeleteSource struct {
	// SourceKey is the source key runtime ingest wrote the rows under
	SourceKey string `json:"sourceKey,omitempty" jsonschema:"title=Source Key,description=Source key the rows were written under"`
	// InstanceID is the source instance an installation wrote the rows from
	InstanceID string `json:"instanceId,omitempty" jsonschema:"title=Instance ID,description=Source instance the rows were written from"`
	// Order lists schemas in ingest order, deleted in reverse so dependents go first, every deletable schema when empty
	Order []string `json:"order,omitempty" jsonschema:"title=Order,description=Schemas in ingest order, every deletable schema when empty"`
	// Keep lists the schemas whose rows are kept
	Keep []string `json:"keep,omitempty" jsonschema:"title=Keep,description=Schemas whose rows are kept"`
	// Subject names what is removed in the outcome notification
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject,description=Names what is removed in the notification"`
}

// DeleteSourceRecordsResult reports the rows removed per schema
type DeleteSourceRecordsResult struct {
	// Deleted is the number of rows removed per schema
	Deleted map[string]int `json:"deleted"`
}

// defaultDeleteSubject names the removal in the notification when the dispatch does not
const defaultDeleteSubject = "Integration data"

// DeleteSourceOp is the operation ref for removing the rows one source key or instance wrote, registered by any definition
var DeleteSourceOp = types.OperationPayloadOf[DeleteSource]().
	HandlesRequest(RunSourceDelete).
	Policy(types.ExecutionPolicy{NoRetry: true}).
	CustomerSelectable(false).
	Internal().
	Description("Remove the records one source key or instance wrote into the organization the operation is dispatched for")

// RunSourceDelete removes the source's rows for the dispatching definition and notifies the organization, for any operation that resolves the source to delete itself
func RunSourceDelete(ctx context.Context, request types.OperationRequest, cfg DeleteSource) (json.RawMessage, error) {
	oc, _ := gala.OperationContextFromContext(ctx)

	definitionID := types.IntegrationSourceFrom(oc).DefinitionID
	if request.Integration != nil {
		definitionID = request.Integration.DefinitionID
	}

	// the delete is scoped by the caller's organization, so never run it without one
	if _, err := utils.ResolveOwnerID(ctx, ""); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSourceOwnerRequired, err)
	}

	filter := entityops.SourceFilter{DefinitionID: definitionID, InstanceID: cfg.InstanceID, Key: cfg.SourceKey}

	deleted, deleteErr := deleteSourceRecords(ctx, request.DB, filter, cfg.Order, cfg.Keep)

	notifyErr := NotifyDeleteOutcome(ctx, lo.CoalesceOrEmpty(cfg.Subject, defaultDeleteSubject), deleted, deleteErr)
	if deleteErr != nil {
		return nil, errors.Join(deleteErr, notifyErr)
	}

	if notifyErr != nil {
		return nil, notifyErr
	}

	return providerkit.EncodeResult(DeleteSourceRecordsResult{Deleted: deleted}, ErrResultEncode)
}

// deleteSourceRecords deletes the rows the filter's source wrote for the schemas in ingest order, leaving the kept
// schemas untouched, and returns the deleted count per schema, including on error
// schemas are deleted in reverse ingest order so dependents go before the records they link to
func deleteSourceRecords(ctx context.Context, db *ent.Client, filter entityops.SourceFilter, order, keep []string) (map[string]int, error) {
	deleted := map[string]int{}

	if filter.DefinitionID == "" || (filter.InstanceID == "" && filter.Key == "") {
		return deleted, ErrSourceFilterRequired
	}

	schemas, err := sourceDeleteOrder(order, keep)
	if err != nil {
		return deleted, err
	}

	for _, name := range schemas {
		schema, ok := entityops.LookupSchema(name)
		if !ok {
			return deleted, fmt.Errorf("%w: %s", ErrIngestSchemaNotFound, name)
		}

		if schema.DeleteBySource == nil {
			return deleted, fmt.Errorf("%w: %s", ErrSourceSchemaNotDeletable, name)
		}

		ids, err := schema.DeleteBySource(ctx, db, filter, ent.EdgeCleanupFor(schema.Name))
		deleted[name] = len(ids)

		if err != nil {
			return deleted, fmt.Errorf("%w: %s: %w", ErrSourceDeleteFailed, name, err)
		}
	}

	return deleted, nil
}

// sourceDeleteOrder returns the schemas to delete in reverse ingest order without the kept ones, every deletable schema
// when no order is given, rejecting a kept schema that is not in the order, which would otherwise silently delete everything
func sourceDeleteOrder(order, keep []string) ([]string, error) {
	if len(order) == 0 {
		order = deletableSchemas()
	}

	if unknown, _ := lo.Difference(keep, order); len(unknown) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrSourceKeepUnknown, unknown)
	}

	names := lo.Without(order, keep...)
	slices.Reverse(names)

	return names, nil
}

// deletableSchemas returns the snake_case names of every schema that supports deleting by source, each after the
// schemas its foreign keys point at, so the reversed delete removes referencing rows before the rows they reference
func deletableSchemas() []string {
	deletable := lo.Filter(entityops.AllSchemas(), func(schema *entityops.Schema, _ int) bool {
		return schema.DeleteBySource != nil
	})

	sorter := dependencySorter{
		included: lo.SliceToMap(deletable, func(schema *entityops.Schema) (*entityops.Schema, bool) { return schema, true }),
		placed:   map[*entityops.Schema]bool{},
		visiting: map[*entityops.Schema]bool{},
	}

	for _, schema := range deletable {
		sorter.visit(schema)
	}

	return lo.Map(sorter.ordered, func(schema *entityops.Schema, _ int) string { return schema.Snake })
}

// dependencySorter orders schemas after the schemas their foreign key edges point at, keeping registry order among
// independent schemas and breaking cycles at the schema reached first
type dependencySorter struct {
	// included are the schemas being ordered
	included map[*entityops.Schema]bool
	// placed are the schemas already in the order
	placed map[*entityops.Schema]bool
	// visiting are the schemas on the current path, a cycle stops at them
	visiting map[*entityops.Schema]bool
	// ordered is the dependency order built so far
	ordered []*entityops.Schema
}

// visit places schema after every included schema its foreign key edges point at
func (d *dependencySorter) visit(schema *entityops.Schema) {
	if d.placed[schema] || d.visiting[schema] {
		return
	}

	d.visiting[schema] = true

	for _, edge := range schema.Edges {
		if edge.Field != "" && edge.Target != schema && d.included[edge.Target] {
			d.visit(edge.Target)
		}
	}

	d.visiting[schema] = false
	d.placed[schema] = true
	d.ordered = append(d.ordered, schema)
}
