package graphsubscriptions

import (
	"sync"
	"testing"
	"time"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	assert.Assert(t, manager != nil)
	assert.Assert(t, manager.subscribers != nil)
	assert.Check(t, is.Len(manager.subscribers, 0))
}

func TestSubscribeAndPublish(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-123"
	orgID := "test-org"

	// Create a channel with the interface type
	notificationChan := make(chan Notification, NotificationChannelBufferSize)

	// Subscribe
	manager.Subscribe(userID, orgID, notificationChan)

	// Verify subscription was added
	manager.mu.RLock()
	assert.Check(t, is.Len(manager.subscribers[subscriberKey{userID: userID, orgID: orgID}], 1))
	manager.mu.RUnlock()

	// Create a mock notification
	notification := &generated.Notification{
		ID:    "notification-123",
		Title: "Test Notification",
	}

	// Publish the notification
	err := manager.Publish(userID, orgID, notification)
	assert.NilError(t, err)

	// Verify the notification was received
	select {
	case receivedNotification := <-notificationChan:
		// Cast back to concrete type for assertions
		concreteNotif, ok := receivedNotification.(*generated.Notification)
		assert.Assert(t, ok, "Should be able to cast to *generated.Notification")
		assert.Check(t, is.Equal(notification.ID, concreteNotif.ID))
		assert.Check(t, is.Equal(notification.Title, concreteNotif.Title))
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for notification")
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-456"
	orgID := "test-org"

	notification := &generated.Notification{
		ID:    "notification-456",
		Title: "Test Notification",
	}

	// Publish to user with no subscribers should not error
	err := manager.Publish(userID, orgID, notification)
	assert.NilError(t, err)
}

func TestUnsubscribe(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-789"
	orgID := "test-org"

	// Create and subscribe a channel
	notificationChan := make(chan Notification, NotificationChannelBufferSize)
	manager.Subscribe(userID, orgID, notificationChan)

	// Verify subscription exists
	manager.mu.RLock()
	assert.Check(t, is.Len(manager.subscribers[subscriberKey{userID: userID, orgID: orgID}], 1))
	manager.mu.RUnlock()

	// Unsubscribe
	manager.Unsubscribe(userID, orgID, notificationChan)

	// Verify subscription was removed
	manager.mu.RLock()
	assert.Check(t, is.Len(manager.subscribers[subscriberKey{userID: userID, orgID: orgID}], 0))
	manager.mu.RUnlock()

	// Verify channel is closed
	_, ok := <-notificationChan
	assert.Check(t, !ok, "Channel should be closed")
}

func TestMultipleSubscribers(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-multi"
	orgID := "test-org"

	// Create multiple channels
	chan1 := make(chan Notification, NotificationChannelBufferSize)
	chan2 := make(chan Notification, NotificationChannelBufferSize)
	chan3 := make(chan Notification, NotificationChannelBufferSize)

	// Subscribe all channels
	manager.Subscribe(userID, orgID, chan1)
	manager.Subscribe(userID, orgID, chan2)
	manager.Subscribe(userID, orgID, chan3)

	// Verify all subscriptions
	manager.mu.RLock()
	assert.Check(t, is.Len(manager.subscribers[subscriberKey{userID: userID, orgID: orgID}], 3))
	manager.mu.RUnlock()

	// Publish a notification
	notification := &generated.Notification{
		ID:    "notification-multi",
		Title: "Multi Subscriber Notification",
	}

	err := manager.Publish(userID, orgID, notification)
	assert.NilError(t, err)

	// Verify all subscribers received the notification
	for i, ch := range []chan Notification{chan1, chan2, chan3} {
		select {
		case receivedNotification := <-ch:
			concreteNotif, ok := receivedNotification.(*generated.Notification)
			assert.Assert(t, ok, "Subscriber %d: should be able to cast to *generated.Notification", i+1)
			assert.Check(t, is.Equal(notification.ID, concreteNotif.ID), "Subscriber %d should receive notification", i+1)
		case <-time.After(1 * time.Second):
			t.Fatalf("Subscriber %d timeout waiting for notification", i+1)
		}
	}
}

func TestUnsubscribeNonExistent(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-nonexistent"
	orgID := "test-org"

	notificationChan := make(chan Notification, NotificationChannelBufferSize)

	// Unsubscribe without subscribing should not panic
	manager.Unsubscribe(userID, orgID, notificationChan)
}

func TestConcurrentPublish(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-concurrent"
	orgID := "test-org"

	notificationChan := make(chan Notification, 100) // Larger buffer for concurrent test
	manager.Subscribe(userID, orgID, notificationChan)

	numGoroutines := 10
	numNotificationsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Publish notifications concurrently
	for i := range numGoroutines {
		go func(goroutineID int) {
			defer wg.Done()
			for j := range numNotificationsPerGoroutine {
				notification := &generated.Notification{
					ID:    "notification-" + string(rune(goroutineID)) + "-" + string(rune(j)),
					Title: "Concurrent Notification",
				}
				err := manager.Publish(userID, orgID, notification)
				assert.NilError(t, err)
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

	assert.Check(t, is.Equal(numGoroutines*numNotificationsPerGoroutine, receivedCount))
}

func TestPublishToFullChannel(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "test-user-full"
	orgID := "test-org"

	// Create a small buffer channel
	notificationChan := make(chan Notification, 2)
	manager.Subscribe(userID, orgID, notificationChan)

	// Fill the channel
	for i := range 2 {
		notification := &generated.Notification{
			ID:    "notification-fill-" + string(rune(i)),
			Title: "Fill Notification",
		}
		err := manager.Publish(userID, orgID, notification)
		assert.NilError(t, err)
	}

	// Try to publish to full channel - should not block or error
	notification := &generated.Notification{
		ID:    "notification-overflow",
		Title: "Overflow Notification",
	}

	done := make(chan bool, 1)
	go func() {
		err := manager.Publish(userID, orgID, notification)
		assert.NilError(t, err)
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
	t.Parallel()

	// Verify the constant is set to expected value
	assert.Check(t, is.Equal(10, NotificationChannelBufferSize))
}

// receives reports whether a notification with the given id arrived on the channel
func receives(t *testing.T, ch chan Notification, id string) bool {
	t.Helper()

	select {
	case received := <-ch:
		notif, ok := received.(*generated.Notification)
		assert.Assert(t, ok)
		assert.Check(t, is.Equal(id, notif.ID))

		return true
	case <-time.After(100 * time.Millisecond):
		return false
	}
}

func TestPublishOrgWide(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	orgID := "org-a"

	// two different users with a session in the same org, plus one in another org
	memberOne := make(chan Notification, NotificationChannelBufferSize)
	memberTwo := make(chan Notification, NotificationChannelBufferSize)
	outsider := make(chan Notification, NotificationChannelBufferSize)

	manager.Subscribe("user-1", orgID, memberOne)
	manager.Subscribe("user-2", orgID, memberTwo)
	manager.Subscribe("user-3", "org-b", outsider)

	// an org-wide notification names no user
	notification := &generated.Notification{ID: "org-wide", Title: "Organization ready"}
	assert.NilError(t, manager.Publish("", orgID, notification))

	assert.Check(t, receives(t, memberOne, "org-wide"), "member of the org should receive it")
	assert.Check(t, receives(t, memberTwo, "org-wide"), "every member of the org should receive it")
	assert.Check(t, !receives(t, outsider, "org-wide"), "member of another org should not receive it")
}

func TestPublishUserScopedToOrg(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	userID := "user-multi-org"

	// the same user with a session open in each of two orgs
	inOrgA := make(chan Notification, NotificationChannelBufferSize)
	inOrgB := make(chan Notification, NotificationChannelBufferSize)

	manager.Subscribe(userID, "org-a", inOrgA)
	manager.Subscribe(userID, "org-b", inOrgB)

	notification := &generated.Notification{ID: "personal", Title: "Assigned to you"}
	assert.NilError(t, manager.Publish(userID, "org-a", notification))

	assert.Check(t, receives(t, inOrgA, "personal"), "session in the owning org should receive it")
	assert.Check(t, !receives(t, inOrgB, "personal"), "session in another org should not receive it")
}

func TestPublishOrgWideDoesNotReachSessionWithoutOrg(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	orgID := "org-a"

	// the resolver rejects callers that have no org, this is the backstop if one ever subscribes
	noOrg := make(chan Notification, NotificationChannelBufferSize)
	manager.Subscribe("user-without-org", "", noOrg)

	notification := &generated.Notification{ID: "org-wide", Title: "Organization ready"}
	assert.NilError(t, manager.Publish("", orgID, notification))

	assert.Check(t, !receives(t, noOrg, "org-wide"), "session with no org should not receive org notifications")
}

func TestPublishNoRoutingTarget(t *testing.T) {
	t.Parallel()

	manager := NewManager()

	ch := make(chan Notification, NotificationChannelBufferSize)
	manager.Subscribe("user-1", "org-a", ch)

	// nothing to route on, so nothing is delivered
	assert.NilError(t, manager.Publish("", "", &generated.Notification{ID: "orphan"}))

	assert.Check(t, !receives(t, ch, "orphan"))
}
