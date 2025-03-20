package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
)

// CustomSchema defines a custom schema that includes additional functionality
// including Name and PluralName methods
type CustomSchema struct {
	SchemaFuncs
}

// SchemaFuncs defines the methods that a custom schema must implement
// in order to use the helper functions provided by this file
type SchemaFuncs interface {
	Name() string
	PluralName() string
	GetType() any
}

// mixinConfig defines the configuration for the mixins that can be used
type mixinConfig struct {
	prefix             string
	excludeTags        bool
	includeRevision    bool
	excludeAnnotations bool
	additionalMixins   []ent.Mixin
}

// baseDefaultMixins defines the default mixins that are always included and cannot be excluded
// using the config
var baseDefaultMixins = []ent.Mixin{
	emixin.AuditMixin{},
	mixin.SoftDeleteMixin{},
}

// getMixins returns the mixins based on the configuration provided
// by default it will include :
// - AuditMixin
// - SoftDeleteMixin
// - AnnotationMixin (set excludeAnnotations to true to disable)
// - IDMixin (with or without prefix)
// - TagMixin (set excludeTags to true to disable)
// - RevisionMixin (set includeRevision to true to enable)
// - any additional mixins can  be appended using the additionalMixins field
func (m mixinConfig) getMixins() []ent.Mixin {
	mixins := baseDefaultMixins

	// always include the IDMixin, if a prefix is provided it will be used
	idMixin := emixin.IDMixin{}
	if m.prefix != "" {
		idMixin = emixin.NewIDMixinWithPrefixedID(m.prefix)
	}

	mixins = append(mixins, idMixin)

	// exclude tags if specified
	if !m.excludeTags {
		mixins = append(mixins, emixin.TagMixin{})
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
func getDefaultMixins() []ent.Mixin {
	return mixinConfig{}.getMixins()
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
	sch, ok := schema.(SchemaFuncs)
	if ok {
		return sch.PluralName()
	}

	log.Fatal().Msgf("schema must implement SchemaFuncs, %T does not", schema)

	return ""
}

// getName returns the name of the schema
func getName(schema any) string {
	sch, ok := schema.(SchemaFuncs)
	if ok {
		return sch.Name()
	}

	log.Fatal().Msgf("schema must implement SchemaFuncs, %T does not", schema)

	return ""
}

// getType returns the type of the schema
func getType(schema any) any {
	sch, ok := schema.(SchemaFuncs)
	if ok {
		return sch.GetType()
	}

	log.Fatal().Msgf("schema must implement ent.Schema, %T does not", schema)

	return nil
}

// edgeToWithPagination uses the provided edge definition to create an edge with pagination
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
// for the given edge schema
func defaultEdgeToWithPagination(from any, edgeSchema any) ent.Edge {
	return edgeToWithPagination(&edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	})
}

// defaultEdgeFromWithPagination uses the default edge definition to create an edge with pagination
// for the given edge schema
func defaultEdgeFromWithPagination(from any, edgeSchema any) ent.Edge {
	return edgeFromWithPagination(&edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	})
}

// defaultEdgeFrom uses the default edge definition to create an inverse edge for a 1:M relationship
func defaultEdgeFrom(from any, edgeSchema any) ent.Edge {
	e := &edgeDefinition{
		edgeSchema: edgeSchema,
		fromSchema: from,
	}

	e.getEdgeDetails(false)

	return basicEdgeFrom(e, false)
}

// getEdgeDetails retrieves the edge details based on the edge schema by retrieving the name and type
// and setting them on the edge definition
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

// uniqueEdgeTo uses the provided edge definition to create an unique edge
func uniqueEdgeTo(e *edgeDefinition) ent.Edge {
	e.getEdgeDetails(false)

	return basicEdgeTo(e, true)
}

// uniqueEdgeFrom uses the provided edge definition to create the inverse unique edge
func uniqueEdgeFrom(e *edgeDefinition) ent.Edge {
	e.getEdgeDetails(false)

	return basicEdgeFrom(e, true)
}

// basicEdgeToDesc uses the provided edge definition to create an edge
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

// basicEdgeFromDesc uses the provided edge definition to create the inverse edge
func basicEdgeFrom(e *edgeDefinition, unique bool) ent.Edge {
	if e.ref == "" {
		if e.fromSchema == nil {
			log.Fatal().Msg("edgeDefinition: edgeFrom must be set")
		}

		e.ref = getPluralName(e.fromSchema)
	}

	validateEdgeDefinition(e)

	edgeFrom := edge.From(e.name, e.t).Ref(e.ref)

	// edgeFrom.Descriptor().RefName = e.ref

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
