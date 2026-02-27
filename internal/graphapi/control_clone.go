package graphapi

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/schema"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
)

type createProgramRequest interface {
	model.CreateFullProgramInput | model.CreateProgramWithMembersInput
}

func hasStandardFilter[T createProgramRequest](value T) bool {
	switch input := any(value).(type) {
	case model.CreateProgramWithMembersInput:
		if input.StandardID != nil {
			return true
		} else if input.StandardShortName != nil {
			return true
		}
	case model.CreateFullProgramInput:
		if input.StandardID != nil {
			return true
		}
	}

	return false
}

// cloneControlsFromStandard clones all controls from a standard into an organization
// if the controls already exist in the organization, they will not be cloned again
func (r *mutationResolver) cloneControlsFromStandard(ctx context.Context, filters cloneFilterOptions, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// first check if the standard exists
	stdWhereFilter := standardFilter(filters)
	stds, err := withTransactionalMutation(ctx).Standard.Query().
		Where(stdWhereFilter...).
		Select(standard.FieldID, standard.FieldIsPublic).
		Order(standard.OrderOption(standard.ByVersion(sql.OrderDesc()))).
		All(ctx)
	if err != nil || stds == nil || len(stds) == 0 {
		logger.Error().Err(err).Msgf("error getting standard with ID")

		return nil, err
	}

	// get the first standard, this will be the most recent revision if multiple revisions exist
	std := stds[0]

	// if we have more than one standard, and the version was provided, return an error
	// because we are unable to determine which standard to use
	if len(stds) > 1 && (filters.standardShortName != nil && filters.standardVersion != nil) {
		logger.Error().Err(err).Msgf("error getting standard with ID")

		return nil, fmt.Errorf("%w: error getting standard, too many results", common.ErrInvalidInput)
	}

	// if we get the standard back, all controls should be accessible so we can allow context to skip checks
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	where, err := controlFilterByStandard(ctx, filters, std)
	if err != nil {
		logger.Error().Err(err).Msg("error getting control filter")

		return nil, err
	}

	controls, err := r.db.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		WithSubcontrols().
		All(allowCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error getting controls to clone")
		return nil, err
	}

	return r.cloneControls(ctx, controls, programID)
}

// subcontrolToCreate is used to track which subcontrols need to be created for a given control
type subcontrolToCreate struct {
	newControlID string
	refControl   *generated.Control
}

// controlToUpdate is used to track existing controls that need to be updated due to changes
// in the revision of their connected standards
type controlToUpdate struct {
	existingControlID string
	sourceControl     *generated.Control
}

// cloneControls clones the given controls into the organization in the context
// and optionally links them to a program if programID is given
// if the controls already exist in the organization, they will not be cloned again
// but will be updated to link to the program if needed
func (r *mutationResolver) cloneControls(ctx context.Context, controlsToClone []*generated.Control, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// keep track of the created control IDs, this includes the ids of controls that already exist in the org
	createdControlIDs := []string{}
	// keep track of the control IDs that already exist in the org to be updated to link to the program if needed
	existingControlIDs := []string{}
	// track subcontrols to create
	subcontrolsToCreate := []subcontrolToCreate{}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	orgID := caller.OrganizationID

	if r.db.EntConfig != nil && r.db.EntConfig.Modules.Enabled {
		// check if the organization has the required modules for Control entities before the parallel execution
		// this prevents multiple queries to the database for each control
		hasModules, _, err := rule.HasAnyFeature(ctx, schema.Control{}.Modules()...)
		if err != nil {
			return nil, err
		}

		if !hasModules {
			logger.Error().Str("organization_id", caller.OrganizationID).Msg("organization does not have required modules enabled for control operations")

			return nil, generated.ErrPermissionDenied
		}
	}

	// do this in a go-routine to allow multiple controls to be cloned in parallel, use the worker pool for this
	// we cannot use a transaction here because we are running multiple go-routines
	// and transactions cannot be used across go-routines
	var (
		errors []error
		mu     sync.Mutex
	)

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
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID, control.FieldReferenceFrameworkRevision).
		All(allowCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error checking for existing controls")
		return nil, err
	}

	// find the ones we do need to clone
	updatedControlsToClone := []*generated.Control{}
	controlsToUpdate := []controlToUpdate{}

	for _, c := range controlsToClone {
		// check if the control already exists in the organization
		exists := false

		for _, existingControl := range existingControls {
			if existingControl.RefCode == c.RefCode && existingControl.StandardID == c.StandardID {
				existingControlIDs = append(existingControlIDs, existingControl.ID)
				exists = true

				// we need to check if the revision of the standard has changed since this control was originally cloned
				if c.Edges.Standard != nil && hasRevisionChanged(existingControl.ReferenceFrameworkRevision, c.Edges.Standard.Revision) {
					controlsToUpdate = append(controlsToUpdate, controlToUpdate{
						existingControlID: existingControl.ID,
						sourceControl:     c,
					})
				}
			}
		}

		if !exists {
			// control does not exist, we will clone it
			updatedControlsToClone = append(updatedControlsToClone, c)
		}
	}

	// check program access if a program is specified
	if programID != nil {
		allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
			ObjectType:  generated.TypeProgram,
			ObjectID:    *programID,
			Relation:    fgax.CanEdit,
			SubjectID:   caller.SubjectID,
			SubjectType: caller.SubjectType(),
		})
		if err != nil {
			return nil, err
		}

		if !allow {
			logger.Error().Str("organization_id", caller.OrganizationID).Str("user_id", caller.SubjectID).Msg("no access to edit specified program")

			return nil, generated.ErrPermissionDenied
		}
	}

	ctrlCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// create a function for each control to clone
	// this will allow us to run the cloning in parallel
	// we will use a mutex to protect the createdControlIDs and existingControlIDs slices
	// and the errors slice
	funcs := make([]func(), len(updatedControlsToClone))

	for i, c := range updatedControlsToClone {
		c := c // capture loop variable
		funcs[i] = func() {
			controlInput, _, isTCControl := createCloneControlInput(c, programID, orgID)

			res, err := r.db.Control.Create().
				SetInput(controlInput).
				SetIsTrustCenterControl(isTCControl).Save(ctrlCtx)
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
					newControlID: newControlID,
					refControl:   c,
				})
			}

			mu.Unlock()
		}
	}

	// run the cloning functions in parallel
	if err := r.withPool().SubmitMultipleAndWait(funcs); err != nil {
		return nil, err
	}

	// check if there were any errors during the cloning process
	if len(errors) > 0 {
		// return the first error encountered
		logger.Error().Errs("errors", errors).
			Msgf("error cloning controls, deleting %d controls that were created before the error occurred", len(createdControlIDs))

		// delete any controls that were created before the error occurred
		if len(createdControlIDs) > 0 {
			// delete any controls that were created before the error occurred
			// this should also cascade delete any subcontrols that were created
			if _, err := withTransactionalMutation(ctx).Control.Delete().
				Where(control.IDIn(createdControlIDs...)).
				Exec(allowCtx); err != nil {
				logger.Error().Err(err).Msg("error deleting controls that were created before the error occurred")
			}
		}

		// we can return the first error encountered, as the rest will be logged
		return nil, errors[0]
	}

	// allow the subcontrols to be cloned without a parent check, this is because the same user already created the control
	// and the subcontrols are part of the control
	if err := r.cloneSubcontrols(allowCtx, subcontrolsToCreate); err != nil {
		logger.Error().Err(err).Msg("error cloning subcontrols, rolling back controls that were created before the error occurred")

		if _, err := withTransactionalMutation(ctx).Control.Delete().
			Where(control.IDIn(createdControlIDs...)).
			Exec(allowCtx); err != nil {
			logger.Error().Err(err).Msg("error deleting controls that were created before the error occurred")

			return nil, err
		}

		return nil, err
	}

	if len(controlsToUpdate) > 0 {
		if err := r.updateControlsOnRevisionChange(allowCtx, controlsToUpdate); err != nil {
			logger.Error().Err(err).Msg("error updating controls on revision change")
			return nil, err
		}
	}

	// update the existing controls to the program if needed
	if len(existingControlIDs) > 0 && programID != nil && *programID != "" {
		// if the control already exists, we just link it to the program
		if err := withTransactionalMutation(ctx).Control.Update().
			Where(
				control.IDIn(existingControlIDs...)).
			AddProgramIDs(*programID).
			Exec(ctrlCtx); err != nil {
			return nil, err
		}
	}

	// add existingControlIDs to createdControlIDs
	createdControlIDs = append(createdControlIDs, existingControlIDs...)

	// get the cloned controls to return in the response
	query, err := withTransactionalMutation(ctx).Control.Query().Where(control.IDIn(createdControlIDs...)).
		WithSubcontrols().
		CollectFields(allowCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error collecting fields for cloned controls")

		return nil, err
	}

	return query.All(allowCtx)
}

// trustCenterStandardShortName is the short name of the trust center standard
// used to identify controls that should be flagged as trust center controls during clone
const trustCenterStandardShortName = "openlane-trust-center"

// isTrustCenterStandard returns true if the standard is the trust center standard
func isTrustCenterStandard(std *generated.Standard) bool {
	return std != nil && std.ShortName == trustCenterStandardShortName
}

// createCloneControlInput creates a CreateControlInput from the given control that is being cloned
// and returns the input, the standard ID that was set, and whether the control is a trust center control
func createCloneControlInput(c *generated.Control, programID *string, orgID string) (generated.CreateControlInput, string, bool) {
	controlInput := generated.CreateControlInput{
		// grab fields from the existing control
		Tags:                   c.Tags,
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

	return controlInput, standardID, isTrustCenterStandard(c.Edges.Standard)
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

	wherePredicate := []predicate.Subcontrol{}

	for _, c := range subcontrolsToCreate {
		// get all the refCodes for the subcontrols in the control
		refCodes := []string{}
		for _, s := range c.refControl.Edges.Subcontrols {
			refCodes = append(refCodes, s.RefCode)
		}

		where := subcontrol.Or(
			subcontrol.RefCodeIn(refCodes...),
			subcontrol.ControlID(c.newControlID),
		)

		wherePredicate = append(wherePredicate, where)
	}

	// check if we can find the subcontrol based on refCode and controlID
	logger := logx.FromContext(ctx)

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
		logger.Error().Err(err).Msg("error checking for existing subcontrols")

		return err
	}

	subcontrolsToClone := []*generated.Subcontrol{}

	// get the subcontrols we actually need to clone
	for _, c := range subcontrolsToCreate {
		for _, toCloneSubcontrol := range c.refControl.Edges.Subcontrols {
			// check if the subcontrol already exists in the organization
			exists := false

			for _, existingSubcontrol := range existingSubcontrols {
				if existingSubcontrol.RefCode == toCloneSubcontrol.RefCode &&
					existingSubcontrol.ControlID == c.newControlID {
					exists = true
					break
				}
			}

			if !exists {
				// add the subcontrol to the list of subcontrols to clone
				toCloneSubcontrol.ControlID = c.newControlID

				if c.refControl.Edges.Standard != nil {
					toCloneSubcontrol.ReferenceFramework = &c.refControl.Edges.Standard.ShortName
					toCloneSubcontrol.ReferenceFrameworkRevision = &c.refControl.Edges.Standard.Revision
				}

				subcontrolsToClone = append(subcontrolsToClone, toCloneSubcontrol)
			}
		}
	}

	subcontrols := make([]*generated.CreateSubcontrolInput, len(subcontrolsToClone))

	for j, subcontrol := range subcontrolsToClone {
		subcontrols[j] = &generated.CreateSubcontrolInput{
			Tags:                       subcontrol.Tags,
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

	return r.bulkCreateSubcontrolNoTransaction(ctx, subcontrols)
}

// bulkCreateSubcontrolNoTransaction creates multiple subcontrols in a single request without a transaction to allow it to be run in parallel
func (r *mutationResolver) bulkCreateSubcontrolNoTransaction(ctx context.Context, input []*generated.CreateSubcontrolInput) error {
	errors := []error{}
	logger := logx.FromContext(ctx)

	var mu sync.Mutex

	funks := make([]func(), len(input))

	for i, data := range input {
		c := data // capture loop variable
		funks[i] = func() {
			if err := r.db.Subcontrol.Create().
				SetInput(*c).Exec(ctx); err != nil {
				mu.Lock()

				errors = append(errors, err)

				mu.Unlock()

				return
			}
		}
	}

	// run the cloning functions in parallel
	if err := r.withPool().SubmitMultipleAndWait(funks); err != nil {
		return err
	}

	if len(errors) == 0 {
		return nil
	}

	logger.Error().Errs("errors", errors).Msg("errors cloning subcontrols")

	// return the first error but log all
	return errors[0]
}

func hasRevisionChanged(existingRevision *string, standardRevision string) bool {
	if existingRevision == nil {
		return standardRevision != ""
	}

	return *existingRevision != standardRevision
}

func createRevisionUpdateInput(c *generated.Control) generated.UpdateControlInput {
	input := generated.UpdateControlInput{
		Title:                  &c.Title,
		Aliases:                c.Aliases,
		Description:            &c.Description,
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

func createSubcontrolRevisionUpdateInput(sc *generated.Subcontrol, standardRevision *string) generated.UpdateSubcontrolInput {
	input := generated.UpdateSubcontrolInput{
		Title:                      &sc.Title,
		Aliases:                    sc.Aliases,
		Description:                &sc.Description,
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

// updateControlsOnRevisionChange updates existing org controls whose source standard
// revision has changed, refreshing framework-defined fields and handling new subcontrols
func (r *mutationResolver) updateControlsOnRevisionChange(ctx context.Context, controls []controlToUpdate) error {
	logger := logx.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	txClient := withTransactionalMutation(ctx)

	for _, cu := range controls {
		updateInput := createRevisionUpdateInput(cu.sourceControl)

		if err := txClient.Control.UpdateOneID(cu.existingControlID).
			SetInput(updateInput).
			Exec(ctx); err != nil {
			logger.Error().Err(err).Str("control_id", cu.existingControlID).Msg("error updating control on revision change")
			return err
		}

		if len(cu.sourceControl.Edges.Subcontrols) == 0 {
			continue
		}

		existingSubcontrols, err := r.db.Subcontrol.Query().
			Where(
				subcontrol.DeletedAtIsNil(),
				subcontrol.OwnerID(orgID),
				subcontrol.ControlID(cu.existingControlID),
			).
			Select(subcontrol.FieldID, subcontrol.FieldRefCode, subcontrol.FieldControlID).
			All(ctx)
		if err != nil {
			logger.Error().Err(err).Str("control_id", cu.existingControlID).Msg("error querying existing subcontrols for revision update")
			return err
		}

		existingByRefCode := make(map[string]string, len(existingSubcontrols))
		for _, es := range existingSubcontrols {
			existingByRefCode[es.RefCode] = es.ID
		}

		var standardRevision *string
		if cu.sourceControl.Edges.Standard != nil {
			standardRevision = &cu.sourceControl.Edges.Standard.Revision
		}

		for _, sc := range cu.sourceControl.Edges.Subcontrols {
			existingID, ok := existingByRefCode[sc.RefCode]
			if !ok {
				continue
			}

			scInput := createSubcontrolRevisionUpdateInput(sc, standardRevision)
			if err := txClient.Subcontrol.UpdateOneID(existingID).
				SetInput(scInput).
				Exec(ctx); err != nil {
				logger.Error().Err(err).Str("subcontrol_id", existingID).Msg("error updating subcontrol on revision change")
				return err
			}
		}

		if err := r.cloneSubcontrols(ctx, []subcontrolToCreate{
			{
				newControlID: cu.existingControlID,
				refControl:   cu.sourceControl,
			},
		}); err != nil {
			logger.Error().Err(err).Str("control_id", cu.existingControlID).Msg("error creating new subcontrols on revision change")
			return err
		}
	}

	return nil
}

func (r *mutationResolver) markSubcontrolsAsNotApplicable(ctx context.Context, input []*model.CloneControlUploadInput, controls []*generated.Control) error {
	// find any subcontrols that were created but aren't in the list, and mark as NOT_APPLICABLE
	for _, c := range controls {
		for _, sc := range c.Edges.Subcontrols {
			found := false
			refCode := sc.RefCode
			aliases := sc.Aliases

			for _, c := range input {
				if c.RefCode == &refCode {
					found = true
					break
				}

				for _, alias := range aliases {
					if c.RefCode == &alias {
						found = true
						break
					}
				}

				if found {
					break
				}
			}

			if !found {
				if err := r.db.Subcontrol.UpdateOneID(sc.ID).SetStatus(enums.ControlStatusNotApplicable).Exec(ctx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// getFieldsToUpdate returns the fields to update for the given control or subcontrol input
// by converting the input to a map and checking for non-empty values
// and then converting back to the appropriate type
// it will return a boolean indicating if there are any fields to update
func getFieldsToUpdate[T generated.UpdateControlInput | generated.UpdateSubcontrolInput](c *model.CloneControlUploadInput) (*T, bool, error) {
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
