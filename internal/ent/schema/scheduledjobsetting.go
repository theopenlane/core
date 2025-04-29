package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx/history"
)

// ScheduledJobSetting holds the schema definition for the ScheduledJobSetting entity
type ScheduledJobSetting struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobSettingis the name of the schema in snake case
const SchemaScheduledJobSetting = "scheduled_job_setting"

func (ScheduledJobSetting) Name() string {
	return SchemaScheduledJobSetting
}

func (ScheduledJobSetting) GetType() any {
	return ScheduledJobSetting.Type
}

func (ScheduledJobSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaScheduledJobSetting)
}

// Fields of the ScheduledJobSetting
func (ScheduledJobSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("scheduled_job_id").
			Optional(),

		field.JSON("configuration", models.JobConfiguration{}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			).
			Comment("the configuration to run this job"),

		field.JSON("cadence", models.JobCadence{}).
			Comment("the schedule to run this job").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			).
			Optional(),

		field.String("cron").
			Comment("cron syntax").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput |
					entgql.SkipOrderField,
				),
			).
			Optional().
			Nillable(),
	}
}

// Mixin of the ScheduledJobSetting
func (s ScheduledJobSetting) Mixin() []ent.Mixin {
	return getDefaultMixins()
}

// Edges of the ScheduledJobSetting
func (s ScheduledJobSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ScheduledJob{},
			ref:        "scheduled_job_setting",
			field:      "scheduled_job_id",
		}),
	}
}

// Indexes of the ScheduledJobSetting
func (ScheduledJobSetting) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJobSetting
func (ScheduledJobSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the ScheduledJobSetting
func (ScheduledJobSetting) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the ScheduledJobSetting
func (ScheduledJobSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ScheduledJobSetting
func (ScheduledJobSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			// entfga.CheckEditAccess[*generated.ScheduledJobSettingMutation](),
		),
	)
}
