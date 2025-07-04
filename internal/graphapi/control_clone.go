package graphapi

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

type subcontrolToCreate struct {
	NewControlID string
	RefControl   *generated.Control
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
	// track subcontrols to create
	subcontrolsToCreate := []subcontrolToCreate{}

	// do this in a go-routine to allow multiple controls to be cloned in parallel, use pond for this
	// we cannot use a transaction here because we are running multiple go-routines
	// and transactions cannot be used across go-routines
	funcs := make([]func(), len(controlsToClone))
	var (
		errors []error
		mu     sync.Mutex
	)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	wherePredicate := []predicate.Control{}
	for _, c := range controlsToClone {
		wherePredicate = append(wherePredicate, control.Or(
			control.RefCode(c.RefCode),
			control.StandardID(c.StandardID),
		))
	}

	// get existing controls that match the refCode and standardID
	// skip the access checks for the controls, we are already filtering on organization id
	// and controls are visible to users in the organization
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	existingControls, err := r.db.Control.Query().
		Where(
			control.And(
				append(
					[]predicate.Control{
						control.DeletedAtIsNil(),
						control.OwnerID(orgID),
					},
					wherePredicate...,
				)...,
			),
		).
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		All(allowCtx)

	if err != nil {
		log.Error().Err(err).Msg("error checking for existing controls")
		return nil, err
	}

	// find the ones we do need to clone
	updateControlsToClone := []*generated.Control{}
	for _, c := range controlsToClone {
		// check if the control already exists in the organization
		exists := false
		for _, existingControl := range existingControls {
			if existingControl.RefCode == c.RefCode && existingControl.StandardID == c.StandardID {

				// control already exists, we will not clone it again
				existingControlIDs = append(existingControlIDs, existingControl.ID)
				exists = true
			}
		}

		if !exists {
			// control does not exist, we will clone it
			updateControlsToClone = append(updateControlsToClone, c)
		}
	}

	// create a function for each control to clone
	// this will allow us to run the cloning in parallel
	// we will use a mutex to protect the createdControlIDs and existingControlIDs slices
	// and the errors slice
	for i, c := range updateControlsToClone {
		c := c // capture loop variable
		funcs[i] = func() {
			controlInput, _ := createCloneControlInput(c, programID)

			res, err := r.db.Control.Create().
				SetInput(controlInput).Save(ctx)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				return
			}

			newControlID := res.ID

			mu.Lock()
			createdControlIDs = append(createdControlIDs, newControlID)

			// add subcontrols to create if they exist
			if len(c.Edges.Subcontrols) > 0 {
				subcontrolsToCreate = append(subcontrolsToCreate, subcontrolToCreate{
					NewControlID: newControlID,
					RefControl:   c,
				})
			}
			mu.Unlock()
		}
	}

	// run the cloning functions in parallel
	r.withPool().SubmitMultipleAndWait(funcs)

	// check if there were any errors during the cloning process
	if len(errors) > 0 {
		// return the first error encountered
		log.Error().Errs("errors", errors).
			Msgf("error cloning controls")

		// delete any controls that were created before the error occurred
		if len(createdControlIDs) > 0 {
			log.Warn().Msgf("error cloning controls, deleting %d controls that were created before the error occurred", len(createdControlIDs))

			// delete any controls that were created before the error occurred
			// this should also cascade delete any subcontrols that were created
			if _, err := withTransactionalMutation(ctx).Control.Delete().
				Where(control.IDIn(createdControlIDs...)).
				Exec(ctx); err != nil {

				log.Error().Err(err).Msg("error deleting controls that were created before the error occurred")
			}
		}

		// we can return the first error encountered, as the rest will be logged
		return nil, errors[0]
	}

	// allow the subcontrols to be cloned without a parent check, this is because the same user already created the control
	// and the subcontrols are part of the control
	if err := r.cloneSubcontrols(allowCtx, subcontrolsToCreate); err != nil {
		log.Error().Err(err).Msg("error cloning subcontrols, rolling back controls that were created before the error occurred")

		if _, err := withTransactionalMutation(ctx).Control.Delete().
			Where(control.IDIn(createdControlIDs...)).
			Exec(ctx); err != nil {

			return nil, err
		}

		return nil, err
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

	// add existingControlIDs to createdControlIDs
	createdControlIDs = append(createdControlIDs, existingControlIDs...)

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

	if c.Edges.Standard != nil {
		// if the control has a standard, we will set the reference framework to the standard
		controlInput.ReferenceFramework = &c.Edges.Standard.ShortName
	}

	// set the owner if the control has one
	if c.OwnerID != "" {
		controlInput.OwnerID = &c.OwnerID
	}

	// set the standard information
	var standardID string
	if c.Edges.Standard != nil {
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
func (r *mutationResolver) cloneSubcontrols(ctx context.Context, subcontrolsToCreate []subcontrolToCreate) error {
	if len(subcontrolsToCreate) == 0 {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	// check to see which subcontrols we need to clone
	// we only want to clone subcontrols that do not already exist in the control
	subcontrolsToClone := []*generated.Subcontrol{}

	wherePredicate := []predicate.Subcontrol{}

	for _, c := range subcontrolsToCreate {
		// get all the refCodes for the subcontrols in the control
		refCodes := []string{}
		for _, s := range c.RefControl.Edges.Subcontrols {
			refCodes = append(refCodes, s.RefCode)
		}

		where := subcontrol.Or(
			subcontrol.RefCodeIn(refCodes...),
			subcontrol.ControlID(c.NewControlID),
		)

		wherePredicate = append(wherePredicate, where)
	}

	// check if we can find the subcontrol based on refCode and controlID
	existingSubcontrols, err := r.db.Subcontrol.Query().
		Where(
			subcontrol.And(
				append(
					[]predicate.Subcontrol{
						subcontrol.DeletedAtIsNil(),
						subcontrol.OwnerID(orgID),
					},
					wherePredicate...,
				)...,
			),
		).
		Select(subcontrol.FieldRefCode, subcontrol.FieldControlID).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error checking for existing subcontrols")

		return err
	}

	// get the subcontrols we actually need to clone
	for _, c := range subcontrolsToCreate {
		for _, toCloneSubcontrol := range c.RefControl.Edges.Subcontrols {
			// check if the subcontrol already exists in the organization
			exists := false
			for _, existingSubcontrol := range existingSubcontrols {
				if existingSubcontrol.RefCode == toCloneSubcontrol.RefCode &&
					existingSubcontrol.ControlID == c.NewControlID {
					exists = true
					break
				}
			}

			if !exists {
				// add the subcontrol to the list of subcontrols to clone
				toCloneSubcontrol.ControlID = c.NewControlID

				if c.RefControl.Edges.Standard != nil {
					toCloneSubcontrol.ReferenceFramework = &c.RefControl.Edges.Standard.ShortName
				}

				subcontrolsToClone = append(subcontrolsToClone, toCloneSubcontrol)
			}
		}
	}

	subcontrols := make([]*generated.CreateSubcontrolInput, len(subcontrolsToClone))

	for j, subcontrol := range subcontrolsToClone {
		subcontrols[j] = &generated.CreateSubcontrolInput{
			Tags:                   subcontrol.Tags,
			RefCode:                subcontrol.RefCode,
			Description:            &subcontrol.Description,
			Source:                 &subcontrol.Source,
			ControlID:              subcontrol.ControlID,
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
			Status:                 &enums.ControlStatusNotImplemented,
			OwnerID:                &subcontrol.OwnerID,
			ReferenceFramework:     subcontrol.ReferenceFramework,
		}
	}

	return r.bulkCreateSubcontrolNoTransaction(ctx, subcontrols)
}

// bulkCreateSubcontrolNoTransaction creates multiple subcontrols in a single request without a transaction to allow it to be run in parallel
func (r *mutationResolver) bulkCreateSubcontrolNoTransaction(ctx context.Context, input []*generated.CreateSubcontrolInput) error {
	builders := make([]*generated.SubcontrolCreate, len(input))
	for i, data := range input {
		builders[i] = r.db.Subcontrol.Create().SetInput(*data)
	}

	if err := r.db.Subcontrol.CreateBulk(builders...).Exec(ctx); err != nil {
		log.Error().Err(err).Msg("error creating subcontrols in bulk")

		return parseRequestError(err, action{action: ActionCreate, object: "subcontrol"})
	}

	// return response
	return nil
}
