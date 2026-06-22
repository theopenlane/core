package operations

import (
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

const (
	executionMetadataCodecKey = gala.ContextKey("integration_execution")
	directorySyncRunCodecKey  = gala.ContextKey("integration_directory_sync_run_id")
	activeTrustCenterCodecKey = gala.ContextKey("active_trust_center_id")
)

// ContextCodecs returns the durable context codecs required by integration dispatch and ingest listeners
func ContextCodecs() []gala.ContextCodec {
	return []gala.ContextCodec{
		gala.NewKeyCodec(executionMetadataCodecKey, types.ExecutionMetadataKey),
		gala.NewKeyCodec(directorySyncRunCodecKey, directorySyncRunIDKey),
		// used to preserve the trust center key context from anon trust center requests, needed for email send to follow same query restrictions
		gala.NewKeyCodec(activeTrustCenterCodecKey, auth.ActiveTrustCenterIDKey),
	}
}
