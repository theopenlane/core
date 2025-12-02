package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/ent/mixin"
)

const (
	minNameLength = 3
)

// SchemaFuncs defines the methods that a custom schema must implement
// in order to use the helper functions provided by this file
// including Name(), PluralName(), and GetType()
type SchemaFuncs interface { //nolint:revive
	Name() string
	PluralName() string
	GetType() any
}

// toSchemaFuncs converts the provided schema of type any to a SchemaFuncs
// in order to call the required methods
// if the schema does not implement SchemaFuncs, it will fatal error
func toSchemaFuncs(schema any) SchemaFuncs {
	sch, ok := schema.(SchemaFuncs)
	if !ok {
		log.Fatal().Msg("schema must implement SchemaFuncs")
	}

	return sch
}

// mixinConfig defines the configuration for the mixins that can be used
type mixinConfig struct {
	// prefix is the prefix to use for the IDMixin, if left empty it will use the default IDMixin
	prefix string
	// excludeTags if true, the TagMixin will be excluded
	excludeTags bool
	// excludeSoftDelete if true, the SoftDeleteMixin will be excluded
	excludeSoftDelete bool
	// includeRevision if true, the RevisionMixin will be included
	includeRevision bool
	// excludeAnnotations if true, the AnnotationMixin will be excluded
	excludeAnnotations bool
	// additionalMixins are any additional mixins that will be included
	additionalMixins []ent.Mixin
}

// baseDefaultMixins defines the default mixins that are always included and cannot be excluded
// using the config
var baseDefaultMixins = []ent.Mixin{
	// audit mixin includes created_at, updated_at, and created_by
	emixin.AuditMixin{},
}

// getMixins returns the mixins based on the configuration provided
// by default it will include :
// - AuditMixin
// - AutoHushEncryptionMixin (automatically detects hush.EncryptField() annotations)
// - SoftDeleteMixin
// - AnnotationMixin (set excludeAnnotations to true to disable)
// - IDMixin (with or without prefix)
// - TagMixin (set excludeTags to true to disable)
// - RevisionMixin (set includeRevision to true to enable)
// - any additional mixins can  be appended using the additionalMixins field
func (m mixinConfig) getMixins(schema ent.Interface) []ent.Mixin {
	// Start with base mixins and add auto-encryption using the passed schema
	mixins := append([]ent.Mixin{}, baseDefaultMixins...)
	mixins = append(mixins, NewAutoHushEncryptionMixin(schema))

	if !m.excludeSoftDelete {
		mixins = append(mixins, mixin.SoftDeleteMixin{})
	} else {
		for i, mixin := range m.additionalMixins {
			if _, ok := mixin.(ObjectOwnedMixin); ok {
				// if the ObjectOwnedMixin is present, set the SkipDeletedAt field
				if o, ok := mixin.(ObjectOwnedMixin); ok {
					o.SkipDeletedAt = true
					m.additionalMixins[i] = o
				}
			}
		}
	}

	// always include the IDMixin, if a prefix is provided it will be used
	idMixin := emixin.IDMixin{}
	if m.prefix != "" {
		idMixin = emixin.NewIDMixinWithPrefixedID(m.prefix)
	}

	if autoSetSkipForSystemAdmin(&m) {
		// if both SystemOwnedMixin and ObjectOwnedMixin are present, set skip for system admin to true
		for i, mixin := range m.additionalMixins {
			if o, ok := mixin.(ObjectOwnedMixin); ok {
				o.AllowEmptyForSystemAdmin = true
				m.additionalMixins[i] = o

				break
			}
		}
	}

	mixins = append(mixins, idMixin)

	// exclude tags if specified
	if !m.excludeTags {
		mixins = append(mixins, mixin.TagMixin{})
	}

	// include revision if specified
	if m.includeRevision {
		mixins = append(mixins, mixin.RevisionMixin{})
	}

	// exclude annotations if specified
	if !m.excludeAnnotations {
		mixins = append(mixins, mixin.GraphQLAnnotationMixin{})
	}

	return append(mixins, m.additionalMixins...)
}

// getDefaultMixins uses the default mixin configuration to return the default mixins
// for the given schema
func getDefaultMixins(schema ent.Interface) []ent.Mixin {
	return mixinConfig{}.getMixins(schema)
}

// edgeDefinition defines the edge schema and its properties
type edgeDefinition struct {
	// edgeSchema is the schema of the edge
	edgeSchema any
	// fromSchema is the schema the edge is coming from
	fromSchema any
	// name of the edge schema, can be used with t to override the default edgeSchema name
	name string
	// t is the type of the edge, can be used with name to override the default edgeSchema type
	t any
	// field is the field name of the unique edge for foreign keys
	field string
	// ref is the reference on the From (inverse) edge, if not set it will be derived from the fromSchema
	ref string
	// annotations are the annotations for the edge in addition to any defaults
	annotations []schema.Annotation
	// comment is the comment for the edge
	comment string
	// required indicates if the edge is required
	required bool
	// immutable indicates if the edge is immutable
	immutable bool
	// cascadeDelete indicates if the edge should cascade delete based on the schema name
	cascadeDelete string
	// cascadeDeleteOwner indicates if the edge should cascade delete based the Owner field
	cascadeDeleteOwner bool
}

// getPluralName returns the plural name of the schema
func getPluralName(schema any) string {
	sch := toSchemaFuncs(schema)

	return sch.PluralName()
}

// getName returns the name of the schema
func getName(schema any) string {
	sch := toSchemaFuncs(schema)

	return sch.Name()
}

// getType returns the type of the schema
func getType(schema any) any {
	sch := toSchemaFuncs(schema)

	return sch.GetType()
}

// edgeToWithPagination uses the provided edge definition to create an edge with pagination
// for the given edge schema, this should be used when there will be a 1:M relationship or M:M relationship
// to the edge schema
func edgeToWithPagination(e *edgeDefinition) ent.Edge {
	defaultAnnotations := []schema.Annotation{entgql.RelayConnection()}
	if e.annotations != nil {
		e.annotations = append(defaultAnnotations, e.annotations...)
	} else {
		e.annotations = defaultAnnotations
	}

	e.getEdgeDetails(true)

	return basicEdgeTo(e, false)
}

// edgeFromWithPagination uses the provided edge definition to create an inverse edge with pagination
// this should be used when there will be a M:M relationship, if its a 1:M relationship use defaultEdgeFrom
// for a single edge
func edgeFromWithPagination(e *edgeDefinition) ent.Edge {
	defaultAnnotations := []schema.Annotation{entgql.RelayConnection()}

	if e.annotations != nil {
		e.annotations = append(defaultAnnotations, e.annotations...)
	} else {
		e.annotations = defaultAnnotations
	}

	e.getEdgeDetails(true)

	return basicEdgeFrom(e, false)
}

// defaultEdgeToWithPagination uses the default edge definition to create an edge with pagination
// for the given edge schema, to be used for a 1:M relationship or M:M relationship using all default
// settings
func defaultEdgeToWithPagination(from any, edgeSchema any) ent.Edge {
	return edgeToWithPagination(&edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	})
}

// defaultEdgeFromWithPagination uses the default edge definition to create an edge with pagination
// for the given edge schema, to be used for a M:M relationship using all default settings
func defaultEdgeFromWithPagination(from any, edgeSchema any) ent.Edge {
	return edgeFromWithPagination(&edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	})
}

// defaultEdgeFrom uses the default edge definition to create an inverse edge for a 1:M relationship
// for the given edge schema, using all default settings (the name will be singular vs the plural used in edgeTo inverse)
// it will not include the pagination annotations because of the singular side of the relationship
func defaultEdgeFrom(from any, edgeSchema any) ent.Edge {
	e := &edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	}

	e.getEdgeDetails(false)

	return basicEdgeFrom(e, false)
}

// getEdgeDetails retrieves the edge details based on the edge schema by retrieving the name and type
// and setting them on the edge definition including the name (singular or plural based on the bool)
// and type of the edge
func (e *edgeDefinition) getEdgeDetails(plural bool) {
	if e.edgeSchema == nil {
		return
	}

	switch {
	case plural:
		e.name = getPluralName(e.edgeSchema)
	default:
		e.name = getName(e.edgeSchema)
	}

	e.t = getType(e.edgeSchema)
}

// uniqueEdgeTo uses the provided edge definition to create an unique edge for 1:1 relationships
func uniqueEdgeTo(e *edgeDefinition) ent.Edge {
	e.getEdgeDetails(false)

	return basicEdgeTo(e, true)
}

// uniqueEdgeFrom uses the provided edge definition to create the inverse unique edge for 1:1 relationships
func uniqueEdgeFrom(e *edgeDefinition) ent.Edge {
	e.getEdgeDetails(false)

	return basicEdgeFrom(e, true)
}

// nonUniqueEdgeFrom uses the provided edge definition to create the inverse edge without unique constraint
// This is useful when you want to define your own unique constraint with soft delete support
func nonUniqueEdgeFrom(e *edgeDefinition) ent.Edge {
	e.getEdgeDetails(false)

	return basicEdgeFrom(e, false)
}

// basicEdgeTo uses the provided edge definition to create an edge with all the properties
// this is used by the above functions, and is not intended to be called directly
func basicEdgeTo(e *edgeDefinition, unique bool) ent.Edge {
	validateEdgeDefinition(e)

	edgeTo := edge.To(e.name, e.t)

	if unique {
		edgeTo = edgeTo.Unique()
	}

	if e.required {
		edgeTo = edgeTo.Required()
	}

	if e.immutable {
		edgeTo = edgeTo.Immutable()
	}

	if e.field != "" {
		edgeTo = edgeTo.Field(e.field)
	}

	annotations := e.annotations

	if e.cascadeDelete != "" || e.cascadeDeleteOwner {
		name := e.cascadeDelete
		if e.cascadeDeleteOwner {
			name = "Owner"
		}

		annotations = append(annotations, entx.CascadeAnnotationField(name))
	}

	if len(annotations) > 0 {
		edgeTo = edgeTo.Annotations(annotations...)
	}

	if e.comment != "" {
		edgeTo = edgeTo.Comment(e.comment)
	}

	return edgeTo
}

// basicEdgeFromDesc uses the provided edge definition to create the inverse edge for a 1:1 relationship
// this is used by the above functions, and is not intended to be called directly
func basicEdgeFrom(e *edgeDefinition, unique bool) ent.Edge {
	if e.ref == "" {
		if e.fromSchema == nil {
			log.Fatal().Msg("edgeDefinition: fromSchema must be set")
		}

		e.ref = getPluralName(e.fromSchema)
	}

	validateEdgeDefinition(e)

	edgeFrom := edge.From(e.name, e.t).Ref(e.ref)

	if unique {
		edgeFrom.Unique()
	}

	if e.required {
		edgeFrom.Required()
	}

	if e.immutable {
		edgeFrom.Immutable()
	}

	if e.field != "" {
		edgeFrom.Field(e.field)
	}

	if len(e.annotations) > 0 {
		edgeFrom.Annotations(e.annotations...)
	}

	if e.comment != "" {
		edgeFrom.Comment(e.comment)
	}

	return edgeFrom
}

// validateEdgeDefinition validates the edge definition to ensure that the name and type are set
// this is called after the name and type have been derived from the edge schema
func validateEdgeDefinition(e *edgeDefinition) {
	if e.name == "" || e.t == nil {
		log.Fatal().Str("schema", getName(e.fromSchema)).Msg("edge_definition: name and type must be set")
	}
}

// autoSetSkipForSystemAdmin checks if both SystemOwnedMixin and ObjectOwnedMixin are present in the mixinConfig
// if both are present it returns true to indicate that the ObjectOwnedMixin should have its AllowEmptyForSystemAdmin
// field set to true to allow system admins to bypass the organization ownership requirement
func autoSetSkipForSystemAdmin(mixinConfig *mixinConfig) bool {
	hasSystemOwnedMixin := false

	for _, m := range mixinConfig.additionalMixins {
		if so, ok := m.(mixin.SystemOwnedMixin); ok {
			// ensure its actually the SystemOwnedMixin by checking the name
			// because the mixin doesn't have any distinguishing fields or methods
			if so.Name() != mixin.SystemOwnedMixinName {
				continue
			}

			hasSystemOwnedMixin = true

			break
		}
	}

	hasObjectOwnedMixin := false

	for _, m := range mixinConfig.additionalMixins {
		if _, ok := m.(ObjectOwnedMixin); ok {
			hasObjectOwnedMixin = true

			break
		}
	}

	return hasSystemOwnedMixin && hasObjectOwnedMixin
}
