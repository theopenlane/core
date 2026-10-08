package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/token"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
	"github.com/theopenlane/core/v2/internal/ent/validator"
)

const (
	orgNameMaxLen = 160
)

// Organization holds the schema definition for the Organization entity - organizations are the top level tenancy construct in the system
type Organization struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrganization = "organization"

func (Organization) Name() string {
	return SchemaOrganization
}

func (Organization) GetType() any {
	return Organization.Type
}

func (Organization) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrganization)
}

// Fields of the Organization
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the organization").
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			Immutable().
			MaxLen(orgNameMaxLen).
			MinLen(minNameLength).
			Validate(validator.SpecialCharValidator).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("display_name").
			Comment("The organization's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			Default("").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("display_name"),
			),
		field.String("description").
			Comment("An optional description of the organization").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("parent_organization_id").Optional().Immutable().
			Comment("The ID of the parent organization for the organization.").
			Annotations(
				entgql.Type("ID"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipType),
			),
		field.Bool("personal_org").
			Comment("orgs directly associated with a user").
			Optional().
			Default(false).
			Immutable(),
		field.String("avatar_remote_url").
			Comment("URL of the user's remote avatar").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("avatar_local_file_id").
			Comment("The organizations's local avatar file id, takes precedence over the avatar remote URL").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
		field.Time("avatar_updated_at").
			Comment("The time the user's (local) avatar was last updated").
			Default(time.Now).
			Optional().
			Nillable(),
		field.String("stripe_customer_id").
			Comment("the stripe customer ID this organization is associated to").
			Optional().
			Nillable().
			Unique().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput |
					entgql.SkipWhereInput | entgql.SkipOrderField),
			),
		field.String("slug_name").
			Comment("a stable slug identifying the organization in its public SSO initiation URL, e.g. /orgs/<sso_slug>/sso").
			MaxLen(orgNameMaxLen).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional(),
	}
}

// Edges of the Organization
func (o Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Organization.Type).
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				accessmap.EdgeNoAuthCheck(),
			).
			From("parent").
			Field("parent_organization_id").
			Immutable().
			Unique().
			Annotations(
				entx.CascadeAnnotationField("Child"),
				accessmap.EdgeNoAuthCheck(),
			),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema:    o,
			name:          "setting",
			t:             OrganizationSetting.Type,
			cascadeDelete: "Organization",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		defaultEdgeToWithPagination(o, PersonalAccessToken{}),
		hiddenOwnerEdge(o, APIToken{}),
		hiddenOwnerEdge(o, EmailTemplate{}),
		hiddenOwnerEdge(o, IntegrationWebhook{}),
		hiddenOwnerEdge(o, IntegrationRun{}),
		hiddenOwnerEdge(o, NotificationPreference{}),
		hiddenOwnerEdge(o, NotificationTemplate{}),
		edge.From("users", User.Type).
			Ref("organizations").
			// Skip the mutation input for the users edge
			// this should be done via the members edge
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entgql.RelayConnection(),
				accessmap.EdgeNoAuthCheck(),
			).
			Through("members", OrgMembership.Type),

		// files can be owned by an organization, but don't have to be
		// only those with the organization id set should be cascade deleted
		edgeToWithPagination(&edgeDefinition{
			fromSchema: o,
			edgeSchema: File{},
			annotations: []schema.Annotation{
				entx.CascadeAnnotationField("Organization"), // 1:m so we override the default
			},
			cascadeDeleteOwner: true,
		}),
		defaultEdgeToWithPagination(o, Event{}),
		hiddenOwnerEdge(o, Hush{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			name:       "avatar_file",
			t:          File.Type,
			field:      "avatar_local_file_id",
		}),

		// Organization owns the following entities
		hiddenOwnerEdge(o, Group{}),
		hiddenOwnerEdge(o, Template{}),
		hiddenOwnerEdge(o, Integration{}),
		hiddenOwnerEdge(o, DocumentData{}),
		edge.To(OrgSubscription{}.PluralName(), OrgSubscription.Type).
			Annotations(
				entx.CascadeAnnotationField("Owner"),
				accessmap.EdgeNoAuthCheck(),
			),
		hiddenOwnerEdge(o, OrgProduct{}),
		hiddenOwnerEdge(o, OrgPrice{}),
		hiddenOwnerEdge(o, OrgModule{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Invite{},
			cascadeDeleteOwner: true,
		}),
		hiddenOwnerEdge(o, Subscriber{}),
		hiddenOwnerEdge(o, Entity{}),
		hiddenOwnerEdge(o, Platform{}),
		hiddenOwnerEdge(o, IdentityHolder{}),
		hiddenOwnerEdge(o, Campaign{}),
		hiddenOwnerEdge(o, CampaignTarget{}),
		hiddenOwnerEdge(o, EntityType{}),
		hiddenOwnerEdge(o, Contact{}),
		hiddenOwnerEdge(o, Note{}),
		hiddenOwnerEdge(o, Task{}),
		hiddenOwnerEdge(o, Program{}),
		hiddenOwnerEdge(o, SystemDetail{}),
		hiddenOwnerEdge(o, Procedure{}),
		hiddenOwnerEdge(o, InternalPolicy{}),
		hiddenOwnerEdge(o, Risk{}),
		hiddenOwnerEdge(o, ControlObjective{}),
		hiddenOwnerEdge(o, Narrative{}),
		hiddenOwnerEdge(o, Control{}),
		hiddenOwnerEdge(o, Subcontrol{}),
		hiddenOwnerEdge(o, ControlImplementation{}),
		hiddenOwnerEdge(o, MappedControl{}),
		hiddenOwnerEdge(o, Evidence{}),
		hiddenOwnerEdge(o, Standard{}),
		hiddenOwnerEdge(o, ActionPlan{}),
		hiddenOwnerEdge(o, CustomDomain{}),
		hiddenOwnerEdge(o, DNSVerification{}),
		hiddenOwnerEdge(o, TrustCenter{}),
		hiddenOwnerEdge(o, Asset{}),
		hiddenOwnerEdge(o, Scan{}),
		hiddenOwnerEdge(o, SLADefinition{}),
		hiddenOwnerEdge(o, Subprocessor{}),
		hiddenOwnerEdge(o, Export{}),
		hiddenOwnerEdge(o, TrustCenterWatermarkConfig{}),
		defaultEdgeToWithPagination(o, ImpersonationEvent{}),
		hiddenOwnerEdge(o, Assessment{}),
		hiddenOwnerEdge(o, AssessmentResponse{}),
		hiddenOwnerEdge(o, AssessmentPolicy{}),
		hiddenOwnerEdge(o, CustomTypeEnum{}),
		hiddenOwnerEdge(o, TagDefinition{}),
		hiddenOwnerEdge(o, Remediation{}),
		hiddenOwnerEdge(o, Finding{}),
		hiddenOwnerEdge(o, FindingControl{}),
		hiddenOwnerEdge(o, Review{}),
		hiddenOwnerEdge(o, Vulnerability{}),
		hiddenOwnerEdge(o, Notification{}),
		hiddenOwnerEdge(o, WorkflowDefinition{}),
		hiddenOwnerEdge(o, WorkflowInstance{}),
		hiddenOwnerEdge(o, WorkflowEvent{}),
		hiddenOwnerEdge(o, WorkflowAssignment{}),
		hiddenOwnerEdge(o, WorkflowAssignmentTarget{}),
		hiddenOwnerEdge(o, WorkflowObjectRef{}),
		hiddenOwnerEdge(o, WorkflowProposal{}),
		hiddenOwnerEdge(o, DirectoryAccount{}),
		hiddenOwnerEdge(o, DirectoryGroup{}),
		hiddenOwnerEdge(o, DirectoryMembership{}),
		hiddenOwnerEdge(o, Discussion{}),
		hiddenOwnerEdge(o, VendorScoringConfig{}),
		hiddenOwnerEdge(o, VendorRiskScore{}),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Annotations of the Organization
func (o Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Delete org members when orgs are deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "Organization",
					Through: "OrgMembership",
				},
			},
		),
		entx.FileCategory(SchemaOrganization),
		entfga.SelfAccessChecks(),
		entx.FGACrudSkip(entx.SkipDelete | entx.SkipCreate),
		entx.SchemaTaskRule(taskrules.OrganizationSuggestedRules...),
	}
}

// Mixin of the Organization
func (Organization) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			// add group based create permissions
			NewGroupBasedCreateAccessMixin(),
		},
	}.getMixins(Organization{})
}

// Policy defines the privacy policy of the Organization.
func (Organization) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](), // Allow invite tokens to query the org ID they are invited to
			rule.AllowIfContextHasPrivacyTokenOfType[*token.SignUpToken](),    // Allow sign-up tokens to query the org ID they are subscribing to
			policy.CheckOrgReadAccess(),                                       // access based on query and auth context
			policy.CheckOrgAuditorAccess(),
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			rule.HasOrgMutationAccess(), // Requires edit for Update, and delete for Delete mutations
			policy.AllowCreate(),        // Allow all other users (e.g. a user with a JWT should be able to create a new org)
		),
	)
}

// Interceptors of the Organization
func (o Organization) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorOrganization(),
	}
}

// Hooks of the Organization
func (Organization) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOrganization(),
		hooks.HookOrganizationDelete(),
		hooks.HookOrganizationCreatePolicy(),
	}
}
