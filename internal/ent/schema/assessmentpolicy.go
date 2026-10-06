package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
)

// AssessmentPolicy defines the mapping between assessments and the internal policies they attest to
type AssessmentPolicy struct {
	SchemaFuncs

	ent.Schema
}

// SchemaAssessmentPolicy is the name of the assessment policy schema
const SchemaAssessmentPolicy = "assessment_policy"

// Name returns the name of the assessment policy schema
func (AssessmentPolicy) Name() string {
	return SchemaAssessmentPolicy
}

// GetType returns the type of the assessment policy schema
func (AssessmentPolicy) GetType() any {
	return AssessmentPolicy.Type
}

// PluralName returns the plural name of the assessment policy schema
func (AssessmentPolicy) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAssessmentPolicy)
}

// Fields returns assessment policy fields
func (AssessmentPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("assessment_id").
			Immutable().
			Comment("the id of the assessment attesting to the policy"),
		field.String("internal_policy_id").
			Immutable().
			Comment("the id of the internal policy being attested to"),
		field.String("policy_revision").
			Optional().
			Immutable().
			Comment("the revision of the internal policy when it was added to the assessment"),
	}
}

// Edges of the AssessmentPolicy
func (ap AssessmentPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: ap,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Assessment{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: ap,
			edgeSchema: InternalPolicy{},
			field:      "internal_policy_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(InternalPolicy{}.Name()),
			},
		}),
	}
}

// Indexes of the AssessmentPolicy
func (AssessmentPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("assessment_id", "internal_policy_id").
			Unique(),
	}
}

// Mixin of the AssessmentPolicy
func (ap AssessmentPolicy) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeSoftDelete: true, // we cannot soft delete on a through table
		excludeTags:       true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(ap),
		},
	}.getMixins(ap)
}

// Modules of the AssessmentPolicy
func (AssessmentPolicy) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Hooks of the AssessmentPolicy
func (AssessmentPolicy) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookAssessmentPolicyRevision(),
	}
}

// Policy of the AssessmentPolicy
func (AssessmentPolicy) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{Assessment{}.Name(), InternalPolicy{}.Name()}),
			policy.CanEditObjectUnderParents([]string{Assessment{}.Name()}, assessmentPolicyParentID),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// assessmentPolicyParentID returns the value of a parent id field on the assessment_policy being mutated
func assessmentPolicyParentID(ctx context.Context, m generated.Mutation, field string) (string, error) {
	apm, ok := m.(*generated.AssessmentPolicyMutation)
	if !ok {
		return "", nil
	}

	id, ok := apm.ID()
	if !ok {
		return "", nil
	}

	ap, err := apm.Client().AssessmentPolicy.Get(ctx, id)
	if err != nil {
		return "", err
	}

	if field == "assessment_id" {
		return ap.AssessmentID, nil
	}

	return "", nil
}
