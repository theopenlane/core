package templatekit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// BuildDispatchPayload overlays the supplied struct values onto template defaults as a JSON object
// and returns the raw payload consumed by operation dispatchers. Each overlay is marshaled
// through its JSON tags, so the overlay struct types remain the single source of truth for
// per-invocation field names; overlays apply in order, so later overlays win on key conflicts
func BuildDispatchPayload(defaults map[string]any, overlays ...any) (json.RawMessage, error) {
	base, err := jsonx.ToRawMessage(defaults)
	if err != nil {
		return nil, fmt.Errorf("%w: defaults: %w", ErrTemplateRenderFailed, err)
	}

	if len(base) == 0 {
		base = json.RawMessage(`{}`)
	}

	for _, overlay := range overlays {
		patch, err := jsonx.ToRawMap(overlay)
		if err != nil {
			return nil, fmt.Errorf("%w: overlay: %w", ErrTemplateRenderFailed, err)
		}

		base, _, err = jsonx.MergeObjectMap(base, patch)
		if err != nil {
			return nil, fmt.Errorf("%w: merge: %w", ErrTemplateRenderFailed, err)
		}
	}

	return base, nil
}

// ResolveOperationTemplate loads a notification template when referenced and merges its
// defaults into cfg, which must be a pointer to the operation config struct. When neither
// templateID nor templateKey is set the call is a no-op
func ResolveOperationTemplate(ctx context.Context, req types.OperationRequest, templateID, templateKey string, cfg any) error {
	if templateID == "" && templateKey == "" {
		return nil
	}

	ownerID := operationOwnerID(ctx, req)
	if req.DB == nil || ownerID == "" {
		return ErrTemplateNotFound
	}

	template, err := LoadNotificationTemplate(ctx, req.DB, ownerID, templateID, templateKey)
	if err != nil {
		return err
	}

	merged, err := BuildDispatchPayload(template.Defaults, cfg)
	if err != nil {
		return err
	}

	return jsonx.RoundTrip(merged, cfg)
}

func operationOwnerID(ctx context.Context, req types.OperationRequest) string {
	if req.Integration != nil && req.Integration.OwnerID != "" {
		return req.Integration.OwnerID
	}

	if meta, ok := types.ExecutionMetadataFromContext(ctx); ok {
		return meta.OwnerID
	}

	return ""
}
