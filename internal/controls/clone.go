package controls

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
)

// SubcontrolToCreate is used to track which subcontrols need to be created for a given control
type SubcontrolToCreate struct {
	NewControlID string
	RefControl   *generated.Control
}

// ControlToUpdate is used to track existing controls that need to be updated due to changes
// in the revision of their connected standards
type ControlToUpdate struct {
	ExistingControlID string
	SourceControl     *generated.Control
}

// CloneControls clones the given controls with the provided options and returns the IDs of the created controls, the subcontrols to create, and any errors that occurred during cloning
func CloneControls(ctx context.Context, client *generated.Client, controlsToClone []*generated.Control, opts ...CloneOption) ([]string, []SubcontrolToCreate, error) {
	if client == nil {
		return nil, nil, nil
	}

	// apply options
	options := CloneOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	// do this in a go-routine to allow multiple controls to be cloned in parallel, use the worker pool for this
	// we cannot use a transaction here because we are running multiple go-routines
	// and transactions cannot be used across go-routines
	var (
		errors []error
		mu     sync.Mutex
	)

	// keep track of the created control IDs, this includes the ids of controls that already exist in the org
	createdControlIDs := []string{}
	// track subcontrols to create
	subcontrolsToCreate := []SubcontrolToCreate{}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// create a function for each control to clone
	// this will allow us to run the cloning in parallel
	// we will use a mutex to protect the createdControlIDs and existingControlIDs slices
	// and the errors slice
	funcs := make([]func(), len(controlsToClone))

	for i, c := range controlsToClone {
		funcs[i] = func() {
			controlInput, isTCControl := CreateCloneControlInput(c, options.programID, options.orgID)

			res, err := client.Control.Create().
				SetInput(controlInput).
				SetIsTrustCenterControl(isTCControl).Save(allowCtx)
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
				subcontrolsToCreate = append(subcontrolsToCreate, SubcontrolToCreate{
					NewControlID: newControlID,
					RefControl:   c,
				})
			}

			mu.Unlock()
		}
	}

	// run the cloning functions in parallel
	if err := client.Pool.SubmitMultipleAndWait(funcs); err != nil {
		return nil, nil, err
	}

	// check if there were any errors during the cloning process
	if len(errors) > 0 {
		// return the first error encountered
		logx.FromContext(ctx).Error().Errs("errors", errors).
			Msgf("error cloning controls, deleting %d controls that were created before the error occurred", len(createdControlIDs))

		// delete any controls that were created before the error occurred
		if len(createdControlIDs) > 0 {
			// delete any controls that were created before the error occurred
			// this should also cascade delete any subcontrols that were created
			if _, err := client.Control.Delete().
				Where(control.IDIn(createdControlIDs...)).
				Exec(allowCtx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error deleting controls that were created before the error occurred")
			}
		}

		// we can return the first error encountered, as the rest will be logged
		return nil, nil, errors[0]
	}

	return createdControlIDs, subcontrolsToCreate, nil
}

// CreateCloneControlInput creates a CreateControlInput from the given control that is being cloned
// and returns the input, the standard ID that was set, and whether the control is a trust center control
func CreateCloneControlInput(c *generated.Control, programID *string, orgID string) (generated.CreateControlInput, bool) {
	controlInput := generated.CreateControlInput{
		// grab fields from the existing control
		RefCode:                c.RefCode,
		Title:                  &c.Title,
		Aliases:                c.Aliases,
		Description:            &c.Description,
		Source:                 &c.Source,
		Category:               &c.Category,
		CategoryID:             &c.CategoryID,
		Subcategory:            &c.Subcategory,
		MappedCategories:       c.MappedCategories,
		AssessmentObjectives:   c.AssessmentObjectives,
		AssessmentMethods:      c.AssessmentMethods,
		ControlQuestions:       c.ControlQuestions,
		ImplementationGuidance: c.ImplementationGuidance,
		ExampleEvidence:        c.ExampleEvidence,
		References:             c.References,
		TestingProcedures:      c.TestingProcedures,
		EvidenceRequests:       c.EvidenceRequests,
		// set default status to not implemented
		Status:  &enums.ControlStatusNotImplemented,
		OwnerID: &orgID,
	}

	if c.Edges.Standard != nil {
		// if the control has a standard, we will set the reference framework to the standard
		controlInput.ReferenceFramework = &c.Edges.Standard.ShortName
		controlInput.ReferenceFrameworkRevision = &c.Edges.Standard.Revision
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

	return controlInput, isTrustCenterStandard(c.Edges.Standard)
}

// CreateCloneSubcontrolInput creates a CreateSubcontrolInput from the given subcontrol that is being cloned
func CreateCloneSubcontrolInput(subcontrol *generated.Subcontrol, orgID string) *generated.CreateSubcontrolInput {
	return &generated.CreateSubcontrolInput{
		RefCode:                    subcontrol.RefCode,
		Title:                      &subcontrol.Title,
		Description:                &subcontrol.Description,
		Source:                     &subcontrol.Source,
		ControlID:                  subcontrol.ControlID,
		Category:                   &subcontrol.Category,
		CategoryID:                 &subcontrol.CategoryID,
		Subcategory:                &subcontrol.Subcategory,
		MappedCategories:           subcontrol.MappedCategories,
		AssessmentObjectives:       subcontrol.AssessmentObjectives,
		AssessmentMethods:          subcontrol.AssessmentMethods,
		ControlQuestions:           subcontrol.ControlQuestions,
		ImplementationGuidance:     subcontrol.ImplementationGuidance,
		ExampleEvidence:            subcontrol.ExampleEvidence,
		TestingProcedures:          subcontrol.TestingProcedures,
		EvidenceRequests:           subcontrol.EvidenceRequests,
		References:                 subcontrol.References,
		Status:                     &enums.ControlStatusNotImplemented,
		ReferenceFramework:         subcontrol.ReferenceFramework,
		ReferenceFrameworkRevision: subcontrol.ReferenceFrameworkRevision,
		OwnerID:                    &orgID,
		// set to empty string to avoid a second query, we know the control owner ID is not set
		ControlOwnerID: lo.ToPtr(""),
	}
}

// HasRevisionChanged checks if the revision of the control has changed compared to the standard revision
func HasRevisionChanged(existingRevision *string, standardRevision string) bool {
	if existingRevision == nil {
		return standardRevision != ""
	}

	return *existingRevision != standardRevision
}

// CreateRevisionUpdateInput creates an UpdateControlInput from the given control that is being updated for a revision change
func CreateRevisionUpdateInput(c *generated.Control) generated.UpdateControlInput {
	input := generated.UpdateControlInput{
		Title:                  &c.Title,
		Aliases:                c.Aliases,
		Description:            &c.Description,
		ClearDescriptionJSON:   true,
		Category:               &c.Category,
		CategoryID:             &c.CategoryID,
		Subcategory:            &c.Subcategory,
		MappedCategories:       c.MappedCategories,
		AssessmentObjectives:   c.AssessmentObjectives,
		AssessmentMethods:      c.AssessmentMethods,
		ControlQuestions:       c.ControlQuestions,
		ImplementationGuidance: c.ImplementationGuidance,
		ExampleEvidence:        c.ExampleEvidence,
		References:             c.References,
		TestingProcedures:      c.TestingProcedures,
		EvidenceRequests:       c.EvidenceRequests,
	}

	if c.Edges.Standard != nil {
		input.ReferenceFrameworkRevision = &c.Edges.Standard.Revision
	}

	return input
}

// CreateSubcontrolRevisionUpdateInput creates an UpdateSubcontrolInput from the given subcontrol that is being updated for a revision change
func CreateSubcontrolRevisionUpdateInput(sc *generated.Subcontrol, standardRevision *string) generated.UpdateSubcontrolInput {
	input := generated.UpdateSubcontrolInput{
		Title:                      &sc.Title,
		Aliases:                    sc.Aliases,
		Description:                &sc.Description,
		ClearDescriptionJSON:       true,
		Category:                   &sc.Category,
		CategoryID:                 &sc.CategoryID,
		Subcategory:                &sc.Subcategory,
		MappedCategories:           sc.MappedCategories,
		AssessmentObjectives:       sc.AssessmentObjectives,
		AssessmentMethods:          sc.AssessmentMethods,
		ControlQuestions:           sc.ControlQuestions,
		ImplementationGuidance:     sc.ImplementationGuidance,
		ExampleEvidence:            sc.ExampleEvidence,
		References:                 sc.References,
		TestingProcedures:          sc.TestingProcedures,
		EvidenceRequests:           sc.EvidenceRequests,
		ReferenceFrameworkRevision: standardRevision,
	}

	return input
}

// GetFieldsToUpdate returns the fields to update for the given control or subcontrol input
// by converting the input to a map and checking for non-empty values
// and then converting back to the appropriate type
// it will return a boolean indicating if there are any fields to update
func GetFieldsToUpdate[T generated.UpdateControlInput | generated.UpdateSubcontrolInput](c *model.CloneControlUploadInput) (*T, bool, error) {
	hasUpdate := false
	updates := map[string]any{}

	input, err := common.ConvertToObject[map[string]any](c.ControlInput)
	if err != nil {
		return nil, false, err
	}

	if input == nil {
		return nil, false, nil
	}

	for k, v := range *input {
		if !common.IsEmpty(v) {
			hasUpdate = true
			updates[k] = v
		}
	}

	out, err := common.ConvertToObject[T](updates)
	if err != nil {
		return nil, false, err
	}

	return out, hasUpdate, nil
}
