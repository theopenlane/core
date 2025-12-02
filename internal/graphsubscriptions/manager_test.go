package graphsubscriptions

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/ent/generated"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	require.NotNil(t, manager)
	require.NotNil(t, manager.subscribers)
	assert.Empty(t, manager.subscribers)
}

func TestSubscribeAndPublish(t *testing.T) {
	manager := NewManager()
	userID := "test-user-123"

	// Create a channel
	taskChan := make(chan *generated.Task, TaskChannelBufferSize)

	// Subscribe
	manager.Subscribe(userID, taskChan)

	// Verify subscription was added
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 1)
	manager.mu.RUnlock()

	// Create a mock task
	task := &generated.Task{
		ID:    "task-123",
		Title: "Test Task",
	}

	// Publish the task
	err := manager.Publish(userID, task)
	require.NoError(t, err)

	// Verify the task was received
	select {
	case receivedTask := <-taskChan:
		assert.Equal(t, task.ID, receivedTask.ID)
		assert.Equal(t, task.Title, receivedTask.Title)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for task")
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	manager := NewManager()
	userID := "test-user-456"

	task := &generated.Task{
		ID:    "task-456",
		Title: "Test Task",
	}

	// Publish to user with no subscribers should not error
	err := manager.Publish(userID, task)
	require.NoError(t, err)
}

func TestUnsubscribe(t *testing.T) {
	manager := NewManager()
	userID := "test-user-789"

	// Create and subscribe a channel
	taskChan := make(chan *generated.Task, TaskChannelBufferSize)
	manager.Subscribe(userID, taskChan)

	// Verify subscription exists
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 1)
	manager.mu.RUnlock()

	// Unsubscribe
	manager.Unsubscribe(userID, taskChan)

	// Verify subscription was removed
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 0)
	manager.mu.RUnlock()

	// Verify channel is closed
	_, ok := <-taskChan
	assert.False(t, ok, "Channel should be closed")
}

func TestMultipleSubscribers(t *testing.T) {
	manager := NewManager()
	userID := "test-user-multi"

	// Create multiple channels
	chan1 := make(chan *generated.Task, TaskChannelBufferSize)
	chan2 := make(chan *generated.Task, TaskChannelBufferSize)
	chan3 := make(chan *generated.Task, TaskChannelBufferSize)

	// Subscribe all channels
	manager.Subscribe(userID, chan1)
	manager.Subscribe(userID, chan2)
	manager.Subscribe(userID, chan3)

	// Verify all subscriptions
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 3)
	manager.mu.RUnlock()

	// Publish a task
	task := &generated.Task{
		ID:    "task-multi",
		Title: "Multi Subscriber Task",
	}

	err := manager.Publish(userID, task)
	require.NoError(t, err)

	// Verify all subscribers received the task
	for i, ch := range []chan *generated.Task{chan1, chan2, chan3} {
		select {
		case receivedTask := <-ch:
			assert.Equal(t, task.ID, receivedTask.ID, "Subscriber %d should receive task", i+1)
		case <-time.After(1 * time.Second):
			t.Fatalf("Subscriber %d timeout waiting for task", i+1)
		}
	}
}

func TestUnsubscribeNonExistent(t *testing.T) {
	manager := NewManager()
	userID := "test-user-nonexistent"

	taskChan := make(chan *generated.Task, TaskChannelBufferSize)

	// Unsubscribe without subscribing should not panic
	require.NotPanics(t, func() {
		manager.Unsubscribe(userID, taskChan)
	})
}

func TestConcurrentPublish(t *testing.T) {
	manager := NewManager()
	userID := "test-user-concurrent"

	taskChan := make(chan *generated.Task, 100) // Larger buffer for concurrent test
	manager.Subscribe(userID, taskChan)

	numGoroutines := 10
	numTasksPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Publish tasks concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numTasksPerGoroutine; j++ {
				task := &generated.Task{
					ID:    "task-" + string(rune(goroutineID)) + "-" + string(rune(j)),
					Title: "Concurrent Task",
				}
				err := manager.Publish(userID, task)
				require.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify we received all tasks
	receivedCount := 0
	timeout := time.After(5 * time.Second)

	for receivedCount < numGoroutines*numTasksPerGoroutine {
		select {
		case <-taskChan:
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout: only received %d/%d tasks", receivedCount, numGoroutines*numTasksPerGoroutine)
		}
	}

	assert.Equal(t, numGoroutines*numTasksPerGoroutine, receivedCount)
}

func TestPublishToFullChannel(t *testing.T) {
	manager := NewManager()
	userID := "test-user-full"

	// Create a small buffer channel
	taskChan := make(chan *generated.Task, 2)
	manager.Subscribe(userID, taskChan)

	// Fill the channel
	for i := 0; i < 2; i++ {
		task := &generated.Task{
			ID:    "task-fill-" + string(rune(i)),
			Title: "Fill Task",
		}
		err := manager.Publish(userID, task)
		require.NoError(t, err)
	}

	// Try to publish to full channel - should not block or error
	task := &generated.Task{
		ID:    "task-overflow",
		Title: "Overflow Task",
	}

	done := make(chan bool, 1)
	go func() {
		err := manager.Publish(userID, task)
		require.NoError(t, err)
		done <- true
	}()

	select {
	case <-done:
		// Should complete without blocking
	case <-time.After(1 * time.Second):
		t.Fatal("Publish blocked on full channel")
	}
}

func TestChannelBufferSize(t *testing.T) {
	// Verify the constant is set to expected value
	assert.Equal(t, 10, TaskChannelBufferSize)
}
