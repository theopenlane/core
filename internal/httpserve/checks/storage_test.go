package checks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/httpserve/checks"
	"github.com/theopenlane/shared/objects/storage"
)

func TestStorageAvailabilityCheck(t *testing.T) {
	check := checks.StorageAvailabilityCheck(func() storage.ProviderConfig {
		return storage.ProviderConfig{Enabled: false}
	})

	require.NoError(t, check(context.Background()))
}
