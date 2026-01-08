package soiree

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestShutdownAll verifies that ShutdownAll closes all registered event buses.
func TestShutdownAll(t *testing.T) {
	b1 := New()
	b2 := New()
	topic := "a"

	errChan := b1.Emit(topic, NewBaseEvent(topic, nil))
	for err := range errChan {
		require.NoError(t, err)
	}

	require.NoError(t, ShutdownAll())

	ch1 := b1.Emit(topic, NewBaseEvent(topic, nil))
	err1 := <-ch1
	require.ErrorIs(t, err1, ErrEmitterClosed)

	ch2 := b2.Emit(topic, NewBaseEvent(topic, nil))
	err2 := <-ch2
	require.ErrorIs(t, err2, ErrEmitterClosed)

	require.NoError(t, ShutdownAll())
}
