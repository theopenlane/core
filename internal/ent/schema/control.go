package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/oscalgen"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/directives"
)

// Control defines the control schema.
type Control struct {
	SchemaFuncs

	ent.Schema
}

// SchemaControl is the name of the control schema.
const SchemaControl = "control"

// Name returns the name of the control schema.
func (Control) Name() string {
	return SchemaControl
}

// GetType returns the type of the control schema.
func (Control) GetType() any {
	return Control.Type
}

// PluralName returns the plural name of the control schema.
func (Control) PluralName() string {
	return pluralize.NewClient().Plural(SchemaControl)
}

// Fields returns control fields.
func (Control) Fields() []ent.Field {
	// add any fields that are specific to the parent control here
	additionalFields := []ent.Field{
		field.String("ref_code").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entx.FieldWebhookPayloadField(),
				entgql.OrderField("ref_code"),
				directives.ExternalSourceDirectiveAnnotation,
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleControlID,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
					oscalgen.WithOSCALIdentityAnchor(),
				),
			).
			Comment("the unique reference code for the control"),
		field.String("standard_id").
			Comment("the id of the standard that the control belongs to, if applicable").
			Optional(),
		field.Enum("trust_center_visibility").
			GoType(enums.TrustCenterControlVisibility("")).
			Default(enums.TrustCenterControlVisibilityNotVisible.String()).
			Optional().
			Comment("visibility of the control on the trust center, controls the publishing state for trust center display"),
		field.Bool("is_trust_center_control").
			Default(false).
			Optional().
			Immutable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Comment("indicates the control is derived from the trust center standard, set by the system during control clone"),
	}

	return additionalFields
}

// Edges of the Control
func (c Control) Edges() []ent.Edge {
	return []ent.Edge{
		// parents of the control (standard, program)
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Standard{},
			field:      "standard_id",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
		defaultEdgeFromWithPagination(c, Program{}),
		defaultEdgeFromWithPagination(c, Platform{}),
		defaultEdgeToWithPagination(c, Asset{}),
		defaultEdgeToWithPagination(c, Scan{}),
		edge.From("findings", Finding.Type).
			Ref("controls").
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
			).
			Through("control_mappings", FindingControl.Type),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ControlImplementation{},
			comment:    "the implementation(s) of the control",
			annotations: []schema.Annotation{
				oscalgen.NewOSCALRelationship(
					oscalgen.OSCALRelationshipRoleImplementedByComponent,
					oscalgen.WithOSCALRelationshipModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
				),
			},
		}),
		// controls have control objectives and subcontrols
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    c,
			edgeSchema:    Subcontrol{},
			cascadeDelete: "Control",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ScheduledJob{},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			ref:        "to_controls",
			name:       "mapped_to_controls",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			ref:        "from_controls",
			name:       "mapped_from_controls",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: WorkflowObjectRef{},
			ref:        "control",
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
	}
}

func (Control) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the standard
		index.Fields("standard_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL AND owner_id is NULL"),
		),
		// ref_code and standard id should be unique within the organization
		index.Fields("standard_id", "ref_code", "owner_id").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL AND owner_id is not NULL and standard_id is not NULL"),
		),
		// ref_code should be unique for controls inside an organization that are not associated with a standard
		index.Fields("ref_code", "owner_id").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL AND owner_id is not NULL and standard_id is NULL"),
		),
		index.Fields("standard_id", "deleted_at", "owner_id"),
		index.Fields("reference_id", "deleted_at", "owner_id"),
		index.Fields("auditor_reference_id", "deleted_at", "owner_id"),
	}
}

// Mixin of the Control
func (c Control) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "CTL",
		additionalMixins: []ent.Mixin{
			// add the common overlap between control and subcontrol
			ControlMixin{
				SchemaType: c,
			},
			// controls must be associated with an organization but do not inherit permissions from the organization
			// controls can inherit permissions from the associated programs
			newObjectOwnedMixin[generated.Control](c,
				withParents(Organization{}, Program{}, Standard{}),
				withOrganizationOwner(true),
				// controls are generally viewable by all users in the organization
				// exceptions are based on group based access so we can safely
				// skip the interceptor
				withSkipFilterInterceptor(interceptors.SkipAllQuery|interceptors.SkipIDsQuery),
				withWorkflowOwnedEdges(),
				withAllowAnonymousTrustCenterAccess(true),
			),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
			// add groups permissions with editor, and blocked groups
			// skip view because controls are automatically viewable by all users in the organization
			newGroupPermissionsMixin(withSkipViewPermissions(), withGroupPermissionsInterceptor(), withWorkflowGroupEdges()),
			newCustomEnumMixin(c, withWorkflowEnumEdges()),
			newCustomEnumMixin(c, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(c, withEnumFieldName("scope"), withGlobalEnum()),
			WorkflowApprovalMixin{},
		},
	}.getMixins(c)
}

func (Control) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookControlReferenceFramework(),
		hooks.HookControlTrustCenterVisibility(),
	}
}

// Interceptors of the Control
func (Control) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterControl(),
	}
}

// Policy of the Control
func (c Control) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			// when an admin deletes a standard, we updated
			// controls to unlink the standard that might belong to an organization
			rule.AllowMutationIfSystemAdmin(),
			rule.AllowIfContextAllowRule(),
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlMutation](),
		),
	)
}

func (Control) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the Control
func (c Control) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
			entx.WithSystemOwned(),
		),
		oscalgen.NewOSCALModel(
			oscalgen.WithOSCALModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
			oscalgen.WithOSCALAssembly("implemented-requirement"),
		),
	}
}
