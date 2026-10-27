package graphapi

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/schema"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
)

type createProgramRequest interface {
	model.CreateFullProgramInput | model.CreateProgramWithMembersInput
}

func hasStandardFilter[T createProgramRequest](value T) bool {
	switch input := any(value).(type) {
	case model.CreateProgramWithMembersInput:
		return input.StandardID != nil || input.StandardShortName != nil
	case model.CreateFullProgramInput:
		return input.StandardID != nil
	}

	return false
}

// controlCloner copies controls into the organization on the request context with the resolver's client and worker pool
func (r *mutationResolver) controlCloner() controls.Cloner {
	return controls.Cloner{DB: r.db, Pool: r.withPool(), RequiredModules: schema.Control{}.Modules()}
}

// getCloneFilterOptions returns the filter options for cloning controls from the input based on the provided fields
func getCloneFilterOptions(input *model.CloneControlInput) controls.CloneFilterOptions {
	filters := controls.CloneFilterOptions{
		Categories: input.Categories,
		RefCodes:   input.RefCodes,
	}

	// if standardID is provided, used that, otherwise use the standardShortName
	if input.StandardID != nil {
		filters.StandardID = input.StandardID
	} else {
		if input.StandardShortName != nil {
			filters.StandardShortName = input.StandardShortName
		}
		if input.StandardVersion != nil {
			filters.StandardVersion = input.StandardVersion
		}
	}

	return filters
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
			if !trimAndCompare(i.StandardVersion, ci.StandardVersion) {
				return nil, fmt.Errorf("%w: all controls for a standard must have the same version", common.ErrInvalidInput)
			}

			if !trimAndCompare(i.OwnerID, ci.OwnerID) {
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

// trimAndCompare trims leading and trailing spaces from two strings and compares them
// returns true if they are equal, false otherwise
func trimAndCompare(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}

	return strings.TrimSpace(*a) == strings.TrimSpace(*b)
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

func (r *mutationResolver) markSubcontrolsAsNotApplicable(ctx context.Context, input []*model.CloneControlUploadInput, controls []*generated.Control) error {
	// find any subcontrols that were created but aren't in the list, and mark as NOT_APPLICABLE
	for _, c := range controls {
		for _, sc := range c.Edges.Subcontrols {
			found := false

			for _, in := range input {
				if in.RefCode != nil && (*in.RefCode == sc.RefCode || slices.Contains(sc.Aliases, *in.RefCode)) {
					found = true
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
