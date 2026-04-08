package graphapi

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/controls"
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
func (r *mutationResolver) cloneControlsFromStandard(ctx context.Context, filters controls.CloneFilterOptions, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// first check if the standard exists
	stdWhereFilter := controls.StandardFilter(filters)
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
	if len(stds) > 1 && (filters.StandardShortName != nil && filters.StandardVersion != nil) {
		logger.Error().Err(err).Msgf("error getting standard with ID")

		return nil, fmt.Errorf("%w: error getting standard, too many results", common.ErrInvalidInput)
	}

	// if we get the standard back, all controls should be accessible so we can allow context to skip checks
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	where, err := controls.ControlFilterByStandard(ctx, filters, std)
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

// cloneControls clones the given controls into the organization in the context
// and optionally links them to a program if programID is given
// if the controls already exist in the organization, they will not be cloned again
// but will be updated to link to the program if needed
func (r *mutationResolver) cloneControls(ctx context.Context, controlsToClone []*generated.Control, programID *string) ([]*generated.Control, error) {
	logger := logx.FromContext(ctx)
	// keep track of the control IDs that already exist in the org to be updated to link to the program if needed
	existingControlIDs := []string{}

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
	controlsToUpdate := []controls.ControlToUpdate{}

	for _, c := range controlsToClone {
		// check if the control already exists in the organization
		exists := false

		for _, existingControl := range existingControls {
			if existingControl.RefCode == c.RefCode && existingControl.StandardID == c.StandardID {
				existingControlIDs = append(existingControlIDs, existingControl.ID)
				exists = true

				// we need to check if the revision of the standard has changed since this control was originally cloned
				if c.Edges.Standard != nil && controls.HasRevisionChanged(existingControl.ReferenceFrameworkRevision, c.Edges.Standard.Revision) {
					controlsToUpdate = append(controlsToUpdate, controls.ControlToUpdate{
						ExistingControlID: existingControl.ID,
						SourceControl:     c,
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

	// add options for cloning controls
	opts := []controls.CloneOption{
		controls.WithOrgID(orgID),
	}

	if programID != nil {
		opts = append(opts, controls.WithProgramID(*programID))
	}

	createdControlIDs, subcontrolsToCreate, err := controls.CloneControls(ctx, r.db, updatedControlsToClone, opts...)
	if err != nil {
		logger.Error().Err(err).Msg("error cloning controls")

		return nil, err
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
			Exec(allowCtx); err != nil {
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

// cloneSubcontrols clones the subcontrols from the given control to the new control ID
// it will only clone subcontrols that do not already exist in the organization
func (r *mutationResolver) cloneSubcontrols(ctx context.Context, subcontrolsToCreate []controls.SubcontrolToCreate) error {
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
					toCloneSubcontrol.ReferenceFrameworkRevision = &c.RefControl.Edges.Standard.Revision
				}

				subcontrolsToClone = append(subcontrolsToClone, toCloneSubcontrol)
			}
		}
	}

	subcontrols := make([]*generated.CreateSubcontrolInput, len(subcontrolsToClone))

	for j, subcontrol := range subcontrolsToClone {
		subcontrols[j] = controls.CreateCloneSubcontrolInput(subcontrol, orgID)
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

// updateControlsOnRevisionChange updates existing org controls whose source standard
// revision has changed, refreshing framework-defined fields and handling new subcontrols
func (r *mutationResolver) updateControlsOnRevisionChange(ctx context.Context, controlsToUpdate []controls.ControlToUpdate) error {
	logger := logx.FromContext(ctx)

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	txClient := withTransactionalMutation(ctx)

	for _, cu := range controlsToUpdate {
		updateInput := controls.CreateRevisionUpdateInput(cu.SourceControl)

		if err := txClient.Control.UpdateOneID(cu.ExistingControlID).
			SetInput(updateInput).
			Exec(ctx); err != nil {
			logger.Error().Err(err).Str("control_id", cu.ExistingControlID).Msg("error updating control on revision change")
			return err
		}

		if len(cu.SourceControl.Edges.Subcontrols) == 0 {
			continue
		}

		existingSubcontrols, err := r.db.Subcontrol.Query().
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

			scInput := controls.CreateSubcontrolRevisionUpdateInput(sc, standardRevision)
			if err := txClient.Subcontrol.UpdateOneID(existingID).
				SetInput(scInput).
				Exec(ctx); err != nil {
				logger.Error().Err(err).Str("subcontrol_id", existingID).Msg("error updating subcontrol on revision change")
				return err
			}
		}

		if err := r.cloneSubcontrols(ctx, []controls.SubcontrolToCreate{
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
