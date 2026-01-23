package schema

import (
	"fmt"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/models"
)

// ResponsibilityMixin provides ownership, assignment, review, and delegation fields for schemas.
type ResponsibilityMixin struct {
	mixin.Schema

	schemaType any

	includeInternalOwner  bool
	includeBusinessOwner  bool
	includeTechnicalOwner bool
	includeSecurityOwner  bool
	includeAssignedTo     bool
	includeReviewedBy     bool
	includeLastReviewedAt bool
	includeDelegate       bool
	reviewedByOrderField  bool
	assignedToOrderField  bool
	accessCheckAnnotation func(string) schema.Annotation
}

// responsibilityOption configures ResponsibilityMixin behavior.
type responsibilityOption func(*ResponsibilityMixin)

// newResponsibilityMixin creates a ResponsibilityMixin for a schema.
func newResponsibilityMixin(schemaType any, opts ...responsibilityOption) ResponsibilityMixin {
	r := ResponsibilityMixin{
		schemaType: schemaType,
		accessCheckAnnotation: func(objectType string) schema.Annotation {
			return accessmap.EdgeViewCheck(objectType)
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	return r
}

func withInternalOwner() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeInternalOwner = true
	}
}

func withBusinessOwner() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeBusinessOwner = true
	}
}

func withTechnicalOwner() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeTechnicalOwner = true
	}
}

func withSecurityOwner() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeSecurityOwner = true
	}
}

func withAssignedTo() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeAssignedTo = true
	}
}

func withReviewedBy() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeReviewedBy = true
	}
}

func withLastReviewedAt() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeLastReviewedAt = true
	}
}

func withDelegate() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.includeDelegate = true
	}
}

func withReviewedByOrderField() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.reviewedByOrderField = true
	}
}

func withAssignedToOrderField() responsibilityOption {
	return func(r *ResponsibilityMixin) {
		r.assignedToOrderField = true
	}
}

func withResponsibilityAccessCheck(accessCheck func(string) schema.Annotation) responsibilityOption {
	return func(r *ResponsibilityMixin) {
		if accessCheck != nil {
			r.accessCheckAnnotation = accessCheck
		}
	}
}

// Fields of the ResponsibilityMixin.
func (r ResponsibilityMixin) Fields() []ent.Field {
	label := schemaLabel(r.schemaType)
	fields := []ent.Field{}

	if r.includeInternalOwner {
		fields = append(fields,
			field.String("internal_owner").
				Comment(fmt.Sprintf("the internal owner for the %s when no user or group is linked", label)).
				Optional().
				Annotations(
					entgql.OrderField("internal_owner"),
					entx.FieldSearchable(),
				),
			field.String("internal_owner_user_id").
				Comment(fmt.Sprintf("the internal owner user id for the %s", label)).
				Optional(),
			field.String("internal_owner_group_id").
				Comment(fmt.Sprintf("the internal owner group id for the %s", label)).
				Optional(),
		)
	}

	if r.includeBusinessOwner {
		fields = append(fields,
			field.String("business_owner").
				Comment(fmt.Sprintf("business owner for the %s when no user or group is linked", label)).
				Optional().
				Annotations(
					entgql.OrderField("business_owner"),
				),
			field.String("business_owner_user_id").
				Comment(fmt.Sprintf("the business owner user id for the %s", label)).
				Optional(),
			field.String("business_owner_group_id").
				Comment(fmt.Sprintf("the business owner group id for the %s", label)).
				Optional(),
		)
	}

	if r.includeTechnicalOwner {
		fields = append(fields,
			field.String("technical_owner").
				Comment(fmt.Sprintf("technical owner for the %s when no user or group is linked", label)).
				Optional().
				Annotations(
					entgql.OrderField("technical_owner"),
				),
			field.String("technical_owner_user_id").
				Comment(fmt.Sprintf("the technical owner user id for the %s", label)).
				Optional(),
			field.String("technical_owner_group_id").
				Comment(fmt.Sprintf("the technical owner group id for the %s", label)).
				Optional(),
		)
	}

	if r.includeSecurityOwner {
		fields = append(fields,
			field.String("security_owner").
				Comment(fmt.Sprintf("security owner for the %s when no user or group is linked", label)).
				Optional().
				Annotations(
					entgql.OrderField("security_owner"),
				),
			field.String("security_owner_user_id").
				Comment(fmt.Sprintf("the security owner user id for the %s", label)).
				Optional(),
			field.String("security_owner_group_id").
				Comment(fmt.Sprintf("the security owner group id for the %s", label)).
				Optional(),
		)
	}

	if r.includeReviewedBy {
		reviewedByField := field.String("reviewed_by").
			Comment(fmt.Sprintf("who reviewed the %s when no user or group is linked", label)).
			Optional()
		if r.reviewedByOrderField {
			reviewedByField = reviewedByField.Annotations(entgql.OrderField("reviewed_by"))
		}

		fields = append(fields,
			reviewedByField,
			field.String("reviewed_by_user_id").
				Comment(fmt.Sprintf("the user id that reviewed the %s", label)).
				Optional(),
			field.String("reviewed_by_group_id").
				Comment(fmt.Sprintf("the group id that reviewed the %s", label)).
				Optional(),
		)
	}

	if r.includeAssignedTo {
		assignedToField := field.String("assigned_to").
			Comment(fmt.Sprintf("who the %s is assigned to when no user or group is linked", label)).
			Optional()
		if r.assignedToOrderField {
			assignedToField = assignedToField.Annotations(entgql.OrderField("assigned_to"))
		}

		fields = append(fields,
			assignedToField,
			field.String("assigned_to_user_id").
				Comment(fmt.Sprintf("the user id assigned to the %s", label)).
				Optional(),
			field.String("assigned_to_group_id").
				Comment(fmt.Sprintf("the group id assigned to the %s", label)).
				Optional(),
		)
	}

	if r.includeLastReviewedAt {
		fields = append(fields,
			field.Time("last_reviewed_at").
				Comment(fmt.Sprintf("when the %s was last reviewed", label)).
				GoType(models.DateTime{}).
				Optional().
				Nillable().
				Annotations(
					entgql.OrderField("last_reviewed_at"),
				),
		)
	}

	if r.includeDelegate {
		fields = append(fields,
			field.String("delegate_id").
				Comment(fmt.Sprintf("the group id delegated for the %s", label)).
				Optional().
				Unique(),
		)
	}

	return fields
}

// Edges of the ResponsibilityMixin.
func (r ResponsibilityMixin) Edges() []ent.Edge {
	check := r.accessCheckAnnotation
	if check == nil {
		check = func(objectType string) schema.Annotation {
			return accessmap.EdgeViewCheck(objectType)
		}
	}

	edges := []ent.Edge{}

	if r.includeInternalOwner {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "internal_owner_user",
				t:          User.Type,
				field:      "internal_owner_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "internal_owner_group",
				t:          Group.Type,
				field:      "internal_owner_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeBusinessOwner {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "business_owner_user",
				t:          User.Type,
				field:      "business_owner_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "business_owner_group",
				t:          Group.Type,
				field:      "business_owner_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeTechnicalOwner {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "technical_owner_user",
				t:          User.Type,
				field:      "technical_owner_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "technical_owner_group",
				t:          Group.Type,
				field:      "technical_owner_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeSecurityOwner {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "security_owner_user",
				t:          User.Type,
				field:      "security_owner_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "security_owner_group",
				t:          Group.Type,
				field:      "security_owner_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeReviewedBy {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "reviewed_by_user",
				t:          User.Type,
				field:      "reviewed_by_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "reviewed_by_group",
				t:          Group.Type,
				field:      "reviewed_by_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeAssignedTo {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "assigned_to_user",
				t:          User.Type,
				field:      "assigned_to_user_id",
				annotations: []schema.Annotation{
					check(User{}.Name()),
				},
			}),
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "assigned_to_group",
				t:          Group.Type,
				field:      "assigned_to_group_id",
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	if r.includeDelegate {
		edges = append(edges,
			uniqueEdgeTo(&edgeDefinition{
				fromSchema: r.schemaType,
				name:       "delegate",
				t:          Group.Type,
				field:      "delegate_id",
				comment:    fmt.Sprintf("temporary delegate for the %s, used for temporary ownership", schemaLabel(r.schemaType)),
				annotations: []schema.Annotation{
					check(Group{}.Name()),
				},
			}),
		)
	}

	return edges
}

func schemaLabel(schemaType any) string {
	schemaName := toSchemaFuncs(schemaType).Name()

	return strings.ReplaceAll(schemaName, "_", " ")
}
