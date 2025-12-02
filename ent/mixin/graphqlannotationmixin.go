package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/mixin"
)

// GraphQLAnnotationMixin implements the revision pattern for schemas.
// it will add default annotations to the schema used for most schemas that have full
// graphql support
type GraphQLAnnotationMixin struct {
	mixin.Schema
}

// Annotations of the AnnotationMixin.
func (GraphQLAnnotationMixin) Annotations() []schema.Annotation {
	return defaultAnnotations
}

// defaultAnnotations defines the default annotations used across the application that should generally be applied to all schemas
var defaultAnnotations = []schema.Annotation{
	entgql.RelayConnection(),
	entgql.QueryField(),
	entgql.MultiOrder(),
	entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
}
