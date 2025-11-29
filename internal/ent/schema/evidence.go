package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// Evidence holds the schema definition for the Evidence entity
type Evidence struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEvidence is the name of the Evidence schema.
const SchemaEvidence = "evidence"

// Name returns the name of the Evidence schema.
func (Evidence) Name() string {
	return SchemaEvidence
}

// GetType returns the type of the Evidence schema.
func (Evidence) GetType() any {
	return Evidence.Type
}

// PluralName returns the plural name of the Evidence schema.
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
			Default(enums.EvidenceStatusSubmitted.String()).
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
			newObjectOwnedMixin[generated.Evidence](e,
				withParents(
					Control{}, Subcontrol{}, ControlObjective{}, Program{},
					Task{}, Procedure{}, InternalPolicy{}), // used to create parent tuples for the evidence
				withOrganizationOwner(true),
			),
		},
	}.getMixins(e)
}

// Edges of the Evidence
func (e Evidence) Edges() []ent.Edge {
	return []ent.Edge{
		// users with only view access should be able to link
		// controls to the evidence
		edgeToWithPagination(&edgeDefinition{
			fromSchema: e,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Control{}.Name()),
			},
		}),
		// users with only view access should be able to link
		// subcontrols to the evidence
		edgeToWithPagination(&edgeDefinition{
			fromSchema: e,
			edgeSchema: Subcontrol{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Subcontrol{}.Name()),
			},
		}),
		// all other edges require edit access to make the association
		defaultEdgeToWithPagination(e, ControlObjective{}),
		defaultEdgeToWithPagination(e, ControlImplementation{}),
		defaultEdgeToWithPagination(e, File{}),
		defaultEdgeFromWithPagination(e, Program{}),
		defaultEdgeFromWithPagination(e, Task{}),
	}
}

func (Evidence) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
	}
}

// Annotations of the Evidence
func (e Evidence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Hooks of the Evidence
func (Evidence) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEvidenceFiles(),
		hooks.HookSystemOwnedControls(),
	}
}

// Policy of the Evidence
func (Evidence) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			// to create evidence under specific objects, this needs to run
			// before the generic CheckCreateAccess because that will
			// return a privacy.Deny, this will only return a privacy.Skip
			// if there are no parents
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
				Control{}.PluralName(),
				Subcontrol{}.PluralName(),
				Task{}.PluralName(),
			}),
			policy.CheckCreateAccess(), // generic create access on the organization level
			// users without org level can_create_evidence should be able
			entfga.CheckEditAccess[*generated.EvidenceMutation](),
		),
	)
}
