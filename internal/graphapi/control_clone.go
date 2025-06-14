package graphapi

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

type createProgramRequest interface {
	model.CreateFullProgramInput | model.CreateProgramWithMembersInput
}

func getStandardID[T createProgramRequest](value T) string {
	switch input := any(value).(type) {
	case model.CreateProgramWithMembersInput:
		if input.StandardID != nil {
			return *input.StandardID
		}
	case model.CreateFullProgramInput:
		if input.StandardID != nil {
			return *input.StandardID
		}
	}

	return ""
}

// cloneControlsFromStandard clones all controls from a standard into an organization
// it will include all subcontrols
// if the controls already exist in the organization, they will not be cloned again
func (r *mutationResolver) cloneControlsFromStandard(ctx context.Context, standardID string) ([]*generated.Control, error) {
	controls, err := withTransactionalMutation(ctx).Control.Query().Where(
		control.StandardID(standardID)).
		WithStandard().
		WithSubcontrols().
		All(ctx)
	if err != nil {
		return nil, err
	}

	return r.cloneControls(ctx, controls, nil)
}

// cloneControls clones the given controls into the organization in the context
// and optionally links them to a program if programID is given
// if the controls already exist in the organization, they will not be cloned again
// but will be updated to link to the program if needed
func (r *mutationResolver) cloneControls(ctx context.Context, controlsToClone []*generated.Control, programID *string) ([]*generated.Control, error) {
	// keep track of the created control IDs, this includes the ids of controls that already exist in the org
	createdControlIDs := []string{}
	// keep track of the control IDs that already exist in the org to be updated to link to the program if needed
	existingControlIDs := []string{}

	for _, c := range controlsToClone {
		controlInput, standardID := createCloneControlInput(c, programID)

		var newControlID string

		// if a control is already in the org we are cloning to, we should not create it again
		// and instead just link it to the program
		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return nil, rout.NewMissingRequiredFieldError("owner_id")
		}

		existingControl, err := withTransactionalMutation(ctx).Control.Query().
			Where(
				control.RefCode(c.RefCode),
				control.StandardID(standardID),
				control.OwnerID(orgID),
			).
			Only(ctx)

		// check results to determine if we found an existing control or not
		switch {
		case err == nil:
			newControlID = existingControl.ID

			// if the control already exists, add to update the program link later if needed
			existingControlIDs = append(existingControlIDs, newControlID)
		case generated.IsNotFound(err):
			// create new control in the org if it doesn't exist
			res, err := withTransactionalMutation(ctx).Control.Create().
				SetInput(controlInput).Save(ctx)
			if err != nil {
				return nil, err
			}

			newControlID = res.ID
		default:
			log.Error().Err(err).Str("ref_code", c.RefCode).Str("standard_id", c.StandardID).
				Msg("error checking for existing control")

			return nil, err
		}

		createdControlIDs = append(createdControlIDs, newControlID)

		// clone the subcontrols if needed
		if err := r.cloneSubcontrols(ctx, c, newControlID); err != nil {
			return nil, err
		}
	}

	// update the controls to the program if needed
	if len(existingControlIDs) > 0 && programID != nil {
		// if the control already exists, we just link it to the program
		if err := withTransactionalMutation(ctx).Control.Update().
			Where(
				control.IDIn(existingControlIDs...)).
			AddProgramIDs(*programID).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

	// get the cloned controls to return in the response
	query, err := withTransactionalMutation(ctx).Control.Query().Where(control.IDIn(createdControlIDs...)).
		CollectFields(ctx)
	if err != nil {
		return nil, err
	}

	return query.All(ctx)
}

// createCloneControlInput creates a CreateControlInput from the given control that is being cloned
// and returns the input and the standard ID that was set
func createCloneControlInput(c *generated.Control, programID *string) (generated.CreateControlInput, string) {
	controlInput := generated.CreateControlInput{
		// grab fields from the existing control
		Tags:                   c.Tags,
		RefCode:                c.RefCode,
		Description:            &c.Description,
		Source:                 &c.Source,
		ControlType:            &c.ControlType,
		Category:               &c.Category,
		CategoryID:             &c.CategoryID,
		Subcategory:            &c.Subcategory,
		MappedCategories:       c.MappedCategories,
		AssessmentObjectives:   c.AssessmentObjectives,
		ControlQuestions:       c.ControlQuestions,
		ImplementationGuidance: c.ImplementationGuidance,
		ExampleEvidence:        c.ExampleEvidence,
		References:             c.References,
		// set default status to not implemented
		Status: &enums.ControlStatusNotImplemented,
	}

	// set the owner if the control has one
	if c.OwnerID != "" {
		controlInput.OwnerID = &c.OwnerID
	}

	// set the standard information
	var standardID string
	if c.Edges.Standard != nil {
		controlInput.ReferenceFramework = &c.Edges.Standard.ShortName
		standardID = c.Edges.Standard.ID
	}

	if standardID == "" {
		standardID = c.StandardID
	}

	if standardID != "" {
		controlInput.StandardID = &standardID
	}

	if programID != nil {
		controlInput.ProgramIDs = []string{*programID}
	}

	return controlInput, standardID
}

// cloneSubcontrols clones the subcontrols from the given control to the new control ID
// it will only clone subcontrols that do not already exist in the organization
func (r *mutationResolver) cloneSubcontrols(ctx context.Context, c *generated.Control, newControlID string) error {
	if c.Edges.Subcontrols == nil {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	// check to see which subcontrols we need to clone
	// we only want to clone subcontrols that do not already exist in the control
	subcontrolsToClone := []*generated.Subcontrol{}

	for _, s := range c.Edges.Subcontrols {
		// Check if we can find the subcontrol based on refCode and controlID
		// ignore errors here, if we get an error we assume it doesn't exist
		exists, _ := withTransactionalMutation(ctx).Subcontrol.Query().
			Where(
				subcontrol.RefCode(s.RefCode),
				subcontrol.ControlID(newControlID),
				subcontrol.OwnerID(orgID),
			).
			Exist(ctx)
		if !exists {
			subcontrolsToClone = append(subcontrolsToClone, s)
		}
	}

	subcontrols := make([]*generated.CreateSubcontrolInput, len(subcontrolsToClone))

	for j, subcontrol := range subcontrolsToClone {
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
			ReferenceFramework:     subcontrol.ReferenceFramework,
			Status:                 &enums.ControlStatusNotImplemented,
		}
	}

	_, err = r.bulkCreateSubcontrol(ctx, subcontrols)

	return err
}
