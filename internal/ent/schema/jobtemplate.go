package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// JobTemplate holds the schema definition for the JobTemplate entity
type JobTemplate struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobTemplate is the name of the schema in snake case
const SchemaJobTemplate = "job_template"

func (JobTemplate) Name() string {
	return SchemaJobTemplate
}

func (JobTemplate) GetType() any {
	return JobTemplate.Type
}

func (JobTemplate) PluralName() string {
	return pluralize.NewClient().Plural(SchemaJobTemplate)
}

// Fields of the JobTemplate
func (JobTemplate) Fields() []ent.Field {
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

// Mixin of the JobTemplate
func (s JobTemplate) Mixin() []ent.Mixin {
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

// Annotations of the JobTemplate
func (JobTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Hooks of the JobTemplate
func (JobTemplate) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookJobTemplateCreate(),
	}
}

// Policy of the JobTemplate
func (JobTemplate) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
