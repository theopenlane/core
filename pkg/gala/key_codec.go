package gala

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/utils/contextx"
)

// KeyCodec connects a contextx.Key to a stable JSON snapshot identifier, making
// any JSON-serializable context value durable across gala event hops.
//
// Declare keys at package scope and register a KeyCodec at application
// startup — gala holds no hardcoded knowledge of domain types.
//
// Example:
//
//	// in iam/auth
//	var CallerKey = contextx.NewKey[*Caller]()
//
//	// at application startup
//	mgr.Register(gala.NewKeyCodec("caller", auth.CallerKey))
type KeyCodec[T any] struct {
	id  ContextKey
	key contextx.Key[T]
}

// NewKeyCodec creates a KeyCodec that captures and restores values from key
// using id as the stable JSON snapshot identifier.
//
// id must be non-empty and unique across all registered codecs. T must be
// JSON-serializable; types containing channels, functions, or unsafe pointers
// will fail at capture time.
func NewKeyCodec[T any](id ContextKey, key contextx.Key[T]) KeyCodec[T] {
	return KeyCodec[T]{id: id, key: key}
}

// Key returns the stable snapshot identifier used for this codec in JSON.
func (c KeyCodec[T]) Key() ContextKey {
	return c.id
}

// Capture extracts the key value from ctx and JSON-encodes it.
// Returns nil, false, nil when the key is not populated.
func (c KeyCodec[T]) Capture(ctx context.Context) (json.RawMessage, bool, error) {
	v, ok := c.key.Get(ctx)
	if !ok {
		return nil, false, nil
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, false, ErrContextSnapshotCaptureFailed
	}

	return append(json.RawMessage(nil), encoded...), true, nil
}

// Restore JSON-decodes raw and stores the result in ctx via the key.
func (c KeyCodec[T]) Restore(ctx context.Context, raw json.RawMessage) (context.Context, error) {
	var v T

	if err := jsonx.RoundTrip(raw, &v); err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	return c.key.Set(ctx, v), nil
}
