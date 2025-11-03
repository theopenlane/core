package soiree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestShutdownAll verifies that ShutdownAll closes all registered event pools.
func TestShutdownAll(t *testing.T) {
	p1 := NewEventPool()
	p2 := NewEventPool()
	topic := NewEventTopic("a")

	// ensure pools accept events before shutdown
	errChan := EmitTopic(p1, topic, Event(NewBaseEvent(topic.Name(), nil)))
	for err := range errChan {
		require.NoError(t, err)
	}

	require.NoError(t, ShutdownAll())

	// both pools should be closed
	ch1 := EmitTopic(p1, topic, Event(NewBaseEvent(topic.Name(), nil)))
	err1 := <-ch1
	require.ErrorIs(t, err1, ErrEmitterClosed)

	ch2 := EmitTopic(p2, topic, Event(NewBaseEvent(topic.Name(), nil)))
	err2 := <-ch2
	require.ErrorIs(t, err2, ErrEmitterClosed)

	// calling ShutdownAll again should be a no-op
	require.NoError(t, ShutdownAll())
}
