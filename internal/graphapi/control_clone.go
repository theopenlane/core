package graphapi

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/schema"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
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

// cloneFilterOptions holds the filter options for cloning controls
type cloneFilterOptions struct {
	// standardID of the standard to filter the controls by
	standardID *string
	// standardShortName is the short name of the standard to filter the controls by
	standardShortName *string
	// standardVersion is the revision of the standard to filter the controls by, this is optional, if more than one revision exists and this is not provided, the most recent revision will be used
	standardVersion *string
	// refCodes is a list of refCodes to filter the controls by
	refCodes []string
	// categories is a list of categories to filter the controls by
	categories []string
}

func getCloneFilterOptions(input *model.CloneControlInput) cloneFilterOptions {
	filters := cloneFilterOptions{
		categories: input.Categories,
		refCodes:   input.RefCodes,
	}

	// if standardID is provided, used that, otherwise use the standardShortName
	if input.StandardID != nil {
		filters.standardID = input.StandardID
	} else if input.StandardShortName != nil {
		filters.standardShortName = input.StandardShortName
	}

	return filters
}

func getUploadFilterOptions(input *model.CloneControlUploadInput) (cloneFilterOptions, bool) {
	hasFilter := false

	filters := cloneFilterOptions{}

	if input.RefCode != nil {
		filters.refCodes = []string{*input.RefCode}
		hasFilter = true
	}

	// if standardID is provided, used that, otherwise use the standardShortName
	if input.StandardID != nil {
		filters.standardID = input.StandardID
		hasFilter = true
	} else if input.StandardShortName != nil {
		filters.standardShortName = input.StandardShortName
		hasFilter = true
	}

	return filters, hasFilter
}

// standardFilter returns the predicate to filter by standard based on the filter options
func standardFilter(opts cloneFilterOptions) []predicate.Standard {
	// safety check to make sure at least the ID or shortName is set
	if !filterByStandard(opts) {
		return nil
	}

	stdWhereFilter := []predicate.Standard{}

	if opts.standardID != nil {
		stdWhereFilter = append(stdWhereFilter, standard.ID(*opts.standardID))
	} else {
		stdWhereFilter = append(stdWhereFilter, standard.ShortNameEqualFold(*opts.standardShortName))
		if opts.standardVersion != nil {
			stdWhereFilter = append(stdWhereFilter, standard.VersionEqualFold(*opts.standardVersion))
		}
	}

	return stdWhereFilter
}

func controlRefCodeFilter(opts cloneFilterOptions) []predicate.Control {
	if len(opts.refCodes) > 0 {
		refOpts := []predicate.Control{control.RefCodeIn(opts.refCodes...)}
		for _, refCode := range opts.refCodes {
			refOpts = append(refOpts, func(s *sql.Selector) {
				s.Where(sqljson.ValueContains(control.FieldAliases, refCode))
			})
		}

		return []predicate.Control{
			control.Or(
				refOpts...,
			),
		}
	}

	return nil
}

// filterByStandard returns true if the filter options contain a standardID or standardShortName
func filterByStandard(opts cloneFilterOptions) bool {
	return opts.standardID != nil || opts.standardShortName != nil
}

// cloneControlsFromStandard clones all controls from a standard into an organization
// if the controls already exist in the organization, they will not be cloned again
func (r *mutationResolver) cloneControlsFromStandard(ctx context.Context, filters cloneFilterOptions, programID *string) ([]*generated.Control, error) {
	// first check if the standard exists
	stdWhereFilter := standardFilter(filters)
	stds, err := withTransactionalMutation(ctx).Standard.Query().
		Where(stdWhereFilter...).
		Select(standard.FieldID, standard.FieldIsPublic).
		Order(standard.OrderOption(standard.ByVersion(sql.OrderDesc()))).
		All(ctx)
	if err != nil || stds == nil || len(stds) == 0 {
		log.Error().Err(err).Msgf("error getting standard with ID %s", filters.standardID)

		return nil, err
	}

	// get the first standard, this will be the most recent revision if multiple revisions exist
	std := stds[0]

	// if we have more than one standard, and the version was provided, return an error
	// because we are unable to determine which standard to use
	if len(stds) > 1 && (filters.standardShortName != nil && filters.standardVersion != nil) {
		log.Error().Err(err).Msgf("error getting standard with ID %s", filters.standardID)
		return nil, fmt.Errorf("%w: error getting standard, too many results", ErrInvalidInput)
	}

	// if we get the standard back, all controls should be accessible so we can allow context to skip checks
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	where := []predicate.Control{
		control.DeletedAtIsNil(),
		control.StandardID(std.ID),
	}

	// if the standard is public, we can get the controls that do not have an owner (public controls)
	// if the standard is not public, we need to get the organization ID from the context
	if std.IsPublic {
		where = append(where, control.OwnerIDIsNil())
	} else {
		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil || orgID == "" {
			return nil, rout.NewMissingRequiredFieldError("owner_id")
		}

		where = append(where, control.OwnerID(orgID))
	}

	if len(filters.categories) > 0 {
		where = append(where, control.CategoryIn(filters.categories...))
	}

	refCodeFilter := controlRefCodeFilter(filters)
	if refCodeFilter != nil {
		where = append(where, refCodeFilter...)
	}

	controls, err := r.db.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		WithSubcontrols().
		All(allowCtx)
	if err != nil {
		log.Error().Err(err).Msg("error getting controls to clone")
		return nil, err
	}

	return r.cloneControls(ctx, controls, programID)
}

// subcontrolToCreate is used to track which subcontrols need to be created for a given control
type subcontrolToCreate struct {
	newControlID string
	refControl   *generated.Control
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

	ac, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || ac.OrganizationID == "" {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	orgID := ac.OrganizationID

	if r.db.EntConfig != nil && r.db.EntConfig.Modules.Enabled {
		// check if the organization has the required modules for Control entities before the parallel execution
		// this prevents multiple queries to the database for each control
		hasModules, _, err := rule.HasAnyFeature(ctx, schema.Control{}.Modules()...)
		if err != nil {
			return nil, err
		}

		if !hasModules {
			return nil, generated.ErrPermissionDenied
		}
	}

	// do this in a go-routine to allow multiple controls to be cloned in parallel, use pond for this
	// we cannot use a transaction here because we are running multiple go-routines
	// and transactions cannot be used across go-routines
	funcs := make([]func(), len(controlsToClone))

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
		Select(control.FieldID, control.FieldRefCode, control.FieldStandardID).
		All(allowCtx)
	if err != nil {
		log.Error().Err(err).Msg("error checking for existing controls")
		return nil, err
	}

	// find the ones we do need to clone
	updatedControlsToClone := []*generated.Control{}

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
			updatedControlsToClone = append(updatedControlsToClone, c)
		}
	}

	// check program access if a program is specified
	if programID != nil {
		allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
			ObjectType:  generated.TypeProgram,
			ObjectID:    *programID,
			Relation:    fgax.CanEdit,
			SubjectID:   ac.SubjectID,
			SubjectType: auth.GetAuthzSubjectType(ctx),
		})
		if err != nil {
			return nil, err
		}

		if !allow {
			return nil, generated.ErrPermissionDenied
		}
	}

	ctrlCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// create a function for each control to clone
	// this will allow us to run the cloning in parallel
	// we will use a mutex to protect the createdControlIDs and existingControlIDs slices
	// and the errors slice
	for i, c := range updatedControlsToClone {
		c := c // capture loop variable
		funcs[i] = func() {
			controlInput, _ := createCloneControlInput(c, programID, orgID)

			res, err := r.db.Control.Create().
				SetInput(controlInput).Save(ctrlCtx)
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
	r.withPool().SubmitMultipleAndWait(funcs)

	// check if there were any errors during the cloning process
	if len(errors) > 0 {
		// return the first error encountered
		log.Error().Errs("errors", errors).
			Msgf("error cloning controls, deleting %d controls that were created before the error occurred", len(createdControlIDs))

		// delete any controls that were created before the error occurred
		if len(createdControlIDs) > 0 {
			// delete any controls that were created before the error occurred
			// this should also cascade delete any subcontrols that were created
			if _, err := withTransactionalMutation(ctx).Control.Delete().
				Where(control.IDIn(createdControlIDs...)).
				Exec(allowCtx); err != nil {
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
			Exec(allowCtx); err != nil {
			log.Error().Err(err).Msg("error deleting controls that were created before the error occurred")

			return nil, err
		}

		return nil, err
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
		CollectFields(allowCtx)
	if err != nil {
		return nil, err
	}

	return query.All(allowCtx)
}

// createCloneControlInput creates a CreateControlInput from the given control that is being cloned
// and returns the input and the standard ID that was set
func createCloneControlInput(c *generated.Control, programID *string, orgID string) (generated.CreateControlInput, string) {
	controlInput := generated.CreateControlInput{
		// grab fields from the existing control
		Tags:                   c.Tags,
		RefCode:                c.RefCode,
		Title:                  &c.Title,
		Aliases:                c.Aliases,
		Description:            &c.Description,
		Source:                 &c.Source,
		ControlType:            &c.ControlType,
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
		// set default status to not implemented
		Status:  &enums.ControlStatusNotImplemented,
		OwnerID: &orgID,
	}

	if c.Edges.Standard != nil {
		// if the control has a standard, we will set the reference framework to the standard
		controlInput.ReferenceFramework = &c.Edges.Standard.ShortName
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
			AssessmentMethods:      subcontrol.AssessmentMethods,
			ControlQuestions:       subcontrol.ControlQuestions,
			ImplementationGuidance: subcontrol.ImplementationGuidance,
			ExampleEvidence:        subcontrol.ExampleEvidence,
			References:             subcontrol.References,
			Status:                 &enums.ControlStatusNotImplemented,
			ReferenceFramework:     subcontrol.ReferenceFramework,
			OwnerID:                &orgID,
			// set to empty string to avoid a second query, we know the control owner ID is not set
			ControlOwnerID: lo.ToPtr(""),
		}
	}

	return r.bulkCreateSubcontrolNoTransaction(ctx, subcontrols)
}

// bulkCreateSubcontrolNoTransaction creates multiple subcontrols in a single request without a transaction to allow it to be run in parallel
func (r *mutationResolver) bulkCreateSubcontrolNoTransaction(ctx context.Context, input []*generated.CreateSubcontrolInput) error {
	errors := []error{}

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
	r.withPool().SubmitMultipleAndWait(funks)

	if len(errors) == 0 {
		return nil
	}

	log.Error().Errs("errors", errors).Msg("errors cloning subcontrols")

	// return the first error but log all
	return errors[0]
}
