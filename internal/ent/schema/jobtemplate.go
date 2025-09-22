package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
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
			Comment("the short description of the job and what it does").
			Optional(),
		field.Enum("platform").
			GoType(enums.JobPlatformType("")).
			Immutable().
			Annotations(
				entgql.OrderField("PLATFORM"),
			).
			Comment("the platform to use to execute this job, e.g. golang, typescript, python, etc."),
		field.String("windmill_path").
			Annotations(
				entgql.Skip(
					entgql.SkipAll, // hidden from the graphql api, this is an internal field used to track the windmill path
				),
			).
			Optional().
			Comment("windmill path used to execute the job"),
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
			Annotations(
				entgql.Skip(
					entgql.SkipWhereInput | entgql.SkipOrderField,
				),
			).
			Comment("the json configuration to run this job, which could be used to template a job, e.g. { \"account_name\": \"my-account\" }"),
		field.String("cron").
			GoType(models.Cron("")).
			Comment("cron schedule to run the job in cron 6-field syntax, e.g. 0 0 0 * * *").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				c := models.Cron(s)

				return c.Validate()
			}).
			Optional().
			Nillable(),
	}
}

// Mixin of the JobTemplate
func (j JobTemplate) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "JBT",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j),
			// TODO: this was added but public access tuples are not
			// yet implemented; so users cannot access job templates
			// created by system admins
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(j)
}

// Annotations of the JobTemplate
func (JobTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the JobTemplate
func (JobTemplate) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookJobTemplate(),
		hook.On(
			hooks.OrgOwnedTuplesHook(),
			ent.OpCreate,
		),
	}
}

// Edges of the JobTemplate
func (j JobTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		// ensure we cascade delete scheduled jobs when a job template is deleted
		// TODO: if a job template is system owned, we should look into protection to prevent deletion
		// if there are schedule jobs linked
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    j,
			edgeSchema:    ScheduledJob{},
			cascadeDelete: "JobTemplate",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipMutationCreateInput),
			},
		}),
	}
}

// Policy of the JobTemplate
func (JobTemplate) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckCreateAccess(),
			// ensure we check edit access, otherwise you can edit a system owned job template
			entfga.CheckEditAccess[*generated.JobTemplateMutation](),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (JobTemplate) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
