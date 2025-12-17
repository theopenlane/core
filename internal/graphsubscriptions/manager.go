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
	log.Info().Str("user_id", userID).Int("subscriber_count", len(sm.subscribers[userID])).Msg("user subscribed to notifications")
}

// Unsubscribe removes a subscriber
func (sm *Manager) Unsubscribe(userID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	channels, ok := sm.subscribers[userID]
	if !ok {
		log.Warn().Str("user_id", userID).Msg("attempted to unsubscribe but no subscribers found")
		return
	}

	// Remove the channel from the list using slices.Delete
	for i, c := range channels {
		if c == ch {
			sm.subscribers[userID] = slices.Delete(channels, i, i+1)
			close(ch)
			log.Info().Str("user_id", userID).Int("remaining_subscribers", len(sm.subscribers[userID])).Msg("user unsubscribed from notifications")
			break
		}
	}

	// Clean up empty lists
	if len(sm.subscribers[userID]) == 0 {
		delete(sm.subscribers, userID)
		log.Info().Str("user_id", userID).Msg("no more subscribers for user, removed from map")
	}
}

// Publish sends a notification to all subscribers for that user
func (sm *Manager) Publish(userID string, notification Notification) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Info().Str("user_id", userID).Int("total_subscribers", len(sm.subscribers)).Msg("Publish called")

	channels, ok := sm.subscribers[userID]
	if !ok {
		// No subscribers for this user, which is fine
		log.Warn().Str("user_id", userID).Msg("no subscribers found for user")
		return nil
	}

	log.Info().Str("user_id", userID).Int("subscriber_count", len(channels)).Msg("found subscribers for user, sending notification")

	// Send to all subscribers
	for i, ch := range channels {
		select {
		case ch <- notification:
			// Successfully sent
			log.Info().Str("user_id", userID).Int("subscriber_index", i).Msg("notification successfully sent to subscriber")
		default:
			// Channel is full or closed, skip
			log.Warn().Str("user_id", userID).Int("subscriber_index", i).Msg("channel closed or full, unable to send notification to subscriber for user")
		}
	}

	return nil
}
