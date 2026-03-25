package graphapi

import (
	"context"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

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
