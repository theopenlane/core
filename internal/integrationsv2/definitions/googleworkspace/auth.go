package googleworkspace

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/internal/integrationsv2/types"
)

// startInstallAuth starts the Google Workspace OAuth install flow
func startInstallAuth(context.Context, json.RawMessage) (types.AuthStartResult, error) {
	return types.AuthStartResult{}, errors.New("googleworkspace: oauth install flow not yet implemented")
}

// completeInstallAuth completes the Google Workspace OAuth install flow
func completeInstallAuth(context.Context, json.RawMessage, json.RawMessage) (types.AuthCompleteResult, error) {
	return types.AuthCompleteResult{}, errors.New("googleworkspace: oauth install flow not yet implemented")
}
