package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestProvider_Copy_MissingEntClient(t *testing.T) {
	provider := &Provider{options: storage.NewProviderOptions(storage.WithBucket("test-bucket"))}

	err := provider.Copy(context.Background(), &storagetypes.File{ID: "01HZZTESTFILEDB000000"}, &storage.CopyOptions{Key: "01HZZTESTFILEDB000001"})
	require.ErrorIs(t, err, ErrMissingEntClient)
}
