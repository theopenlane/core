package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// Evidence holds the schema definition for the Evidence entity
type Evidence struct {
	SchemaFuncs

	ent.Schema
}

const SchemaEvidence = "evidence"

func (Evidence) Name() string {
	return SchemaEvidence
}

func (Evidence) GetType() any {
	return Evidence.Type
}

func (Evidence) PluralName() string {
	return SchemaEvidence // special case because evidences is a weird plural
}

// Fields of the Evidence
func (Evidence) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the evidence").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			NotEmpty(),
		field.String("description").
			Comment("the description of the evidence, what is contained in the uploaded file(s) or url(s)").
			Optional(),
		field.Text("collection_procedure").
			Comment("description of how the evidence was collected").
			Optional(),
		field.Time("creation_date").
			Comment("the date the evidence was retrieved").
			Annotations(
				entgql.OrderField("creation_date"),
			).
			Default(time.Now),
		field.Time("renewal_date").
			Comment("the date the evidence should be renewed, defaults to a year from entry date").
			Default(time.Now().AddDate(1, 0, 0)).
			Annotations(
				entgql.OrderField("renewal_date"),
			).
			Optional(),
		field.String("source").
			Comment("the source of the evidence, e.g. system the evidence was retrieved from (splunk, github, etc)").
			Optional(),
		field.Bool("is_automated").
			Comment("whether the evidence was automatically generated").
			Optional().
			Default(false),
		field.String("url").
			Optional().
			Validate(validator.ValidateURL()).
			Comment("the url of the evidence if not uploaded directly to the system"),
		field.Enum("status").
			GoType(enums.EvidenceStatus("")).
			Default(enums.EvidenceReady.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Comment("the status of the evidence, ready, approved, needs renewal, missing artifact, rejected").
			Optional(),
	}
}

// Mixin of the Evidence
func (e Evidence) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "EVD",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin(e,
				withParents(
					Control{}, Subcontrol{}, ControlObjective{}, Program{},
					Task{}, Procedure{}, InternalPolicy{}), // used to create parent tuples for the evidence
				withOrganizationOwner(false),
			),
		},
	}.getMixins()
}

// Edges of the Evidence
func (e Evidence) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, ControlObjective{}),
		defaultEdgeToWithPagination(e, Control{}),
		defaultEdgeToWithPagination(e, Subcontrol{}),
		defaultEdgeToWithPagination(e, File{}),
		defaultEdgeFromWithPagination(e, Program{}),
		defaultEdgeFromWithPagination(e, Task{}),
	}
}

// Annotations of the Evidence
func (Evidence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Evidence
func (Evidence) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEvidenceFiles(),
	}
}

// Policy of the Evidence
func (Evidence) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.EvidenceMutation](),
		),
	)
}
