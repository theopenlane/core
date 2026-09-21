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
)

// HookAudienceValidateFilters validates audience filters before writes.
func HookAudienceValidateFilters() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AudienceFunc(func(ctx context.Context, m *generated.AudienceMutation) (generated.Value, error) {

			if entx.CheckIsSoftDeleteType(ctx, m.Type()) {
				return next.Mutate(ctx, m)
			}

			typ, ok, err := getAudienceType(ctx, m)
			if err != nil {
				return nil, err
			}
			if !ok {
				return next.Mutate(ctx, m)
			}

			if filters, ok := m.Filters(); ok && typ == enums.AudienceTypeDynamic {
				if err := audiences.ValidateFilters(typ, filters); err != nil {
					return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

func getAudienceType(ctx context.Context, m *generated.AudienceMutation) (enums.AudienceType, bool, error) {
	if typ, ok := m.AudienceType(); ok {
		return typ, true, nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		typ, err := m.OldAudienceType(ctx)
		if err != nil {
			return "", false, err
		}

		return typ, true, nil
	}

	if m.Op().Is(ent.OpUpdate) {
		if _, ok := m.Filters(); ok || m.FiltersCleared() {
			return "", false, fmt.Errorf("%w: %w", ErrInvalidInput, errAudienceFilterMissingBulkType)
		}

		return "", false, nil
	}

	return enums.AudienceTypeManual, true, nil
}
