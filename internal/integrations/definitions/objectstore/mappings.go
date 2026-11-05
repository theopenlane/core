package objectstore

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// recordExternalIDField is the match key records link to each other through, the provenance mixin's external id column
	recordExternalIDField = "external_id"
	// recordEntityTypeKey is the payload key naming the entity type an entity is linked to, the vendor type when unset
	recordEntityTypeKey = "entity_type_name"
	// recordControlRefCodesKey is the payload key listing the ref codes of the controls a record links to
	recordControlRefCodesKey = "control_ref_codes"
	// recordFromControlRefCodesKey is the payload key listing the ref codes of the controls a mapped control maps from
	recordFromControlRefCodesKey = "from_control_ref_codes"
	// recordToControlRefCodesKey is the payload key listing the ref codes of the controls a mapped control maps to
	recordToControlRefCodesKey = "to_control_ref_codes"
)

// recordExternalIDExpr reads a record's external id, empty when the record has none
var recordExternalIDExpr = fmt.Sprintf("payload.?%s.orValue(\"\")", recordExternalIDField)

// recordFilterExpr skips records without an external id, which ingest would otherwise match to existing rows by name,
// ref code, or email
var recordFilterExpr = recordExternalIDExpr + ` != ""`

// recordEntityExpr links an entity to the vendor entity type when the record names none
var recordEntityExpr = "merge(payload, " + providerkit.CelMapExpr(entityops.MappingEntry{
	Key:  recordEntityTypeKey,
	Expr: fmt.Sprintf("payload.?%s.orValue(%q)", recordEntityTypeKey, vendorEntityTypeName),
}) + ")"

// recordMapExprs are the map expressions per schema, schemas without one pass through unchanged
var recordMapExprs = map[string]string{
	entityops.SchemaEntity.Name: recordEntityExpr,
}

// recordSchemas are the schemas the import may ingest, every schema annotated for integration ingest
var recordSchemas = lo.Filter(entityops.AllSchemas(), func(schema *entityops.Schema, _ int) bool {
	return schema.Ingest != nil
})

// recordLinks are the control ref code links each schema resolves beyond the edge external id and name links every
// schema gets, naming the edge only where the source reaches the controls through more than one
var recordLinks = map[string][]types.LinkRule{
	entityops.SchemaProgram.Name: {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaMappedControl.Name: {
		recordThrough(mappedcontrol.EdgeFromControls, recordRefCodeLink(recordFromControlRefCodesKey)),
		recordThrough(mappedcontrol.EdgeToControls, recordRefCodeLink(recordToControlRefCodesKey)),
	},
	entityops.SchemaControlObjective.Name:      {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaControlImplementation.Name: {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaInternalPolicy.Name:        {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaProcedure.Name:             {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaTask.Name:                  {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaEvidence.Name:              {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaRemediation.Name:           {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaActionPlan.Name:            {recordRefCodeLink(recordControlRefCodesKey)},
	entityops.SchemaReview.Name:                {recordRefCodeLink(recordControlRefCodesKey)},
}

// recordThrough names the edge a link goes through, for a source that reaches the link target through more than one edge
func recordThrough(edge string, link types.LinkRule) types.LinkRule {
	link.Edge = edge

	return link
}

// recordRefCodeLink links to the controls whose ref code is listed under key
func recordRefCodeLink(key string) types.LinkRule {
	return types.LinkRule{TargetSchema: entityops.SchemaControl.Name, TargetField: control.FieldRefCode, SourceList: key, PayloadKeys: true}
}
