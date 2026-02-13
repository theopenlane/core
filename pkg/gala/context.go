package gala

import (
	"context"
	"encoding/json"
	"slices"
	"sync"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/contextx"
)

// ContextCodec captures and restores one typed context value
type ContextCodec interface {
	// Key returns the stable snapshot key
	Key() ContextKey
	// Capture extracts and encodes context data
	Capture(context.Context) (json.RawMessage, bool, error)
	// Restore decodes and re-attaches context data
	Restore(context.Context, json.RawMessage) (context.Context, error)
}

// TypedContextCodec captures/restores context values stored via contextx.With
type TypedContextCodec[T any] struct {
	key ContextKey
}

// ContextManager manages context codecs and snapshot round-trips
type ContextManager struct {
	mu     sync.RWMutex
	codecs map[ContextKey]ContextCodec
}

// contextFlagSet stores boolean flags in context
type contextFlagSet struct {
	Flags map[ContextFlag]bool
}

// NewTypedContextCodec creates a typed context codec for a specific snapshot key
func NewTypedContextCodec[T any](key ContextKey) TypedContextCodec[T] {
	return TypedContextCodec[T]{key: key}
}

// Key returns the codec snapshot key
func (c TypedContextCodec[T]) Key() ContextKey {
	return c.key
}

// Capture extracts a typed context value and JSON encodes it
func (c TypedContextCodec[T]) Capture(ctx context.Context) (json.RawMessage, bool, error) {
	value, ok := contextx.From[T](ctx)
	if !ok {
		return nil, false, nil
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, false, ErrContextSnapshotCaptureFailed
	}

	return append(json.RawMessage(nil), encoded...), true, nil
}

// Restore JSON decodes a typed context value and re-attaches it
func (c TypedContextCodec[T]) Restore(ctx context.Context, raw json.RawMessage) (context.Context, error) {
	var value T

	if err := json.Unmarshal(raw, &value); err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	return contextx.With(ctx, value), nil
}

// NewContextManager creates a context manager and registers any initial codecs
func NewContextManager(codecs ...ContextCodec) (*ContextManager, error) {
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
	if codec == nil {
		return ErrContextCodecRequired
	}

	key := codec.Key()
	if key == "" {
		return ErrContextCodecKeyRequired
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.codecs[key]; exists {
		return ErrContextCodecAlreadyRegistered
	}

	m.codecs[key] = codec

	return nil
}

// Capture captures all registered context codec values and current context flags
func (m *ContextManager) Capture(ctx context.Context) (ContextSnapshot, error) {
	codecs := m.codecsSnapshot()
	codecKeys := lo.Keys(codecs)
	slices.Sort(codecKeys)

	snapshot := ContextSnapshot{
		Values: map[ContextKey]json.RawMessage{},
	}

	for _, key := range codecKeys {
		codec := codecs[key]
		raw, present, err := codec.Capture(ctx)
		if err != nil {
			return ContextSnapshot{}, ErrContextSnapshotCaptureFailed
		}

		if !present {
			continue
		}

		snapshot.Values[key] = append(json.RawMessage(nil), raw...)
	}

	flags := flagsFromContext(ctx)
	if len(flags) > 0 {
		snapshot.Flags = flags
	}

	if len(snapshot.Values) == 0 {
		snapshot.Values = nil
	}

	return snapshot, nil
}

// Restore restores snapshot values into a new context
func (m *ContextManager) Restore(ctx context.Context, snapshot ContextSnapshot) (context.Context, error) {
	restored := ctx
	codecs := m.codecsSnapshot()

	valueKeys := lo.Keys(snapshot.Values)
	slices.Sort(valueKeys)

	for _, key := range valueKeys {
		codec, exists := codecs[key]
		if !exists {
			continue
		}

		raw := snapshot.Values[key]
		next, err := codec.Restore(restored, raw)
		if err != nil {
			return restored, ErrContextSnapshotRestoreFailed
		}

		restored = next
	}

	flagKeys := lo.Keys(snapshot.Flags)
	slices.Sort(flagKeys)

	for _, flag := range flagKeys {
		if !snapshot.Flags[flag] {
			continue
		}

		restored = WithFlag(restored, flag)
	}

	return restored, nil
}

// WithFlag sets a typed context flag
func WithFlag(ctx context.Context, flag ContextFlag) context.Context {
	flags := flagsFromContext(ctx)
	flags[flag] = true

	return contextx.With(ctx, contextFlagSet{Flags: flags})
}

// HasFlag reports whether a typed context flag is set
func HasFlag(ctx context.Context, flag ContextFlag) bool {
	flags := flagsFromContext(ctx)

	return flags[flag]
}

// flagsFromContext extracts the current context flags and clones the map
func flagsFromContext(ctx context.Context) map[ContextFlag]bool {
	existing, exists := contextx.From[contextFlagSet](ctx)
	if !exists {
		return map[ContextFlag]bool{}
	}

	return cloneFlags(existing.Flags)
}

// cloneFlags clones a flag map
func cloneFlags(flags map[ContextFlag]bool) map[ContextFlag]bool {
	if len(flags) == 0 {
		return map[ContextFlag]bool{}
	}

	return lo.Assign(map[ContextFlag]bool{}, flags)
}

// codecsSnapshot clones the registered codec map for lock-free processing
func (m *ContextManager) codecsSnapshot() map[ContextKey]ContextCodec {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.Assign(map[ContextKey]ContextCodec{}, m.codecs)
}
