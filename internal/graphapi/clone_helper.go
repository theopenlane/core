package graphapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/model"
)

func convertToCloneControlInput(input []*model.CloneControlUploadInput) ([]*model.CloneControlInput, error) {
	out := []*model.CloneControlInput{}

	// create a map of standards first
	stds := sliceToMap(input)

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
				return nil, fmt.Errorf("%w: all controls for a standard must have the same version", ErrInvalidInput)
			}

			if !stripeAndCompare(i.OwnerID, ci.OwnerID) {
				return nil, fmt.Errorf("%w: all controls for a standard must have the same owner", ErrInvalidInput)
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

func sliceToMap(input []*model.CloneControlUploadInput) map[string][]*model.CloneControlUploadInput {
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

func getControlIDFromRefCode(ctx context.Context, refCode string, controls []*generated.Control) (*string, bool) {
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
		sc, err := c.Subcontrols(ctx, nil, nil, nil, nil, nil, nil)
		if err != nil {
			continue
		}

		for _, s := range sc.Edges {
			if s.Node.RefCode == refCode {
				return &c.ID, true
			}

			for _, alias := range s.Node.Aliases {
				if alias == refCode {
					return &c.ID, true
				}
			}
		}
	}

	return nil, false
}
