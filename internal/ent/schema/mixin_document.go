package schema

import (
	"fmt"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
)

// DocumentMixin implements the document pattern with approver for schemas.
type DocumentMixin struct {
	mixin.Schema

	DocumentType string
}

// NewDocumentMixin creates a new DocumentMixin with the given schema
// the schema must implement the SchemaFuncs interface
func NewDocumentMixin(schema any) DocumentMixin {
	sch := toSchemaFuncs(schema)

	return DocumentMixin{
		DocumentType: sch.Name(),
	}
}

// Fields of the DocumentMixin.
func (d DocumentMixin) Fields() []ent.Field {
	return getDocumentFields(d.DocumentType)
}

// Edges of the DocumentMixin.
func (d DocumentMixin) Edges() []ent.Edge {
	return getApproverEdges(d.DocumentType)
}

// Hooks of the DocumentMixin.
func (d DocumentMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"approver": "group",
			}, "approver"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"delegate": "group",
			}, "delegate"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hooks.HookImportDocument(),
		hooks.HookSummarizeDetails(),
		hook.On(
			hooks.HookStatusApproval(),
			ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne,
		),
	}
}

// getDocumentFields returns the fields for a document type schema
// for example a policy or procedure
func getDocumentFields(documentType string) []ent.Field {
	return []ent.Field{field.String("name").
		Comment(fmt.Sprintf("the name of the %s", documentType)).
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("name"),
		).
		NotEmpty(),
		field.Enum("status").
			GoType(enums.DocumentStatus("")).
			Default(enums.DocumentDraft.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Optional().
			Comment(fmt.Sprintf("status of the %s, e.g. draft, published, archived, etc.", documentType)),
		field.Text("details").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment(fmt.Sprintf("details of the %s", documentType)),
		field.JSON("details_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment(fmt.Sprintf("structured details of the %s in JSON format", documentType)),
		field.Bool("approval_required").
			Default(true).
			Optional().
			Comment(fmt.Sprintf("whether approval is required for edits to the %s", documentType)),
		field.Time("review_due").
			Default(time.Now().AddDate(1, 0, 0)).
			Annotations(
				entgql.OrderField("review_due"),
			).
			Optional().
			Comment(fmt.Sprintf("the date the %s should be reviewed, calculated based on the review_frequency if not directly set", documentType)),
		field.Enum("review_frequency").
			Optional().
			GoType(enums.Frequency("")).
			Annotations(
				entgql.OrderField("REVIEW_FREQUENCY"),
			).
			Default(enums.FrequencyYearly.String()).
			Comment(fmt.Sprintf("the frequency at which the %s should be reviewed, used to calculate the review_due date", documentType)),
		field.String("approver_id").
			Optional().
			Unique().
			Comment(fmt.Sprintf("the id of the group responsible for approving the %s", documentType)).
			StructTag(`json:"approver_id,omitempty"`),
		field.String("delegate_id").
			Optional().
			Unique().
			Comment(fmt.Sprintf("the id of the group responsible for approving the %s", documentType)),
		field.String("summary").
			Optional().
			Annotations(
				entgql.Skip(^entgql.SkipType),
			),
		field.Strings("tag_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("auto-generated tag suggestions for the %s", documentType)),
		field.Strings("dismissed_tag_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("tag suggestions dismissed by the user for the %s", documentType)),
		field.Strings("control_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("proposed controls referenced in the %s", documentType)),
		field.Strings("dismissed_control_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("control suggestions dismissed by the user for the %s", documentType)),
		field.Strings("improvement_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("suggested improvements for the %s", documentType)),
		field.Strings("dismissed_improvement_suggestions").
			Optional().
			Default([]string{}).
			Comment(fmt.Sprintf("improvement suggestions dismissed by the user for the %s", documentType)),
		field.String("url").
			Comment(fmt.Sprintf("This will contain the url used to create or update the %s", documentType)).
			Optional().
			Nillable(),
		field.String("file_id").
			Comment(fmt.Sprintf("This will contain the most recent file id if this %s was created from a file", documentType)).
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
	}
}

func getApproverEdges(documentType string) []ent.Edge {
	return []ent.Edge{
		edge.To("approver", Group.Type).
			Unique().
			Field("approver_id").
			Annotations(
				accessmap.EdgeViewCheck(Group{}.Name()),
				entx.FieldWorkflowEligible(),
			).
			Comment(fmt.Sprintf("the group of users who are responsible for approving the %s", documentType)),
		edge.To("delegate", Group.Type).
			Unique().
			Field("delegate_id").
			Annotations(
				accessmap.EdgeViewCheck(Group{}.Name()),
				entx.FieldWorkflowEligible(),
			).
			Comment(fmt.Sprintf("temporary delegates for the %s, used for temporary approval", documentType)),
	}
}
