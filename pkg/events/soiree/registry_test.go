package soiree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestShutdownAll verifies that ShutdownAll closes all registered event pools.
func TestShutdownAll(t *testing.T) {
	p1 := NewEventPool()
	p2 := NewEventPool()
	topic := "a"

	// ensure pools accept events before shutdown
	errChan := p1.Emit(topic, NewBaseEvent(topic, nil))
	for err := range errChan {
		require.NoError(t, err)
	}

	require.NoError(t, ShutdownAll())

	// both pools should be closed
	ch1 := p1.Emit(topic, NewBaseEvent(topic, nil))
	err1 := <-ch1
	require.ErrorIs(t, err1, ErrEmitterClosed)

	ch2 := p2.Emit(topic, NewBaseEvent(topic, nil))
	err2 := <-ch2
	require.ErrorIs(t, err2, ErrEmitterClosed)

	// calling ShutdownAll again should be a no-op
	require.NoError(t, ShutdownAll())
}
