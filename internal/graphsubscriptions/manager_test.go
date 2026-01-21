package graphsubscriptions

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
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

	// Create a channel with the interface type
	notificationChan := make(chan Notification, NotificationChannelBufferSize)

	// Subscribe
	manager.Subscribe(userID, notificationChan)

	// Verify subscription was added
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 1)
	manager.mu.RUnlock()

	// Create a mock notification
	notification := &generated.Notification{
		ID:    "notification-123",
		Title: "Test Notification",
	}

	// Publish the notification
	err := manager.Publish(userID, notification)
	require.NoError(t, err)

	// Verify the notification was received
	select {
	case receivedNotification := <-notificationChan:
		// Cast back to concrete type for assertions
		concreteNotif, ok := receivedNotification.(*generated.Notification)
		require.True(t, ok, "Should be able to cast to *generated.Notification")
		assert.Equal(t, notification.ID, concreteNotif.ID)
		assert.Equal(t, notification.Title, concreteNotif.Title)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for notification")
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	manager := NewManager()
	userID := "test-user-456"

	notification := &generated.Notification{
		ID:    "notification-456",
		Title: "Test Notification",
	}

	// Publish to user with no subscribers should not error
	err := manager.Publish(userID, notification)
	require.NoError(t, err)
}

func TestUnsubscribe(t *testing.T) {
	manager := NewManager()
	userID := "test-user-789"

	// Create and subscribe a channel
	notificationChan := make(chan Notification, NotificationChannelBufferSize)
	manager.Subscribe(userID, notificationChan)

	// Verify subscription exists
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 1)
	manager.mu.RUnlock()

	// Unsubscribe
	manager.Unsubscribe(userID, notificationChan)

	// Verify subscription was removed
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 0)
	manager.mu.RUnlock()

	// Verify channel is closed
	_, ok := <-notificationChan
	assert.False(t, ok, "Channel should be closed")
}

func TestMultipleSubscribers(t *testing.T) {
	manager := NewManager()
	userID := "test-user-multi"

	// Create multiple channels
	chan1 := make(chan Notification, NotificationChannelBufferSize)
	chan2 := make(chan Notification, NotificationChannelBufferSize)
	chan3 := make(chan Notification, NotificationChannelBufferSize)

	// Subscribe all channels
	manager.Subscribe(userID, chan1)
	manager.Subscribe(userID, chan2)
	manager.Subscribe(userID, chan3)

	// Verify all subscriptions
	manager.mu.RLock()
	assert.Len(t, manager.subscribers[userID], 3)
	manager.mu.RUnlock()

	// Publish a notification
	notification := &generated.Notification{
		ID:    "notification-multi",
		Title: "Multi Subscriber Notification",
	}

	err := manager.Publish(userID, notification)
	require.NoError(t, err)

	// Verify all subscribers received the notification
	for i, ch := range []chan Notification{chan1, chan2, chan3} {
		select {
		case receivedNotification := <-ch:
			concreteNotif, ok := receivedNotification.(*generated.Notification)
			require.True(t, ok, "Subscriber %d: should be able to cast to *generated.Notification", i+1)
			assert.Equal(t, notification.ID, concreteNotif.ID, "Subscriber %d should receive notification", i+1)
		case <-time.After(1 * time.Second):
			t.Fatalf("Subscriber %d timeout waiting for notification", i+1)
		}
	}
}

func TestUnsubscribeNonExistent(t *testing.T) {
	manager := NewManager()
	userID := "test-user-nonexistent"

	notificationChan := make(chan Notification, NotificationChannelBufferSize)

	// Unsubscribe without subscribing should not panic
	require.NotPanics(t, func() {
		manager.Unsubscribe(userID, notificationChan)
	})
}

func TestConcurrentPublish(t *testing.T) {
	manager := NewManager()
	userID := "test-user-concurrent"

	notificationChan := make(chan Notification, 100) // Larger buffer for concurrent test
	manager.Subscribe(userID, notificationChan)

	numGoroutines := 10
	numNotificationsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Publish notifications concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numNotificationsPerGoroutine; j++ {
				notification := &generated.Notification{
					ID:    "notification-" + string(rune(goroutineID)) + "-" + string(rune(j)),
					Title: "Concurrent Notification",
				}
				err := manager.Publish(userID, notification)
				require.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify we received all notifications
	receivedCount := 0
	timeout := time.After(5 * time.Second)

	for receivedCount < numGoroutines*numNotificationsPerGoroutine {
		select {
		case <-notificationChan:
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout: only received %d/%d notifications", receivedCount, numGoroutines*numNotificationsPerGoroutine)
		}
	}

	assert.Equal(t, numGoroutines*numNotificationsPerGoroutine, receivedCount)
}

func TestPublishToFullChannel(t *testing.T) {
	manager := NewManager()
	userID := "test-user-full"

	// Create a small buffer channel
	notificationChan := make(chan Notification, 2)
	manager.Subscribe(userID, notificationChan)

	// Fill the channel
	for i := 0; i < 2; i++ {
		notification := &generated.Notification{
			ID:    "notification-fill-" + string(rune(i)),
			Title: "Fill Notification",
		}
		err := manager.Publish(userID, notification)
		require.NoError(t, err)
	}

	// Try to publish to full channel - should not block or error
	notification := &generated.Notification{
		ID:    "notification-overflow",
		Title: "Overflow Notification",
	}

	done := make(chan bool, 1)
	go func() {
		err := manager.Publish(userID, notification)
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
	assert.Equal(t, 10, NotificationChannelBufferSize)
}
