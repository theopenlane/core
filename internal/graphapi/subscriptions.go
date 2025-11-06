package graphapi

import (
	"fmt"
	"sync"

	"github.com/theopenlane/core/internal/ent/generated"
)

// SubscriptionManager manages all active subscriptions for real-time updates
type SubscriptionManager struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *generated.Task // map of userID to list of task channels
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscribers: make(map[string][]chan *generated.Task),
	}
}

// Subscribe adds a new subscriber for a user's task creations
func (sm *SubscriptionManager) Subscribe(userID string, ch chan *generated.Task) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscribers[userID] = append(sm.subscribers[userID], ch)
}

// Unsubscribe removes a subscriber
func (sm *SubscriptionManager) Unsubscribe(userID string, ch chan *generated.Task) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	channels, ok := sm.subscribers[userID]
	if !ok {
		return
	}

	// Remove the channel from the list
	for i, c := range channels {
		if c == ch {
			sm.subscribers[userID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}

	// Clean up empty lists
	if len(sm.subscribers[userID]) == 0 {
		delete(sm.subscribers, userID)
	}
}

// Publish sends a task to all subscribers for that user
func (sm *SubscriptionManager) Publish(userID string, task *generated.Task) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	channels, ok := sm.subscribers[userID]
	if !ok {
		// No subscribers for this user, which is fine
		return nil
	}

	// Send to all subscribers
	for _, ch := range channels {
		select {
		case ch <- task:
			// Successfully sent
		default:
			// Channel is full or closed, skip
			fmt.Printf("warning: could not send task to subscriber for user %s\n", userID)
		}
	}

	return nil
}
