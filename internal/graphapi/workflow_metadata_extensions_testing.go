//go:build test

package graphapi

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	intr "github.com/theopenlane/core/internal/integrations/runtime"
)

// WorkflowMetadataExtensions exposes workflowMetadataExtensions for cross-package integration tests
func WorkflowMetadataExtensions(ctx context.Context, rt *intr.Runtime, db *ent.Client) map[string]any {
	return workflowMetadataExtensions(ctx, rt, db)
}
