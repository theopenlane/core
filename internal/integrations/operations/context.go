package operations

import (
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	executionMetadataCodecKey = gala.ContextKey("integration_execution")
	directorySyncRunCodecKey  = gala.ContextKey("integration_directory_sync_run_id")
)

// ContextCodecs returns the durable context codecs required by integration dispatch and ingest listeners
func ContextCodecs() []gala.ContextCodec {
	return []gala.ContextCodec{
		gala.NewKeyCodec(executionMetadataCodecKey, types.ExecutionMetadataKey),
		gala.NewKeyCodec(directorySyncRunCodecKey, directorySyncRunIDKey),
	}
}
