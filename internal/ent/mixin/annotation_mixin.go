package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/mixin"
)

// AnnotationMixin implements the revision pattern for schemas.
type AnnotationMixin struct {
	mixin.Schema
}

// Annotations of the AnnotationMixin.
func (AnnotationMixin) Annotations() []schema.Annotation {
	return defaultAnnotations
}

// defaultAnnotations defines the default annotations used across the application that should generally be applied to all schemas
var defaultAnnotations = []schema.Annotation{
	entgql.RelayConnection(),
	entgql.QueryField(),
	entgql.MultiOrder(),
	entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
}
