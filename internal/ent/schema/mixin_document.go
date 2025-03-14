package schema

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
)

// DocumentMixin implements the document pattern with approver for schemas.
type DocumentMixin struct {
	mixin.Schema

	DocumentType string
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
	// TODO (sfunk): add hook to update permissions for the approver
	return []ent.Hook{}
}

// getDocumentFields returns the fields for a document type schema
// for example a policy or procedure
func getDocumentFields(documentType string) []ent.Field {
	return []ent.Field{field.String("name").
		Comment(fmt.Sprintf("the name of the %s", documentType)).
		Annotations(
			entx.FieldSearchable(),
		).
		NotEmpty(),
		field.Enum("status").
			GoType(enums.DocumentStatus("")).
			Default(enums.DocumentDraft.String()).
			Optional().
			Comment(fmt.Sprintf("status of the %s, e.g. draft, published, archived, etc.", documentType)),
		field.String(fmt.Sprintf("%s_type", documentType)).
			Optional().
			Comment(fmt.Sprintf("type of the %s, e.g. compliance, operational, health and safety, etc.", documentType)),
		field.Text("details").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment(fmt.Sprintf("details of the %s", documentType)),
		field.Bool("approval_required").
			Default(true).
			Optional().
			Comment(fmt.Sprintf("whether approval is required for edits to the %s", documentType)),
		field.Time("review_due").
			Default(time.Now().AddDate(1, 0, 0)).
			Optional().
			Comment(fmt.Sprintf("the date the %s should be reviewed, calculated based on the review_frequency if not directly set", documentType)),
		field.Enum("review_frequency").
			Optional().
			GoType(enums.Frequency("")).
			Default(enums.FrequencyYearly.String()).
			Comment(fmt.Sprintf("the frequency at which the %s should be reviewed, used to calculate the review_due date", documentType)),
	}
}

func getApproverEdges(documentType string) []ent.Edge {
	return []ent.Edge{
		edge.To("approver", Group.Type).
			Unique().
			Comment(fmt.Sprintf("the group of users who are responsible for approving the %s", documentType)),
		edge.To("delegate", Group.Type).
			Unique().
			Comment(fmt.Sprintf("temporary delegates for the %s, used for temporary approval", documentType)),
	}
}
