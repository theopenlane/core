package checks

import (
	"context"
	"errors"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/shared/objects/storage"
	"github.com/theopenlane/shared/objects/validators"
)

// StorageAvailabilityCheck returns a handlers.CheckFunc that validates storage provider availability
func StorageAvailabilityCheck(cfgProvider func() storage.ProviderConfig) handlers.CheckFunc {
	return func(ctx context.Context) error {
		errs := validators.ValidateAvailabilityByProvider(ctx, cfgProvider())
		if len(errs) == 0 {
			return nil
		}

		return errors.Join(errs...)
	}
}
