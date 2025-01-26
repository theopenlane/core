package schema

import (
	"net/url"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

const (
	orgNameMaxLen = 160
)

// Organization holds the schema definition for the Organization entity - organizations are the top level tenancy construct in the system
type Organization struct {
	ent.Schema
}

// Fields of the Organization
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the organization").
			MaxLen(orgNameMaxLen).
			NotEmpty().
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
			Validate(func(s string) error {
				_, err := url.Parse(s)
				return err
			}).
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
	}
}

// Edges of the Organization
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Organization.Type).
			Annotations(
				entgql.RelayConnection(),
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			From("parent").
			Field("parent_organization_id").
			Immutable().
			Unique().
			Annotations(
				entx.CascadeAnnotationField("Child"),
			),

		edge.To("setting", OrganizationSetting.Type).
			Unique().
			Annotations(
				entx.CascadeAnnotationField("Organization"),
			),

		edge.To("personal_access_tokens", PersonalAccessToken.Type),
		edge.To("api_tokens", APIToken.Type),
		edge.From("users", User.Type).
			Ref("organizations").
			Through("members", OrgMembership.Type),

		edge.To("events", Event.Type),
		edge.To("secrets", Hush.Type),
		edge.To("files", File.Type),
		edge.To("avatar_file", File.Type).
			Field("avatar_local_file_id").Unique(),

		// Organization owns the following entities
		edge.To("groups", Group.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("templates", Template.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("integrations", Integration.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("document_data", DocumentData.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("org_subscriptions", OrgSubscription.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("invites", Invite.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("subscribers", Subscriber.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("entities", Entity.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("entity_types", EntityType.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("contacts", Contact.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("notes", Note.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("tasks", Task.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("programs", Program.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("procedures", Procedure.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("internal_policies", InternalPolicy.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("risks", Risk.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("control_objectives", ControlObjective.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("narratives", Narrative.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("controls", Control.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("subcontrols", Subcontrol.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
		edge.To("evidence", Evidence.Type).
			Annotations(entx.CascadeAnnotationField("Owner")),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the Organization
func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
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
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.SoftDeleteMixin{},
		// add group based create permissions
		NewGroupBasedCreateAccessMixin(true),
	}
}

// Policy defines the privacy policy of the Organization.
func (Organization) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfContextHasPrivacyTokenOfType(&token.OrgInviteToken{}), // Allow invite tokens to query the org ID they are invited to
			rule.AllowIfContextHasPrivacyTokenOfType(&token.SignUpToken{}),    // Allow sign-up tokens to query the org ID they are subscribing to
			entfga.CheckReadAccess[*generated.OrganizationQuery](),            // access based on query context
			policy.CheckOrgReadAccess(),                                       // access based on auth context
		),
		policy.WithMutationRules(
			rule.HasOrgMutationAccess(), // Requires edit for Update, and delete for Delete mutations
			privacy.AlwaysAllowRule(),   // Allow all other users (e.g. a user with a JWT should be able to create a new org)
		),
	)
}

// Interceptors of the Organization
func (Organization) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorOrganization(),
	}
}

// Hooks of the Organization
func (Organization) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOrganization(),
		hooks.HookOrganizationDelete(),
	}
}
