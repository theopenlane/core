package gala

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/theopenlane/core/pkg/logx"
)

// logFieldsCodec captures and restores durable log fields, ensuring the zerolog
// Logger on the restored context is rebuilt with the captured fields. This
// replaces the generic KeyCodec for log_fields because KeyCodec only stores the
// map in the contextx.Key store without enriching the zerolog Logger
type logFieldsCodec struct{}

// Key returns the stable snapshot key
func (logFieldsCodec) Key() ContextKey { return "log_fields" }

// Capture extracts durable log fields from context and JSON-encodes them
func (logFieldsCodec) Capture(ctx context.Context) (json.RawMessage, bool, error) {
	fields := logx.FieldsFromContext(ctx)
	if len(fields) == 0 {
		return nil, false, nil
	}

	encoded, err := json.Marshal(maps.Clone(fields))
	if err != nil {
		return nil, false, ErrContextSnapshotCaptureFailed
	}

	return append(json.RawMessage(nil), encoded...), true, nil
}

// Restore decodes log fields and attaches them to both the zerolog Logger and the durable field store
func (logFieldsCodec) Restore(ctx context.Context, raw json.RawMessage) (context.Context, error) {
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	if len(fields) == 0 {
		return ctx, nil
	}

	return logx.WithFields(ctx, fields), nil
}
