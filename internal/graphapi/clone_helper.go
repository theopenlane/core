package graphapi

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

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

// getCloneFilterOptions returns the filter options for cloning controls from the input based on the provided fields
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

// controlRefCodeFilter returns the predicate to filter by control refCode or alias based on the filter options
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

// controlFilterByStandard returns the predicate to filter controls by standard and other filter options
// to return the correct controls to clone
func controlFilterByStandard(ctx context.Context, opts cloneFilterOptions, std *generated.Standard) ([]predicate.Control, error) {
	where := []predicate.Control{
		control.DeletedAtIsNil(),
		control.StandardID(std.ID),
	}

	if std.IsPublic {
		where = append(where, control.SystemOwned(true))
	} else {
		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil || orgID == "" {
			return nil, rout.NewMissingRequiredFieldError("owner_id")
		}

		where = append(where, control.OwnerID(orgID))
	}

	if len(opts.categories) > 0 {
		where = append(where, control.CategoryIn(opts.categories...))
	}

	refCodeFilter := controlRefCodeFilter(opts)
	if refCodeFilter != nil {
		where = append(where, refCodeFilter...)
	}

	return where, nil
}

// filterByStandard returns true if the filter options contain a standardID or standardShortName
func filterByStandard(opts cloneFilterOptions) bool {
	return opts.standardID != nil || opts.standardShortName != nil
}

// convertToCloneControlInput converts a slice of CloneControlUploadInput to a slice of CloneControlInput
// this is used to process a bulk CSV upload of controls to be cloned and group them by standard
func convertToCloneControlInput(input []*model.CloneControlUploadInput) ([]*model.CloneControlInput, error) {
	out := []*model.CloneControlInput{}

	// create a map of standards first
	stds := controlUploadSliceToMap(input)

	for stdName, controlInputs := range stds {
		// sanity check if there are no controls keep going
		if len(controlInputs) == 0 {
			continue
		}

		i := &model.CloneControlInput{}

		_, err := ulid.Parse(stdName)
		if err == nil {
			i.StandardID = &stdName
		} else {
			i.StandardShortName = &stdName
		}

		if controlInputs[0].StandardVersion != nil {
			stdVersion := strings.TrimSpace(*controlInputs[0].StandardVersion)
			i.StandardVersion = &stdVersion
		}

		if controlInputs[0].OwnerID != nil {
			ownerID := strings.TrimSpace(*controlInputs[0].OwnerID)
			i.OwnerID = &ownerID
		}

		for _, ci := range controlInputs {
			if !stripeAndCompare(i.StandardVersion, ci.StandardVersion) {
				return nil, fmt.Errorf("%w: all controls for a standard must have the same version", common.ErrInvalidInput)
			}

			if !stripeAndCompare(i.OwnerID, ci.OwnerID) {
				return nil, fmt.Errorf("%w: all controls for a standard must have the same owner", common.ErrInvalidInput)
			}

			if ci.RefCode != nil {
				i.RefCodes = append(i.RefCodes, *ci.RefCode)
			}

			if ci.ControlID != nil {
				i.ControlIDs = append(i.ControlIDs, *ci.ControlID)
			}
		}

		out = append(out, i)
	}

	return out, nil
}

// controlUploadSliceToMap converts a slice of CloneControlUploadInput to a map grouped by standard
func controlUploadSliceToMap(input []*model.CloneControlUploadInput) map[string][]*model.CloneControlUploadInput {
	out := map[string][]*model.CloneControlUploadInput{}

	for _, i := range input {
		key := ""
		if i.StandardID != nil {
			key = *i.StandardID
		} else if i.StandardShortName != nil {
			key = *i.StandardShortName
		}

		if key != "" {
			out[key] = append(out[key], i)
		}
	}

	return out
}

// stripeAndCompare trims leading and trailing spaces from two strings and compares them
// returns true if they are equal, false otherwise
func stripeAndCompare(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil && b != nil {
		return false
	}

	if a != nil && b == nil {
		return false
	}

	return strings.TrimSpace(*a) == strings.TrimSpace(*b)
}

// getControlIDFromRefCode searches for a control ID by ref code or alias in the provided controls
// returns the control ID and a boolean indicating if it is a subcontrol or not
func getControlIDFromRefCode(refCode string, controls []*generated.Control) (*string, bool) {
	for _, c := range controls {
		if c.RefCode == refCode {
			return &c.ID, false
		}
	}

	for _, c := range controls {
		for _, alias := range c.Aliases {
			if alias == refCode {
				return &c.ID, false
			}
		}
	}

	for _, c := range controls {
		sc := c.Edges.Subcontrols
		if sc == nil {
			continue
		}

		for _, s := range sc {
			if s.RefCode == refCode {
				return &c.ID, true
			}

			for _, alias := range s.Aliases {
				if alias == refCode {
					return &c.ID, true
				}
			}
		}
	}

	return nil, false
}

// cleanImplementationGuidance cleans and formats the implementation guidance input
// by splitting on new lines and trimming spaces
func cleanImplementationGuidance(input *string) *models.ImplementationGuidance {
	if input == nil || *input == "" {
		return nil
	}

	parts := strings.Split(*input, "\n")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)

	}

	guide := models.ImplementationGuidance{
		Guidance: parts,
	}

	return &guide
}

// createControlImplementation creates a control implementation for the given control ID and owner ID with the provided implementation guidance and makes it verified and published
func (r *mutationResolver) createControlImplementation(ctx context.Context, ownerID *string, controlID string, input *string) error {
	if input == nil || *input == "" {
		return nil
	}

	cleanImp := strings.Trim(*input, "\"")
	cleanImp = strings.TrimSpace(cleanImp)

	if cleanImp == "" {
		return nil
	}

	coInput := generated.CreateControlImplementationInput{
		Status:     &enums.DocumentPublished,
		Verified:   lo.ToPtr(true),
		Details:    &cleanImp,
		OwnerID:    ownerID,
		ControlIDs: []string{controlID},
	}

	return r.db.ControlImplementation.Create().SetInput(coInput).Exec(ctx)
}

// createControlObjective creates a control objective for the given control ID and owner ID with the provided objective details and makes it verified and published
func (r *mutationResolver) createControlObjective(ctx context.Context, ownerID *string, controlID string, input *string) error {
	if input == nil || *input == "" {
		return nil
	}

	// create control implementation
	co := strings.Trim(*input, "\"")
	co = strings.TrimSpace(co)

	if co == "" {
		return nil
	}

	// create control objective
	coInput := generated.CreateControlObjectiveInput{
		DesiredOutcome: &co,
		Status:         &enums.ObjectiveActiveStatus,
		OwnerID:        ownerID,
		ControlIDs:     []string{controlID},
	}

	return r.db.ControlObjective.Create().SetInput(coInput).Exec(ctx)
}

// createComment creates a comment for the given owner ID with the provided comment text,  and returns
// the created comment ID in a slice
func (r *mutationResolver) createComment(ctx context.Context, ownerID *string, input *string) ([]string, error) {
	if input == nil || *input == "" {
		return nil, nil
	}

	cleanComment := strings.Trim(*input, "\"")
	cleanComment = strings.TrimSpace(cleanComment)

	if cleanComment == "" {
		return nil, nil
	}

	commentInput := generated.CreateNoteInput{
		Text:    cleanComment,
		OwnerID: ownerID,
	}

	res, err := r.db.Note.Create().SetInput(commentInput).Save(ctx)
	if err != nil {
		return nil, err
	}

	return []string{res.ID}, nil
}
