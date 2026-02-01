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

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/validator"
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
			UpdateDefault(time.Now).
			Optional().
			Nillable(),
		field.Bool("dedicated_db").
			Comment("Whether the organization has a dedicated database").
			Default(false). // default to shared db
			// TODO: https://github.com/theopenlane/core/issues/734
			// update this once feature functionality is enabled
			// Annotations(
			// 	entgql.Skip(),
			// ),
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationUpdateInput, entgql.SkipOrderField),
			),
		field.String("stripe_customer_id").
			Comment("the stripe customer ID this organization is associated to").
			Optional().
			Nillable().
			Unique().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput |
					entgql.SkipWhereInput | entgql.SkipOrderField),
			),
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
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         APIToken{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         EmailBranding{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         NotificationPreference{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         NotificationTemplate{},
			cascadeDeleteOwner: true,
		}),
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
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Hush{},
			cascadeDeleteOwner: true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			name:       "avatar_file",
			t:          File.Type,
			field:      "avatar_local_file_id",
		}),

		// Organization owns the following entities
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Group{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Template{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Integration{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DocumentData{},
			cascadeDeleteOwner: true,
		}),
		edge.To(OrgSubscription{}.PluralName(), OrgSubscription.Type).
			Annotations(
				entx.CascadeAnnotationField("Owner"),
				accessmap.EdgeNoAuthCheck(),
			),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         OrgProduct{},
			cascadeDeleteOwner: true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         OrgPrice{},
			cascadeDeleteOwner: true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         OrgModule{},
			cascadeDeleteOwner: true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Invite{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Subscriber{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Entity{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Platform{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         IdentityHolder{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Campaign{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         CampaignTarget{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         EntityType{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Contact{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Note{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Task{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Program{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Procedure{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         InternalPolicy{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Risk{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         ControlObjective{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Narrative{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Control{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Subcontrol{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         ControlImplementation{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         MappedControl{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Evidence{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Standard{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         ActionPlan{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         CustomDomain{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         JobRunner{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         JobRunnerToken{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         JobRunnerRegistrationToken{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DNSVerification{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         JobTemplate{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         ScheduledJob{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         JobResult{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         ScheduledJobRun{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         TrustCenter{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Asset{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Scan{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Subprocessor{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Export{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         TrustCenterWatermarkConfig{},
			cascadeDeleteOwner: true,
		}),
		defaultEdgeToWithPagination(o, ImpersonationEvent{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Assessment{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         AssessmentResponse{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         CustomTypeEnum{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         TagDefinition{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Remediation{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Finding{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Review{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Vulnerability{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Notification{},
			cascadeDeleteOwner: true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowDefinition{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowInstance{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowEvent{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowAssignment{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowAssignmentTarget{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowObjectRef{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         WorkflowProposal{},
			cascadeDeleteOwner: true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DirectoryAccount{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DirectoryGroup{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DirectoryMembership{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         DirectorySyncRun{},
			cascadeDeleteOwner: true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:         o,
			edgeSchema:         Discussion{},
			cascadeDeleteOwner: true,
		}),
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
		entfga.SelfAccessChecks(),
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
		),
		policy.WithMutationRules(
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
