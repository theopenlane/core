package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/slateparser"
)

// ErrTextContainsComments is returned when attempting to set a text field with a corresponding JSON field that contains comments, this will cause the comment links to be lost in conversion and is not allowed
var ErrTextContainsComments = errors.New("text contains comments, unable to set description due to potential loss of data in conversion")

// detailsWithJSONMutation is a mutation interface for entities with details and details_json fields
// this is used for the document mixin schemas and control implementation schema
type detailsWithJSONMutation interface {
	Details() (string, bool)
	DetailsJSON() ([]interface{}, bool)
	OldDetailsJSON(ctx context.Context) (v []interface{}, err error)
	ClearDetailsJSON()
}

// descriptionWithJSONMutation is a mutation interface for entities with description and description_json fields, used by the control mixin
type descriptionWithJSONMutation interface {
	Description() (string, bool)
	DescriptionJSON() ([]interface{}, bool)
	OldDescriptionJSON(ctx context.Context) (v []interface{}, err error)
	ClearDescriptionJSON()
}

// riskWithJSONMutation is a mutation interface for the risk schema with mitigation, business_costs and details fields, along with their corresponding JSON fields, used by the risk schema
type riskWithJSONMutation interface {
	Mitigation() (string, bool)
	MitigationJSON() ([]interface{}, bool)
	OldMitigationJSON(ctx context.Context) (v []interface{}, err error)
	ClearMitigationJSON()

	BusinessCosts() (string, bool)
	BusinessCostsJSON() ([]interface{}, bool)
	OldBusinessCostsJSON(ctx context.Context) (v []interface{}, err error)
	ClearBusinessCostsJSON()

	detailsWithJSONMutation
}

// controlObjectiveWithJSONMutation is a mutation interface for the control implementation schema with desired_outcome, along with their corresponding JSON fields, used by the control implementation schema
type controlObjectiveWithJSONMutation interface {
	DesiredOutcome() (string, bool)
	DesiredOutcomeJSON() ([]interface{}, bool)
	OldDesiredOutcomeJSON(ctx context.Context) (v []interface{}, err error)
	ClearDesiredOutcomeJSON()
}

// noteWithJSONMutation is a mutation interface for the note schema with text and text_json fields
type noteWithJSONMutation interface {
	Text() (string, bool)
	TextJSON() ([]interface{}, bool)
	OldTextJSON(ctx context.Context) (v []interface{}, err error)
	ClearTextJSON()
}

// HookSlateJSON is an ent hook that will handle clearing JSON fields if description is set, this will
// prevent stale JSON data from remaining when a user sets description via the API (or bulk csv operations)
// that does not include the JSON field data
func HookSlateJSON() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// ignore soft delete operations
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			switch mut := m.(type) {
			case *generated.RiskMutation:
				if err := setRiskJSONFields(ctx, mut); err != nil {
					return nil, err
				}
			case *generated.ControlObjectiveMutation:
				if err := setControlObjectiveJSONFields(ctx, mut); err != nil {
					return nil, err
				}
			case *generated.NoteMutation:
				if err := setNoteJSONFields(ctx, mut); err != nil {
					return nil, err
				}
			case *generated.ControlMutation, *generated.SubcontrolMutation:
				if err := setDescriptionJSONFields(ctx, mut); err != nil {
					return nil, err
				}
			default:
				// fallback to document JSON fields, this is used across many schemas
				// including all with the document mixin and the control implementation schema
				if err := setDetailsJSONFields(ctx, mut); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.HasOp(ent.OpUpdate|ent.OpUpdateOne),
	)
}

// setRiskJSONFields sets the slate JSON fields for the risk schema
func setRiskJSONFields(ctx context.Context, mut ent.Mutation) error {
	riskMut, ok := mut.(riskWithJSONMutation)
	if !ok {
		return nil
	}

	if _, ok := riskMut.Details(); ok {
		detailsJSON, ok := riskMut.DetailsJSON()
		if !ok || detailsJSON == nil {
			if oldJSON, err := riskMut.OldDetailsJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			riskMut.ClearDetailsJSON()
		}
	}

	if _, ok := riskMut.Mitigation(); ok {
		mitigationJSON, ok := riskMut.MitigationJSON()
		if !ok || mitigationJSON == nil {
			if oldJSON, err := riskMut.OldMitigationJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			riskMut.ClearMitigationJSON()
		}
	}

	if _, ok := riskMut.BusinessCosts(); ok {
		businessCostsJSON, ok := riskMut.BusinessCostsJSON()
		if !ok || businessCostsJSON == nil {
			if oldJSON, err := riskMut.OldBusinessCostsJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			riskMut.ClearBusinessCostsJSON()
		}
	}

	return nil
}

// setControlObjectiveJSONFields sets the slate JSON fields for the control objective schema
func setControlObjectiveJSONFields(ctx context.Context, mut ent.Mutation) error {
	ciMut, ok := mut.(controlObjectiveWithJSONMutation)
	if !ok {
		return nil
	}

	if _, ok := ciMut.DesiredOutcome(); ok {
		ciMutJSON, ok := ciMut.DesiredOutcomeJSON()
		if !ok || ciMutJSON == nil {
			if oldJSON, err := ciMut.OldDesiredOutcomeJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			ciMut.ClearDesiredOutcomeJSON()

		}
	}

	return nil
}

// setNoteJSONFields sets the slate JSON fields for the note schema
func setNoteJSONFields(ctx context.Context, mut ent.Mutation) error {
	noteMut, ok := mut.(noteWithJSONMutation)
	if !ok {
		return nil
	}

	if _, ok := noteMut.Text(); ok {
		textJSON, ok := noteMut.TextJSON()
		if !ok || textJSON == nil {
			if oldJSON, err := noteMut.OldTextJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			noteMut.ClearTextJSON()
		}
	}

	return nil
}

// setDescriptionJSONFields sets the slate JSON fields for mutations with description fields
func setDescriptionJSONFields(ctx context.Context, mut ent.Mutation) error {
	docMut, ok := mut.(descriptionWithJSONMutation)
	if !ok {
		return nil
	}

	if _, ok := docMut.Description(); ok {
		descJSON, ok := docMut.DescriptionJSON()
		if !ok || descJSON == nil {
			if oldJSON, err := docMut.OldDescriptionJSON(ctx); err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}

			docMut.ClearDescriptionJSON()
		}
	}

	return nil
}

// setDetailsJSONFields sets the slate JSON fields for mutations with details fields
func setDetailsJSONFields(ctx context.Context, mut ent.Mutation) error {
	detailsMut, ok := mut.(detailsWithJSONMutation)
	if !ok {
		return nil
	}

	if _, ok := detailsMut.Details(); ok {
		detailsJSON, ok := detailsMut.DetailsJSON()
		if !ok || detailsJSON == nil {
			oldJSON, err := detailsMut.OldDetailsJSON(ctx)
			if err == nil && oldJSON != nil {
				// check if old JSON contains any comments
				if slateparser.ContainsCommentsInTextJSON(oldJSON) {
					return ErrTextContainsComments
				}
			}
			detailsMut.ClearDetailsJSON()
		}
	}

	return nil
}
