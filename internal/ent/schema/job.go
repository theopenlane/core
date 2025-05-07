package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	emixin "github.com/theopenlane/entx/mixin"
)

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobis the name of the schema in snake case
const SchemaScheduledJob = "scheduled_job"

func (ScheduledJob) Name() string {
	return SchemaScheduledJob
}

func (ScheduledJob) GetType() any {
	return ScheduledJob.Type
}

func (ScheduledJob) PluralName() string {
	return pluralize.NewClient().Plural(SchemaScheduledJob)
}

// Fields of the ScheduledJob
func (ScheduledJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Comment("the title of the task").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the task").
			Optional(),

		field.Enum("environment").
			GoType(enums.JobEnvironment("")).
			Default(enums.JobEnvironmentOpenlane.String()).
			Annotations(
				entgql.OrderField("environment"),
				entgql.Skip(entgql.SkipWhereInput|entgql.SkipOrderField),
			).
			Comment("where this job is going to run?"),

		field.String("script").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Comment("the script to run").
			Optional(),

		// schema can be org owned but that would be
		// for the future when orgs can add their own
		// jobs on their infra
		//
		// all marketplaces and native jobs would be "system owned"
		// but `is_marketplace` will be a marker for which
		// one is provided by Openlane and which ones were
		// submitted by the community
		field.Bool("is_marketplace").
			Default(false).
			Annotations(
				entgql.OrderField("is_marketplace"),
				entgql.Skip(^entgql.SkipType),
			).
			Comment("Is this provided by Openlane?"),
	}
}

// Mixin of the ScheduledJob
func (s ScheduledJob) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.IDMixin{
			HumanIdentifierPrefix: "JOB",
			SingleFieldIndex:      true,
		},
		mixin.SoftDeleteMixin{},
		emixin.AuditMixin{},
		newOrgOwnedMixin(s,
			withSkipForSystemAdmin(true),
		),
		mixin.SystemOwnedMixin{},
	}
}

// Edges of the ScheduledJob
func (s ScheduledJob) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the ScheduledJob
func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJob
func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the ScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the ScheduledJob
func (ScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ScheduledJob
func (ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
		// entfga.CheckEditAndDeleteAccess[*generated.ScheduledJobMutation](),
		),
	)
}
