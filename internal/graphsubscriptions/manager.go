package graphsubscriptions

import (
	"context"
	"encoding/json"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// NotificationChannelBufferSize is the buffer size for task subscription channels
const NotificationChannelBufferSize = 10

// redisChannelName is the pub/sub channel used to distribute notifications across processes
const redisChannelName = "graphsubscriptions:notifications"

var (
	// globalManager is the singleton subscription manager instance
	globalManager *Manager
	globalMu      sync.RWMutex
)

// Manager manages all active subscriptions for real-time updates
type Manager struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Notification // map of userID to list of notification channels

	// redisClient, when set via WithRedis, distributes published notifications to other processes over Redis
	redisClient *redis.Client
	// instanceID identifies this process's Manager so it can ignore its own messages when they're echoed back by Redis
	// since Publish already delivers to local subscribers directly
	instanceID string
}

// redisEnvelope is the payload published to the Redis notification channel
type redisEnvelope struct {
	OriginID string          `json:"origin_id"`
	UserID   string          `json:"user_id"`
	Payload  json.RawMessage `json:"payload"`
}

// NewManager creates a new subscription manager
func NewManager() *Manager {
	m := &Manager{
		subscribers: make(map[string][]chan Notification),
		instanceID:  uuid.NewString(),
	}

	// Set as global manager
	globalMu.Lock()
	globalManager = m
	globalMu.Unlock()

	return m
}

// WithRedis enables cross-process notification delivery
func (sm *Manager) WithRedis(client *redis.Client) *Manager {
	sm.redisClient = client

	go sm.subscribeRedis(context.Background())

	return sm
}

// subscribeRedis listens for notifications published by other processes and forwards them to local subscribers
func (sm *Manager) subscribeRedis(ctx context.Context) {
	pubsub := sm.redisClient.Subscribe(ctx, redisChannelName)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		log.Error().Err(err).Str("instance_id", sm.instanceID).Msg("graphsubscriptions: failed to subscribe to redis channel")
		return
	}

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				log.Warn().Str("instance_id", sm.instanceID).Msg("graphsubscriptions: redis subscription channel closed")
				return
			}

			var env redisEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				log.Error().Err(err).Msg("graphsubscriptions: failed to unmarshal redis notification envelope")
				continue
			}

			if env.OriginID == sm.instanceID {
				log.Debug().Str("user_id", env.UserID).Msg("graphsubscriptions: skipping self-originated redis message, already delivered locally")
				continue
			}

			sm.dispatchLocal(env.UserID, RawNotification{Payload: env.Payload})
		}
	}
}

// HasRedis reports whether this manager is configured to distribute notifications across processes via redis
func (sm *Manager) HasRedis() bool {
	return sm.redisClient != nil
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
	log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Int("subscriber_count", len(sm.subscribers[userID])).Msg("graphsubscriptions: user subscribed to notifications")
}

// Unsubscribe removes a subscriber
func (sm *Manager) Unsubscribe(userID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	channels, ok := sm.subscribers[userID]
	if !ok {
		log.Info().Str("user_id", userID).Msg("attempted to unsubscribe but no subscribers found")
		return
	}

	// Remove the channel from the list using slices.Delete
	for i, c := range channels {
		if c == ch {
			sm.subscribers[userID] = slices.Delete(channels, i, i+1)
			close(ch)
			log.Debug().Str("user_id", userID).Int("remaining_subscribers", len(sm.subscribers[userID])).Msg("user unsubscribed from notifications")
			break
		}
	}

	// Clean up empty lists
	if len(sm.subscribers[userID]) == 0 {
		delete(sm.subscribers, userID)
		log.Debug().Str("user_id", userID).Msg("no more subscribers for user, removed from map")
	}
}

// Publish sends a notification to all subscribers for that user in this process, and, when
// Redis is configured, to subscribers of other processes as well
func (sm *Manager) Publish(userID string, notification Notification) error {
	sm.dispatchLocal(userID, notification)

	if sm.redisClient == nil {
		log.Debug().Str("user_id", userID).Msg("graphsubscriptions: redis not configured on this manager, notification only delivered to subscribers in this process")
		return nil
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("graphsubscriptions: failed to marshal notification for redis publish")
		return nil
	}

	envelope, err := json.Marshal(redisEnvelope{
		OriginID: sm.instanceID,
		UserID:   userID,
		Payload:  payload,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("graphsubscriptions: failed to marshal redis notification envelope")
		return nil
	}

	_, err = sm.redisClient.Publish(context.Background(), redisChannelName, envelope).Result()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("graphsubscriptions: failed to publish notification to redis")
		return nil
	}

	return nil
}

// dispatchLocal sends a notification to all subscribers for that user within this process
func (sm *Manager) dispatchLocal(userID string, notification Notification) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Int("total_subscribed_users", len(sm.subscribers)).Msg("graphsubscriptions: dispatchLocal called")

	channels, ok := sm.subscribers[userID]
	if !ok {
		// No subscribers for this user in this process, which is fine
		log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Msg("graphsubscriptions: no local subscribers found for user")
		return
	}

	log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Int("subscriber_count", len(channels)).Msg("graphsubscriptions: found local subscribers for user, sending notification")

	// Send to all subscribers
	for i, ch := range channels {
		select {
		case ch <- notification:
			// Successfully sent
			log.Debug().Str("user_id", userID).Int("subscriber_index", i).Msg("notification successfully sent to subscriber")
		default:
			// Channel is full or closed, skip
			log.Info().Str("user_id", userID).Int("subscriber_index", i).Msg("channel closed or full, unable to send notification to subscriber for user")
		}
	}
}
