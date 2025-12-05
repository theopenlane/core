package graphsubscriptions

import (
	"slices"
	"sync"

	"github.com/rs/zerolog/log"
)

// TaskChannelBufferSize is the buffer size for task subscription channels
const TaskChannelBufferSize = 10

var (
	// globalManager is the singleton subscription manager instance
	globalManager *Manager
	globalMu      sync.RWMutex
)

// Manager manages all active subscriptions for real-time updates
type Manager struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Notification // map of userID to list of notification channels
}

// NewManager creates a new subscription manager
func NewManager() *Manager {
	m := &Manager{
		subscribers: make(map[string][]chan Notification),
	}

	// Set as global manager
	globalMu.Lock()
	globalManager = m
	globalMu.Unlock()

	return m
}

// GetGlobalManager returns the global subscription manager instance
func GetGlobalManager() *Manager {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalManager
}

// Subscribe adds a new subscriber for a user's notification creations
func (sm *Manager) Subscribe(userID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscribers[userID] = append(sm.subscribers[userID], ch)
}

// Unsubscribe removes a subscriber
func (sm *Manager) Unsubscribe(userID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	channels, ok := sm.subscribers[userID]
	if !ok {
		return
	}

	// Remove the channel from the list using slices.Delete
	for i, c := range channels {
		if c == ch {
			sm.subscribers[userID] = slices.Delete(channels, i, i+1)
			close(ch)
			break
		}
	}

	// Clean up empty lists
	if len(sm.subscribers[userID]) == 0 {
		delete(sm.subscribers, userID)
	}
}

// Publish sends a notification to all subscribers for that user
func (sm *Manager) Publish(userID string, notification Notification) error {
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
		case ch <- notification:
			// Successfully sent
			log.Debug().Str("user_id", userID).Msg("notification successfully sent to subscriber")
		default:
			// Channel is full or closed, skip
			log.Warn().Str("user_id", userID).Msg("channel closed, unable to send notification to subscriber for user")
		}
	}

	return nil
}
