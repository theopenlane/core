package gala

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/samber/do/v2"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/contextx"
)

// ContextKey identifies a restorable context value key
// using string alias for better readability and to avoid collisions with other context keys
// this has to be a string to be used as a JSON key for durability rather than a strict type + contextx
type ContextKey = string

// ContextSnapshot captures context data that can be restored after durable hops
type ContextSnapshot struct {
	// Values contains codec-managed context values
	Values map[ContextKey]json.RawMessage `json:"values,omitempty"`
}

// ContextCodec captures and restores one durable context value; construct via
// NewKeyCodec or a codec builder, never as a bare literal
type ContextCodec struct {
	// id is the stable snapshot identifier
	id ContextKey
	// capture extracts and encodes context data
	capture func(context.Context) (json.RawMessage, bool, error)
	// restore decodes and re-attaches context data
	restore func(context.Context, json.RawMessage) (context.Context, error)
}

// ContextManager manages context codecs and snapshot round-trips; codecs are registered
// during wiring only, so snapshot round-trips read the codec map without locking
type ContextManager struct {
	codecs map[ContextKey]ContextCodec
}

// newContextManager creates a context manager and registers any initial codecs
func newContextManager(codecs ...ContextCodec) (*ContextManager, error) {
	manager := &ContextManager{codecs: map[ContextKey]ContextCodec{}}

	for _, codec := range codecs {
		if err := manager.Register(codec); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

// Register registers a context codec by key
func (m *ContextManager) Register(codec ContextCodec) error {
	if codec.id == "" {
		return ErrContextCodecKeyRequired
	}

	if _, exists := m.codecs[codec.id]; exists {
		return ErrContextCodecAlreadyRegistered
	}

	m.codecs[codec.id] = codec

	return nil
}

// Capture captures all registered context codec values
func (m *ContextManager) Capture(ctx context.Context) (ContextSnapshot, error) {
	values := map[ContextKey]json.RawMessage{}

	for _, key := range slices.Sorted(maps.Keys(m.codecs)) {
		raw, present, err := m.codecs[key].capture(ctx)
		if err != nil {
			return ContextSnapshot{Values: values}, fmt.Errorf("%w: %s", err, key)
		}

		if present {
			values[key] = append(json.RawMessage(nil), raw...)
		}
	}

	var snapshot ContextSnapshot

	if len(values) > 0 {
		snapshot.Values = values
	}

	return snapshot, nil
}

// Restore restores snapshot values into a new context
func (m *ContextManager) Restore(ctx context.Context, snapshot ContextSnapshot) (context.Context, error) {
	restored := ctx

	for _, key := range slices.Sorted(maps.Keys(snapshot.Values)) {
		codec, exists := m.codecs[key]
		if !exists {
			continue
		}

		next, err := codec.restore(restored, snapshot.Values[key])
		if err != nil {
			return restored, fmt.Errorf("%w: %s", err, key)
		}

		restored = next
	}

	return restored, nil
}

// newInjectorCodec creates a codec that restores a dependency from the gala injector
// onto the handler context without serializing the value itself: capture stores a
// sentinel marker so restore is invoked on the handler side, where the value is
// resolved from the injector and attached via setter
func newInjectorCodec[T any](id ContextKey, injector do.Injector, setter func(context.Context, T) context.Context) ContextCodec {
	return ContextCodec{
		id: id,
		capture: func(context.Context) (json.RawMessage, bool, error) {
			return json.RawMessage(`true`), true, nil
		},
		restore: func(ctx context.Context, _ json.RawMessage) (context.Context, error) {
			val, err := do.Invoke[T](injector)
			if err != nil {
				return ctx, ErrContextSnapshotRestoreFailed
			}

			return setter(ctx, val), nil
		},
	}
}

// NewKeyCodec creates a codec that captures and restores values from key using id as
// the stable JSON snapshot identifier
func NewKeyCodec[T any](id ContextKey, key contextx.Key[T]) ContextCodec {
	return newAccessorCodec(id, key.Get, key.Set)
}

// newAccessorCodec builds a codec from a typed get/set accessor pair
func newAccessorCodec[T any](id ContextKey, get func(context.Context) (T, bool), set func(context.Context, T) context.Context) ContextCodec {
	return ContextCodec{
		id: id,
		capture: func(ctx context.Context) (json.RawMessage, bool, error) {
			v, ok := get(ctx)
			if !ok {
				return nil, false, nil
			}

			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, false, ErrContextSnapshotCaptureFailed
			}

			return encoded, true, nil
		},
		restore: func(ctx context.Context, raw json.RawMessage) (context.Context, error) {
			var v T

			if err := jsonx.RoundTrip(raw, &v); err != nil {
				return ctx, ErrContextSnapshotRestoreFailed
			}

			return set(ctx, v), nil
		},
	}
}

// logFieldsCodec builds the durable log fields codec; restoring via logx.WithFields rebuilds the zerolog logger, not just the field store
func logFieldsCodec() ContextCodec {
	return newAccessorCodec("log_fields", func(ctx context.Context) (map[string]any, bool) {
		fields := logx.FieldsFromContext(ctx)

		return fields, len(fields) > 0
	}, logx.WithFields)
}

// WorkflowFlags carries the workflow bypass controls across durable dispatch hops
type WorkflowFlags struct {
	// Bypass skips workflow approval interceptors for system operations
	Bypass bool `json:"bypass,omitempty"`
	// AllowEventEmission keeps workflow listener execution enabled while Bypass is set
	AllowEventEmission bool `json:"allow_event_emission,omitempty"`
}

// WorkflowFlagsKey stores the workflow bypass controls in context
var WorkflowFlagsKey = contextx.NewKey[WorkflowFlags]()

// DirectorySyncRunIDKey carries the directory sync run id to downstream ingest handlers
var DirectorySyncRunIDKey = contextx.NewKey[string]()
