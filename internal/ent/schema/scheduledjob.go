package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
)

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJob is the name of the schema in snake case
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
		field.String("job_id").
			NotEmpty().
			Comment("the scheduled_job id to take the script to run from"),
		field.Bool("active").
			Default(true).
			Comment("whether the scheduled job is active"),
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
			Comment("cron 6-field syntax, defaults to the job template's cron if not provided").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput |
					entgql.SkipOrderField,
				),
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
		field.String("job_runner_id").
			Optional().
			Comment("the runner that this job will run on. If not set, it will scheduled on a general runner instead"),
		// TODO: we should consider adding a hook that will round-robin the orgs runners
		// or instead of setting a runner, say "org_runner" so we know they are using their own runner vs. a potentially
		// openlane shared runner; we shouldn't have to have the user (or UI) look up this ID every time to create a job
	}
}

// Mixin of the ScheduledJob
func (c ScheduledJob) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:      "SJB",
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.ScheduledJob](c,
				withParents(Control{}, Subcontrol{}),
				withOrganizationOwner(true),
				withSkipForSystemAdmin(true),
			),
		},
	}.getMixins(c)
}

// Edges of the ScheduledJob
func (c ScheduledJob) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: JobTemplate{},
			field:      "job_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(JobTemplate{}.Name()),
			},
		}),

		defaultEdgeToWithPagination(c, Control{}),
		defaultEdgeToWithPagination(c, Subcontrol{}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			required:   false,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the ScheduledJob
func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJob
func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// this schema shouldn't be need to be searchable, most results would be the job template,
		// not the scheduled job itself
		entx.SchemaSearchable(false),
	}
}

// Hooks of the ScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookScheduledJobCreate(),
	}
}

// Interceptors of the ScheduledJob
func (ScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ScheduledJob
func (ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Control{}.PluralName(),
				Subcontrol{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (ScheduledJob) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
