package graphapi

import (
	"context"
	"sync"

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
	type controlKey struct {
		ref, std string
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	type entry struct {
		src        *generated.Control
		input      generated.CreateControlInput
		standardID string
		newID      string
	}

	entries := make([]*entry, len(controlsToClone))
	refCodes := make([]string, len(controlsToClone))

	standardSet := map[string]struct{}{}

	for i, c := range controlsToClone {
		in, stdID := createCloneControlInput(c, programID)
		entries[i] = &entry{src: c, input: in, standardID: stdID}
		refCodes[i] = c.RefCode
		standardSet[stdID] = struct{}{}
	}

	standardIDs := make([]string, 0, len(standardSet))
	for id := range standardSet {
		standardIDs = append(standardIDs, id)
	}

	existingControls, err := r.db.Control.Query().
		Where(
			control.RefCodeIn(refCodes...),
			control.StandardIDIn(standardIDs...),
			control.OwnerID(orgID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	existingMap := map[controlKey]*generated.Control{}
	for _, ec := range existingControls {
		existingMap[controlKey{ref: ec.RefCode, std: ec.StandardID}] = ec
	}

	var createInputs []*generated.CreateControlInput
	var createEntries []*entry
	createdControlIDs := []string{}
	existingControlIDs := []string{}

	for _, e := range entries {
		key := controlKey{ref: e.src.RefCode, std: e.standardID}
		if ex, ok := existingMap[key]; ok {
			e.newID = ex.ID
			createdControlIDs = append(createdControlIDs, ex.ID)
			existingControlIDs = append(existingControlIDs, ex.ID)
			continue
		}

		createInputs = append(createInputs, &e.input)
		createEntries = append(createEntries, e)
	}

	if len(createInputs) > 0 {
		res, err := r.bulkCreateControl(ctx, createInputs)
		if err != nil {
			return nil, err
		}

		for i, c := range res.Controls {
			createEntries[i].newID = c.ID
			createdControlIDs = append(createdControlIDs, c.ID)
		}
	}

	funcs := make([]func(), len(entries))
	var (
		errors []error
		mu     sync.Mutex
	)

	for i, e := range entries {
		e := e
		funcs[i] = func() {
			if err := r.cloneSubcontrols(ctx, e.src, e.newID); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}
	}

	r.withPool().SubmitMultipleAndWait(funcs)

	if len(errors) > 0 {
		log.Error().Errs("errors", errors).
			Msgf("error cloning controls")

		if len(createEntries) > 0 {
			ids := make([]string, 0, len(createEntries))
			for _, e := range createEntries {
				if e.newID != "" {
					ids = append(ids, e.newID)
				}
			}

			if len(ids) > 0 {
				log.Warn().Msgf("error cloning controls, deleting %d controls that were created before the error occurred", len(ids))

				if _, err := withTransactionalMutation(ctx).Control.Delete().
					Where(control.IDIn(ids...)).
					Exec(ctx); err != nil {
					log.Error().Err(err).Msg("error deleting controls that were created before the error occurred")
				}
			}
		}

		return nil, errors[0]
	}

	if len(existingControlIDs) > 0 && programID != nil {
		if err := withTransactionalMutation(ctx).Control.Update().
			Where(control.IDIn(existingControlIDs...)).
			AddProgramIDs(*programID).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

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

	refCodes := []string{}
	for _, s := range c.Edges.Subcontrols {
		refCodes = append(refCodes, s.RefCode)
	}

	// check if we can find the subcontrol based on refCode and controlID
	existingSubcontrols, err := r.db.Subcontrol.Query().
		Where(
			subcontrol.RefCodeIn(refCodes...),
			subcontrol.ControlID(newControlID),
			subcontrol.OwnerID(orgID),
		).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error checking for existing subcontrols")

		return err
	}

	// get the subcontrols we actually need to clone
	for _, subcontrol := range c.Edges.Subcontrols {
		// check if the subcontrol already exists in the organization
		exists := false
		for _, existingSubcontrol := range existingSubcontrols {
			if existingSubcontrol.RefCode == subcontrol.RefCode && existingSubcontrol.ControlID == newControlID {
				exists = true
				break
			}
		}

		if !exists {
			// if it does not exist, we need to clone it
			subcontrolsToClone = append(subcontrolsToClone, subcontrol)
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
			Status:                 &enums.ControlStatusNotImplemented,
		}
	}

	_, err = r.bulkCreateSubcontrolNoTransaction(ctx, subcontrols)

	return err
}

// bulkCreateSubcontrolNoTransaction creates multiple subcontrols in a single request without a transaction to allow it to be run in parallel
func (r *mutationResolver) bulkCreateSubcontrolNoTransaction(ctx context.Context, input []*generated.CreateSubcontrolInput) (*model.SubcontrolBulkCreatePayload, error) {
	builders := make([]*generated.SubcontrolCreate, len(input))
	for i, data := range input {
		builders[i] = r.db.Subcontrol.Create().SetInput(*data)
	}

	res, err := r.db.Subcontrol.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "subcontrol"})
	}

	// return response
	return &model.SubcontrolBulkCreatePayload{
		Subcontrols: res,
	}, nil
}
