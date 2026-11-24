package objectstore

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/disk"
)

func TestReadObjectFromDisk(t *testing.T) {
	provider, err := disk.NewDiskProvider(&storage.ProviderOptions{Bucket: t.TempDir()})
	assert.NilError(t, err)

	missing, err := ReadObject{Key: "vulnintel/osv/CVE-2025-65637.json"}.Run(context.Background(), provider)
	assert.NilError(t, err)
	assert.Check(t, !missing.Found)

	_, err = WriteObject{Key: "vulnintel/osv/CVE-2025-65637.json", Content: map[string]any{"found": true}}.Run(context.Background(), provider)
	assert.NilError(t, err)

	read, err := ReadObject{Key: "vulnintel/osv/CVE-2025-65637.json"}.Run(context.Background(), provider)
	assert.NilError(t, err)
	assert.Check(t, read.Found)
	assert.Check(t, is.Equal(`{"found":true}`, string(read.Content)))
}
