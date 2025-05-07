package graphapi

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/pkg/enums"
)

func (r *mutationResolver) cloneControls(ctx context.Context, existingControls []*generated.Control, programID *string, ignoreStandard bool) ([]*generated.Control, error) {
	createdControlIDs := make([]string, len(existingControls))

	for _, control := range existingControls {
		mappedControlIDs := []string{}
		mappedControls := control.Edges.MappedControls
		for _, mc := range mappedControls {
			mappedControlIDs = append(mappedControlIDs, mc.ID)
		}

		controlInput := generated.CreateControlInput{
			Tags:                   control.Tags,
			RefCode:                control.RefCode,
			Description:            &control.Description,
			Source:                 &control.Source,
			ControlType:            &control.ControlType,
			Category:               &control.Category,
			CategoryID:             &control.CategoryID,
			Subcategory:            &control.Subcategory,
			MappedCategories:       control.MappedCategories,
			AssessmentObjectives:   control.AssessmentObjectives,
			ControlQuestions:       control.ControlQuestions,
			ImplementationGuidance: control.ImplementationGuidance,
			ExampleEvidence:        control.ExampleEvidence,
			References:             control.References,
			MappedControlIDs:       mappedControlIDs,
			Status:                 &enums.ControlStatusPreparing,
		}

		if !ignoreStandard && control.StandardID != "" {
			controlInput.StandardID = &control.StandardID
		}

		if programID != nil {
			controlInput.ProgramIDs = []string{*programID}
		}

		res, err := withTransactionalMutation(ctx).Control.Create().SetInput(controlInput).Save(ctx)
		if err != nil {
			return nil, parseRequestError(err, action{action: ActionCreate, object: "control"})
		}

		createdControlIDs = append(createdControlIDs, res.ID)

		if err := r.cloneSubcontrols(ctx, control, res.ID); err != nil {
			return nil, err
		}
	}

	// get the created controls
	query, err := withTransactionalMutation(ctx).Control.Query().Where(control.IDIn(createdControlIDs...)).CollectFields(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "control"})
	}

	return query.All(ctx)
}

func (r *mutationResolver) cloneSubcontrols(ctx context.Context, control *generated.Control, newControlID string) error {
	if control.Edges.Subcontrols == nil {
		return nil
	}

	mappedControlIDs := []string{}
	mappedControls := control.Edges.MappedControls
	for _, mc := range mappedControls {
		mappedControlIDs = append(mappedControlIDs, mc.ID)
	}

	subcontrols := make([]*generated.CreateSubcontrolInput, len(control.Edges.Subcontrols))
	for j, subcontrol := range control.Edges.Subcontrols {
		subcontrols[j] = &generated.CreateSubcontrolInput{
			Tags:                   subcontrol.Tags,
			RefCode:                subcontrol.RefCode,
			Description:            &subcontrol.Description,
			Source:                 &subcontrol.Source,
			ControlID:              newControlID,
			ControlType:            &subcontrol.ControlType,
			Category:               &subcontrol.Category,
			CategoryID:             &subcontrol.CategoryID,
			Subcategory:            &subcontrol.Subcategory,
			MappedCategories:       subcontrol.MappedCategories,
			AssessmentObjectives:   subcontrol.AssessmentObjectives,
			ControlQuestions:       subcontrol.ControlQuestions,
			ImplementationGuidance: subcontrol.ImplementationGuidance,
			ExampleEvidence:        subcontrol.ExampleEvidence,
			References:             subcontrol.References,
			MappedControlIDs:       mappedControlIDs,
			Status:                 &enums.ControlStatusPreparing,
		}
	}

	_, err := r.bulkCreateSubcontrol(ctx, subcontrols)
	return err
}
