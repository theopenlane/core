package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobis the name of the schema in snake case
const SchemaScheduledJob = "job"

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
			Comment("the title of the job").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the job").
			Optional(),

		field.Enum("platform").
			GoType(enums.JobPlatformType("")).
			Immutable().
			Annotations(
				entgql.OrderField("platform"),
			).
			Comment("the platform to use to execute this job"),

		field.String("windmill_path").
			Annotations(
				entgql.Skip(
					entgql.SkipOrderField |
						entgql.SkipWhereInput |
						entgql.SkipMutationCreateInput |
						entgql.SkipMutationUpdateInput,
				),
			).
			Comment("Windmill path"),

		field.String("download_url").
			Annotations(
				entgql.Skip(
					entgql.SkipOrderField |
						entgql.SkipWhereInput,
				),
			).
			Comment("the url from where to download the script from"),

		field.JSON("configuration", models.JobConfiguration{}).
			Optional().
			Comment("the configuration to run this job"),

		field.String("cron").
			GoType(models.Cron("")).
			Comment("cron syntax").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Optional().
			Nillable(),
	}
}

// Mixin of the ScheduledJob
func (s ScheduledJob) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "JOB",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s,
				withSkipForSystemAdmin(true),
			),
			mixin.SystemOwnedMixin{},
		},
	}.getMixins()
}

func (ScheduledJob) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the ScheduledJob
func (s ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// entfga.SelfAccessChecks(),
	}
}

// Interceptors of the ScheduledJob
func (s ScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorRequireAnyFeature("scheduledjob", s.Features()...),
	}
}

// Hooks of the ScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookScheduledJobCreate(),
	}
}

// Policy of the ScheduledJob
func (s ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.DenyIfMissingAllFeatures(s.Features()...),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
