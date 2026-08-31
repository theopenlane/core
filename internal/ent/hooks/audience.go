package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
)

var (
	errAudienceFilterMissingBulkType = errors.New("bulk audience filter updates must include audience_type")
	errAudienceFilterUnsupportedOp   = errors.New("bulk dynamic audience updates must include filters")
	errAudienceFilterManualBulk      = errors.New("bulk manual audience updates must clear filters")
)

// HookAudienceValidateFilters validates audience filters before writes.
func HookAudienceValidateFilters() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AudienceFunc(func(ctx context.Context, m *generated.AudienceMutation) (generated.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, m.Type()) {
				return next.Mutate(ctx, m)
			}

			audienceTyp, err := getAudienceType(ctx, m)
			if err != nil {
				return nil, err
			}

			filters, err := getAudienceFilters(ctx, m)
			if err != nil {
				return nil, err
			}

			if err := audiences.ValidateAudienceFilters(audienceTyp, filters); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func getAudienceType(ctx context.Context, m *generated.AudienceMutation) (enums.AudienceType, error) {
	if audienceType, ok := m.AudienceType(); ok {
		return audienceType, nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		return m.OldAudienceType(ctx)
	}

	if m.Op().Is(ent.OpUpdate) {
		if _, ok := m.Filters(); ok || m.FiltersCleared() {
			return "", fmt.Errorf("%w: %w", ErrInvalidInput, errAudienceFilterMissingBulkType)
		}
	}

	return enums.AudienceTypeManual, nil
}

func getAudienceFilters(ctx context.Context, m *generated.AudienceMutation) (map[string]any, error) {
	if m.FiltersCleared() {
		return map[string]any{}, nil
	}

	if f, ok := m.Filters(); ok {
		return f, nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		return m.OldFilters(ctx)
	}

	audienceType, ok := m.AudienceType()
	if m.Op().Is(ent.OpUpdate) && ok && audienceType == enums.AudienceTypeDynamic {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, errAudienceFilterUnsupportedOp)
	}

	if m.Op().Is(ent.OpUpdate) && ok && audienceType == enums.AudienceTypeManual {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, errAudienceFilterManualBulk)
	}

	return map[string]any{}, nil
}
