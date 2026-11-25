package controls

import (
	"context"
	"sync"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/logx"
)

// Cloner copies controls into an organization, skipping controls the organization already has, linking them
// to the program instead, and refreshing those whose standard revision changed
type Cloner struct {
	// DB reads existing rows and creates the copies
	DB *generated.Client
	// Pool creates the copies in parallel, nil creates them one at a time for callers inside a transaction
	Pool *gala.Pool
	// OrgID is the organization the controls are copied into, empty uses the organization on the context
	OrgID string
	// RequiredModules are the modules the organization needs for control operations, nil skips the check
	RequiredModules []models.OrgModule
	// SkipProgramAccess skips the caller's edit check on the program, for callers that created the program themselves
	SkipProgramAccess bool
}

// FromStandard clones all controls from a standard into an organization
// if the controls already exist in the organization, they will not be cloned again
func (c Cloner) FromStandard(ctx context.Context, filters CloneFilterOptions, programID *string) ([]*generated.Control, error) {
	sourceControls, err := c.SourceControlsFromStandard(ctx, filters)
	if err != nil {
		return nil, err
	}

	return c.Controls(ctx, sourceControls, programID)
}

// SourceControlsFromStandard loads the controls and subcontrols from the accessible standard selected by the filters.
func (c Cloner) SourceControlsFromStandard(ctx context.Context, filters CloneFilterOptions) ([]*generated.Control, error) {
	// first check if the standard exists
	std, err := findStandard(ctx, c.DB, filters)
	if err != nil {
		return nil, err
	}

	// if we get the standard back, all controls should be accessible so we can run as an internal read to skip checks
	readCtx := auth.WithInternalReadContext(ctx)
	where, err := ControlFilterByStandard(ctx, filters, std)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting control filter")

		return nil, err
	}

	sourceControls, err := c.DB.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		WithSubcontrols().
		All(readCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting controls to clone")
		return nil, err
	}

	return sourceControls, nil
}

// Controls clones the given controls into the organization in the context
// and optionally links them to a program if programID is given
// if the controls already exist in the organization, they will not be cloned again
// but will be updated to link to the program if needed
func (c Cloner) Controls(ctx context.Context, controlsToClone []*generated.Control, programID *string) ([]*generated.Control, error) {
	orgID, err := utils.ResolveOwnerID(ctx, c.OrgID)
	if err != nil {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	// check if the organization has the required modules for Control entities before the parallel execution
	// this prevents multiple queries to the database for each control
	if err := c.checkRequiredModules(ctx); err != nil {
		return nil, err
	}

	// check program access if a program is specified
	if programID != nil {
		if err := c.checkProgram(ctx, *programID); err != nil {
			return nil, err
		}
	}

	// get existing controls that match the refCode and standardID
	// skip the access checks for the controls, we are already filtering on organization id
	// and controls are visible to users in the organization
	readCtx := auth.WithInternalReadContext(ctx)

	existingControls, err := c.DB.Control.Query().
		Where(
			existingControlsFilter(controlsToClone, orgID),
		).
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID, control.FieldReferenceFrameworkRevision).
		All(readCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking for existing controls")
		return nil, err
	}

	// find the ones we do need to clone
	updatedControlsToClone, controlsToUpdate, existingControlIDs := splitExistingControls(controlsToClone, existingControls)

	// add options for cloning controls
	opts := c.cloneOptions(orgID, programID)

	createdControlIDs, subcontrolsToCreate, err := CloneControls(ctx, c.DB, updatedControlsToClone, opts...)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error cloning controls")

		return nil, err
	}

	internalCtx := auth.WithInternalOperationContext(ctx)

	// allow the subcontrols to be cloned without a parent check, this is because the same user already created the control
	// and the subcontrols are part of the control
	if err := c.cloneSubcontrols(internalCtx, orgID, subcontrolsToCreate); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error cloning subcontrols, rolling back controls that were created before the error occurred")

		if err := c.rollbackControls(internalCtx, createdControlIDs); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error deleting controls that were created before the error occurred")

			return nil, err
		}

		return nil, err
	}

	if len(controlsToUpdate) > 0 {
		if err := c.updateOnRevisionChange(internalCtx, orgID, controlsToUpdate); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error updating controls on revision change")
			return nil, err
		}
	}

	// update the existing controls to the program if needed
	if err := c.linkControlsToProgram(internalCtx, existingControlIDs, programID); err != nil {
		return nil, err
	}

	// add existingControlIDs to createdControlIDs
	createdControlIDs = append(createdControlIDs, existingControlIDs...)

	// get the cloned controls to return in the response
	query, err := c.DB.Control.Query().Where(control.IDIn(createdControlIDs...)).
		WithSubcontrols().
		CollectFields(readCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error collecting fields for cloned controls")

		return nil, err
	}

	return query.All(readCtx)
}

// checkRequiredModules returns a permission error when the organization lacks the cloner's required modules
func (c Cloner) checkRequiredModules(ctx context.Context) error {
	if !utils.ModulesEnabled(c.DB) || len(c.RequiredModules) == 0 {
		return nil
	}

	hasModules, _, err := rule.HasAnyFeature(ctx, c.RequiredModules...)
	if err != nil {
		return err
	}

	if !hasModules {
		logx.FromContext(ctx).Error().Msg("organization does not have required modules enabled for control operations")

		return generated.ErrPermissionDenied
	}

	return nil
}

// existingControlsFilter returns the predicate to find controls in the organization matching the refCode and standardID of the controls to clone
// the owner filter is required because the org interceptor also returns the system owned source controls
func existingControlsFilter(controlsToClone []*generated.Control, orgID string) predicate.Control {
	existingWhere := []predicate.Control{
		control.DeletedAtIsNil(),
		control.OwnerID(orgID),
	}

	wherePredicate := []predicate.Control{}
	for _, source := range controlsToClone {
		where := []predicate.Control{control.RefCode(source.RefCode)}
		if standardID := cloneStandardID(source); standardID != nil {
			where = append(where, control.StandardID(*standardID))
		} else {
			where = append(where, control.StandardIDIsNil())
		}

		wherePredicate = append(wherePredicate, control.And(where...))
	}

	if len(wherePredicate) > 0 {
		existingWhere = append(existingWhere, control.Or(wherePredicate...))
	}

	return control.And(existingWhere...)
}

// splitExistingControls separates the controls that need to be cloned from those that already exist in the organization,
// returning the controls to clone, the existing controls whose standard revision changed, and the IDs of all existing controls
func splitExistingControls(controlsToClone, existingControls []*generated.Control) ([]*generated.Control, []ControlToUpdate, []string) {
	// keep track of the control IDs that already exist in the org to be updated to link to the program if needed
	existingControlIDs := []string{}
	updatedControlsToClone := []*generated.Control{}
	controlsToUpdate := []ControlToUpdate{}

	for _, source := range controlsToClone {
		targetStandardID := cloneStandardID(source)

		// check if the control already exists in the organization
		exists := false

		for _, existingControl := range existingControls {
			targetStandardMatches := targetStandardID == nil && existingControl.StandardID == ""
			if targetStandardID != nil {
				targetStandardMatches = *targetStandardID == existingControl.StandardID
			}

			if existingControl.RefCode == source.RefCode && targetStandardMatches {
				existingControlIDs = append(existingControlIDs, existingControl.ID)
				exists = true

				// we need to check if the revision of the standard has changed since this control was originally cloned
				if targetStandardID != nil && source.Edges.Standard != nil && HasRevisionChanged(existingControl.ReferenceFrameworkRevision, source.Edges.Standard.Revision) {
					controlsToUpdate = append(controlsToUpdate, ControlToUpdate{
						ExistingControlID: existingControl.ID,
						SourceControl:     source,
					})
				}
			}
		}

		if !exists {
			// control does not exist, we will clone it
			updatedControlsToClone = append(updatedControlsToClone, source)
		}
	}

	return updatedControlsToClone, controlsToUpdate, existingControlIDs
}

// cloneOptions returns the options for cloning controls into the organization and optional program
func (c Cloner) cloneOptions(orgID string, programID *string) []CloneOption {
	opts := []CloneOption{
		WithOrgID(orgID),
		WithPool(c.Pool),
	}

	if programID != nil {
		opts = append(opts, WithProgramID(*programID))
	}

	return opts
}

// linkControlsToProgram links the existing controls to the program, when a program is given
func (c Cloner) linkControlsToProgram(ctx context.Context, controlIDs []string, programID *string) error {
	if len(controlIDs) == 0 || programID == nil || *programID == "" {
		return nil
	}

	// if the control already exists, we just link it to the program
	return c.DB.Control.Update().
		Where(control.IDIn(controlIDs...)).
		AddProgramIDs(*programID).
		Exec(ctx)
}

// rollbackControls deletes controls created outside the caller's transaction
// without a pool the controls were created in the transaction and its rollback removes them
func (c Cloner) rollbackControls(ctx context.Context, controlIDs []string) error {
	if c.Pool == nil || len(controlIDs) == 0 {
		return nil
	}

	_, err := c.DB.Control.Delete().Where(control.IDIn(controlIDs...)).Exec(ctx)

	return err
}

// checkProgram checks the program exists and, unless skipped, that the caller can edit it
func (c Cloner) checkProgram(ctx context.Context, programID string) error {
	exists, err := c.DB.Program.Query().Where(program.ID(programID)).Exist(auth.WithInternalReadContext(ctx))
	if err != nil {
		return err
	}

	if !exists {
		return generated.ErrPermissionDenied
	}

	if c.SkipProgramAccess {
		return nil
	}

	return checkProgramAccess(ctx, c.DB, programID)
}

// checkProgramAccess checks the users access to the specific program id
func checkProgramAccess(ctx context.Context, db *generated.Client, programID string) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok {
		return auth.ErrNoAuthUser
	}

	// support users should skip this check but for other users,
	// we still want to verify they have edit access to the program
	if caller.Has(auth.CapOrgSupport) {
		return nil
	}

	allow, err := db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  generated.TypeProgram,
		ObjectID:    programID,
		Relation:    fgax.CanEdit,
		SubjectID:   caller.SubjectID,
		SubjectType: caller.SubjectType(),
	})
	if err != nil {
		return err
	}

	if !allow {
		logx.FromContext(ctx).Error().Msg("no access to edit specified program")

		return generated.ErrPermissionDenied
	}

	return nil
}

// cloneSubcontrols clones the subcontrols from the given control to the new control ID
// it will only clone subcontrols that do not already exist in the organization
func (c Cloner) cloneSubcontrols(ctx context.Context, orgID string, subcontrolsToCreate []SubcontrolToCreate) error {
	if len(subcontrolsToCreate) == 0 {
		return nil
	}

	// check if we can find the subcontrol based on refCode and controlID
	existingSubcontrols, err := c.DB.Subcontrol.Query().
		Where(existingSubcontrolsFilter(subcontrolsToCreate)).
		Select(subcontrol.FieldRefCode, subcontrol.FieldControlID).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking for existing subcontrols")

		return err
	}

	return c.createSubcontrols(ctx, splitExistingSubcontrols(subcontrolsToCreate, existingSubcontrols, orgID))
}

// existingSubcontrolsFilter returns the predicate to find subcontrols in the organization matching the refCodes or new control of the subcontrols to create
func existingSubcontrolsFilter(subcontrolsToCreate []SubcontrolToCreate) predicate.Subcontrol {
	wherePredicate := []predicate.Subcontrol{}

	for _, toCreate := range subcontrolsToCreate {
		// get all the refCodes for the subcontrols in the control
		refCodes := []string{}
		for _, s := range toCreate.RefControl.Edges.Subcontrols {
			refCodes = append(refCodes, s.RefCode)
		}

		where := subcontrol.Or(
			subcontrol.RefCodeIn(refCodes...),
			subcontrol.ControlID(toCreate.NewControlID),
		)

		wherePredicate = append(wherePredicate, where)
	}

	return subcontrol.And(
		subcontrol.DeletedAtIsNil(),
		subcontrol.Or(wherePredicate...),
	)
}

// splitExistingSubcontrols returns the create inputs for the subcontrols that do not already exist on their new control
func splitExistingSubcontrols(subcontrolsToCreate []SubcontrolToCreate, existingSubcontrols []*generated.Subcontrol, orgID string) []*generated.CreateSubcontrolInput {
	existing := make(map[[2]string]bool, len(existingSubcontrols))
	for _, es := range existingSubcontrols {
		existing[[2]string{es.ControlID, es.RefCode}] = true
	}

	subcontrolsToClone := []*generated.CreateSubcontrolInput{}

	// get the subcontrols we actually need to clone
	for _, toCreate := range subcontrolsToCreate {
		for _, toCloneSubcontrol := range toCreate.RefControl.Edges.Subcontrols {
			// check if the subcontrol already exists in the organization
			if existing[[2]string{toCreate.NewControlID, toCloneSubcontrol.RefCode}] {
				continue
			}

			// add the subcontrol to the list of subcontrols to clone
			toCloneSubcontrol.ControlID = toCreate.NewControlID

			subcontrolsToClone = append(subcontrolsToClone, CreateCloneSubcontrolInput(toCloneSubcontrol, orgID, toCreate))
		}
	}

	return subcontrolsToClone
}

// createSubcontrols creates the subcontrols in parallel on the pool, or one at a time inside the caller's transaction without one
func (c Cloner) createSubcontrols(ctx context.Context, input []*generated.CreateSubcontrolInput) error {
	errs := []error{}

	var mu sync.Mutex

	tasks := make([]func(), len(input))

	for i, sub := range input {
		tasks[i] = func() {
			if err := c.DB.Subcontrol.Create().
				SetInput(*sub).Exec(ctx); err != nil {
				mu.Lock()

				errs = append(errs, err)

				mu.Unlock()

				return
			}
		}
	}

	// run the cloning functions in parallel, or one at a time without a pool
	if err := runTasks(c.Pool, tasks); err != nil {
		return err
	}

	if len(errs) == 0 {
		return nil
	}

	logx.FromContext(ctx).Error().Errs("errors", errs).Msg("errors cloning subcontrols")

	// return the first error but log all
	return errs[0]
}

// updateOnRevisionChange updates existing org controls whose source standard
// revision has changed, refreshing framework-defined fields and handling new subcontrols
func (c Cloner) updateOnRevisionChange(ctx context.Context, orgID string, controlsToUpdate []ControlToUpdate) error {
	for _, cu := range controlsToUpdate {
		if err := c.updateControlOnRevisionChange(ctx, orgID, cu); err != nil {
			return err
		}
	}

	return nil
}

// updateControlOnRevisionChange refreshes one existing control and its subcontrols from the new standard revision
func (c Cloner) updateControlOnRevisionChange(ctx context.Context, orgID string, cu ControlToUpdate) error {
	if err := c.DB.Control.UpdateOneID(cu.ExistingControlID).
		SetInput(CreateRevisionUpdateInput(cu.SourceControl)).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error updating control on revision change")
		return err
	}

	if len(cu.SourceControl.Edges.Subcontrols) == 0 {
		return nil
	}

	if err := c.updateSubcontrolsOnRevisionChange(ctx, cu); err != nil {
		return err
	}

	// create any subcontrols added in the new revision
	if err := c.cloneSubcontrols(ctx, orgID, []SubcontrolToCreate{
		{
			NewControlID: cu.ExistingControlID,
			RefControl:   cu.SourceControl,
		},
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error creating new subcontrols on revision change")
		return err
	}

	return nil
}

// updateSubcontrolsOnRevisionChange refreshes the existing subcontrols of the control that are still in the new standard revision
func (c Cloner) updateSubcontrolsOnRevisionChange(ctx context.Context, cu ControlToUpdate) error {
	existingSubcontrols, err := c.DB.Subcontrol.Query().
		Where(
			subcontrol.DeletedAtIsNil(),
			subcontrol.ControlID(cu.ExistingControlID),
		).
		Select(subcontrol.FieldID, subcontrol.FieldRefCode).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error querying existing subcontrols for revision update")
		return err
	}

	existingByRefCode := make(map[string]string, len(existingSubcontrols))
	for _, es := range existingSubcontrols {
		existingByRefCode[es.RefCode] = es.ID
	}

	var standardRevision *string
	if cu.SourceControl.Edges.Standard != nil {
		standardRevision = &cu.SourceControl.Edges.Standard.Revision
	}

	for _, sc := range cu.SourceControl.Edges.Subcontrols {
		existingID, ok := existingByRefCode[sc.RefCode]
		if !ok {
			continue
		}

		if err := c.DB.Subcontrol.UpdateOneID(existingID).
			SetInput(CreateSubcontrolRevisionUpdateInput(sc, standardRevision)).
			Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("subcontrol_id", existingID).Msg("error updating subcontrol on revision change")
			return err
		}
	}

	return nil
}
