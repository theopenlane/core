package controls

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
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

// orgID returns the organization the controls are copied into
func (c Cloner) orgID(ctx context.Context) (string, error) {
	if c.OrgID != "" {
		return c.OrgID, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return "", rout.NewMissingRequiredFieldError("owner_id")
	}

	return orgID, nil
}

// FromStandard clones all controls from a standard into an organization
// if the controls already exist in the organization, they will not be cloned again
func (c Cloner) FromStandard(ctx context.Context, filters CloneFilterOptions, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// first check if the standard exists
	stdWhereFilter := StandardFilter(filters)
	stds, err := utils.EntClientFromContext(ctx).Standard.Query().
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
	if len(stds) > 1 && (filters.StandardShortName != nil && filters.StandardVersion != nil) {
		logger.Error().Err(err).Msgf("error getting standard with ID")

		return nil, fmt.Errorf("%w: error getting standard, too many results", common.ErrInvalidInput)
	}

	// if we get the standard back, all controls should be accessible so we can run as an internal read to skip checks
	internalCtx := auth.WithInternalReadContext(ctx)
	where, err := ControlFilterByStandard(ctx, filters, std)
	if err != nil {
		logger.Error().Err(err).Msg("error getting control filter")

		return nil, err
	}

	sourceControls, err := c.DB.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		WithSubcontrols().
		All(internalCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error getting controls to clone")
		return nil, err
	}

	return c.Controls(ctx, sourceControls, programID)
}

// Controls clones the given controls into the organization in the context
// and optionally links them to a program if programID is given
// if the controls already exist in the organization, they will not be cloned again
// but will be updated to link to the program if needed
func (c Cloner) Controls(ctx context.Context, controlsToClone []*generated.Control, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// keep track of the control IDs that already exist in the org to be updated to link to the program if needed
	existingControlIDs := []string{}

	orgID, err := c.orgID(ctx)
	if err != nil {
		return nil, err
	}

	if c.DB.EntConfig != nil && c.DB.EntConfig.Modules.Enabled && len(c.RequiredModules) > 0 {
		// check if the organization has the required modules for Control entities before the parallel execution
		// this prevents multiple queries to the database for each control
		hasModules, _, err := rule.HasAnyFeature(ctx, c.RequiredModules...)
		if err != nil {
			return nil, err
		}

		if !hasModules {
			logger.Error().Msg("organization does not have required modules enabled for control operations")

			return nil, generated.ErrPermissionDenied
		}
	}

	wherePredicate := []predicate.Control{}
	for _, source := range controlsToClone {
		cloneInput, _ := CreateCloneControlInput(source, nil, orgID)
		where := []predicate.Control{control.RefCode(source.RefCode)}
		if standardID := cloneInput.StandardID; standardID != nil {
			where = append(where, control.StandardID(*standardID))
		} else {
			where = append(where, control.StandardIDIsNil())
		}

		wherePredicate = append(wherePredicate, control.And(where...))
	}

	// get existing controls that match the refCode and standardID
	// skip the access checks for the controls, we are already filtering on organization id
	// and controls are visible to users in the organization
	internalCtx := auth.WithInternalOperationContext(ctx)
	existingWhere := []predicate.Control{
		control.DeletedAtIsNil(),
		control.OwnerID(orgID),
	}
	if len(wherePredicate) > 0 {
		existingWhere = append(existingWhere, control.Or(wherePredicate...))
	}

	existingControls, err := c.DB.Control.Query().
		Where(
			control.And(existingWhere...),
		).
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID, control.FieldReferenceFrameworkRevision).
		All(internalCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error checking for existing controls")
		return nil, err
	}

	// find the ones we do need to clone
	updatedControlsToClone := []*generated.Control{}
	controlsToUpdate := []ControlToUpdate{}

	for _, source := range controlsToClone {
		cloneInput, _ := CreateCloneControlInput(source, nil, orgID)
		targetStandardID := cloneInput.StandardID

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

	// check program access if a program is specified
	if programID != nil {
		exists, err := c.DB.Program.Query().Where(program.ID(*programID)).Exist(internalCtx)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, generated.ErrPermissionDenied
		}

		if !c.SkipProgramAccess {
			if err := checkProgramAccess(ctx, c.DB, *programID); err != nil {
				return nil, err
			}
		}
	}

	// add options for cloning controls
	opts := []CloneOption{
		WithOrgID(orgID),
	}

	if programID != nil {
		opts = append(opts, WithProgramID(*programID))
	}

	if c.Pool == nil {
		opts = append(opts, WithoutPool())
	}

	createdControlIDs, subcontrolsToCreate, err := CloneControls(ctx, c.DB, updatedControlsToClone, opts...)
	if err != nil {
		logger.Error().Err(err).Msg("error cloning controls")

		return nil, err
	}

	// allow the subcontrols to be cloned without a parent check, this is because the same user already created the control
	// and the subcontrols are part of the control
	if err := c.cloneSubcontrols(internalCtx, orgID, subcontrolsToCreate); err != nil {
		logger.Error().Err(err).Msg("error cloning subcontrols, rolling back controls that were created before the error occurred")

		if _, err := utils.EntClientFromContext(ctx).Control.Delete().
			Where(control.IDIn(createdControlIDs...)).
			Exec(internalCtx); err != nil {
			logger.Error().Err(err).Msg("error deleting controls that were created before the error occurred")

			return nil, err
		}

		return nil, err
	}

	if len(controlsToUpdate) > 0 {
		if err := c.updateOnRevisionChange(internalCtx, orgID, controlsToUpdate); err != nil {
			logger.Error().Err(err).Msg("error updating controls on revision change")
			return nil, err
		}
	}

	// update the existing controls to the program if needed
	if len(existingControlIDs) > 0 && programID != nil && *programID != "" {
		// if the control already exists, we just link it to the program
		if err := utils.EntClientFromContext(ctx).Control.Update().
			Where(
				control.IDIn(existingControlIDs...)).
			AddProgramIDs(*programID).
			Exec(internalCtx); err != nil {
			return nil, err
		}
	}

	// add existingControlIDs to createdControlIDs
	createdControlIDs = append(createdControlIDs, existingControlIDs...)

	// get the cloned controls to return in the response
	query, err := utils.EntClientFromContext(ctx).Control.Query().Where(control.IDIn(createdControlIDs...)).
		WithSubcontrols().
		CollectFields(internalCtx)
	if err != nil {
		logger.Error().Err(err).Msg("error collecting fields for cloned controls")

		return nil, err
	}

	return query.All(internalCtx)
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

	// check if we can find the subcontrol based on refCode and controlID
	logger := logx.FromContext(ctx)

	existingSubcontrols, err := c.DB.Subcontrol.Query().
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

	subcontrolsToClone := []*generated.CreateSubcontrolInput{}

	// get the subcontrols we actually need to clone
	for _, toCreate := range subcontrolsToCreate {
		for _, toCloneSubcontrol := range toCreate.RefControl.Edges.Subcontrols {
			// check if the subcontrol already exists in the organization
			exists := false

			for _, existingSubcontrol := range existingSubcontrols {
				if existingSubcontrol.RefCode == toCloneSubcontrol.RefCode &&
					existingSubcontrol.ControlID == toCreate.NewControlID {
					exists = true
					break
				}
			}

			if !exists {
				// add the subcontrol to the list of subcontrols to clone
				toCloneSubcontrol.ControlID = toCreate.NewControlID

				subcontrolsToClone = append(subcontrolsToClone, CreateCloneSubcontrolInput(toCloneSubcontrol, orgID, toCreate))
			}
		}
	}

	return c.bulkCreateSubcontrolNoTransaction(ctx, subcontrolsToClone)
}

// bulkCreateSubcontrolNoTransaction creates multiple subcontrols in a single request without a transaction to allow it to be run in parallel
func (c Cloner) bulkCreateSubcontrolNoTransaction(ctx context.Context, input []*generated.CreateSubcontrolInput) error {
	errors := []error{}
	logger := logx.FromContext(ctx)

	var mu sync.Mutex

	funks := make([]func(), len(input))

	for i, data := range input {
		sub := data // capture loop variable
		funks[i] = func() {
			if err := c.DB.Subcontrol.Create().
				SetInput(*sub).Exec(ctx); err != nil {
				mu.Lock()

				errors = append(errors, err)

				mu.Unlock()

				return
			}
		}
	}

	// run the cloning functions in parallel, or one at a time without a pool
	if err := c.runTasks(funks); err != nil {
		return err
	}

	if len(errors) == 0 {
		return nil
	}

	logger.Error().Errs("errors", errors).Msg("errors cloning subcontrols")

	// return the first error but log all
	return errors[0]
}

// runTasks runs the tasks on the pool, or in order when the cloner has none
func (c Cloner) runTasks(tasks []func()) error {
	if c.Pool != nil {
		return c.Pool.SubmitMultipleAndWait(tasks)
	}

	for _, task := range tasks {
		task()
	}

	return nil
}

// updateOnRevisionChange updates existing org controls whose source standard
// revision has changed, refreshing framework-defined fields and handling new subcontrols
func (c Cloner) updateOnRevisionChange(ctx context.Context, orgID string, controlsToUpdate []ControlToUpdate) error {
	logger := logx.FromContext(ctx)

	txClient := utils.EntClientFromContext(ctx)

	for _, cu := range controlsToUpdate {
		updateInput := CreateRevisionUpdateInput(cu.SourceControl)

		if err := txClient.Control.UpdateOneID(cu.ExistingControlID).
			SetInput(updateInput).
			Exec(ctx); err != nil {
			logger.Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error updating control on revision change")
			return err
		}

		if len(cu.SourceControl.Edges.Subcontrols) == 0 {
			continue
		}

		existingSubcontrols, err := c.DB.Subcontrol.Query().
			Where(
				subcontrol.DeletedAtIsNil(),
				subcontrol.OwnerID(orgID),
				subcontrol.ControlID(cu.ExistingControlID),
			).
			Select(subcontrol.FieldID, subcontrol.FieldRefCode, subcontrol.FieldControlID).
			All(ctx)
		if err != nil {
			logger.Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error querying existing subcontrols for revision update")
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

			scInput := CreateSubcontrolRevisionUpdateInput(sc, standardRevision)
			if err := txClient.Subcontrol.UpdateOneID(existingID).
				SetInput(scInput).
				Exec(ctx); err != nil {
				logger.Error().Err(err).Str("subcontrol_id", existingID).Msg("error updating subcontrol on revision change")
				return err
			}
		}

		if err := c.cloneSubcontrols(ctx, orgID, []SubcontrolToCreate{
			{
				NewControlID: cu.ExistingControlID,
				RefControl:   cu.SourceControl,
			},
		}); err != nil {
			logger.Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error creating new subcontrols on revision change")
			return err
		}
	}

	return nil
}
