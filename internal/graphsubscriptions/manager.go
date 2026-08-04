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

// subscriberKey identifies a session by the user it belongs to and the org that user is currently
// in, so a session only receives notifications belonging to the org it is viewing
type subscriberKey struct {
	userID string
	orgID  string
}

// Manager manages all active subscriptions for real-time updates
type Manager struct {
	mu          sync.RWMutex
	subscribers map[subscriberKey][]chan Notification // map of session to list of notification channels

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
	OrgID    string          `json:"org_id"`
	Payload  json.RawMessage `json:"payload"`
}

// NewManager creates a new subscription manager
func NewManager() *Manager {
	m := &Manager{
		subscribers: make(map[subscriberKey][]chan Notification),
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
				log.Debug().Str("user_id", env.UserID).Str("org_id", env.OrgID).Msg("graphsubscriptions: skipping self-originated redis message, already delivered locally")
				continue
			}

			if env.UserID == "" && env.OrgID == "" {
				log.Warn().Msg("graphsubscriptions: redis envelope carried no routing target, dropping")
				continue
			}

			sm.dispatchLocal(env.UserID, env.OrgID, RawNotification{Payload: env.Payload})
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

// Subscribe adds a new subscriber for notifications addressed to the user directly as well as
// those addressed to the whole org the user is currently in
func (sm *Manager) Subscribe(userID, orgID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := subscriberKey{userID: userID, orgID: orgID}

	sm.subscribers[key] = append(sm.subscribers[key], ch)
	log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Str("org_id", orgID).Int("subscriber_count", len(sm.subscribers[key])).Msg("graphsubscriptions: user subscribed to notifications")
}

// Unsubscribe removes a subscriber
func (sm *Manager) Unsubscribe(userID, orgID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := subscriberKey{userID: userID, orgID: orgID}

	channels, ok := sm.subscribers[key]
	if !ok {
		log.Info().Str("user_id", userID).Str("org_id", orgID).Msg("attempted to unsubscribe but no subscribers found")
		return
	}

	// Remove the channel from the list using slices.Delete
	for i, c := range channels {
		if c == ch {
			sm.subscribers[key] = slices.Delete(channels, i, i+1)
			close(ch)
			log.Debug().Str("user_id", userID).Str("org_id", orgID).Int("remaining_subscribers", len(sm.subscribers[key])).Msg("user unsubscribed from notifications")
			break
		}
	}

	// Clean up empty lists
	if len(sm.subscribers[key]) == 0 {
		delete(sm.subscribers, key)
		log.Debug().Str("user_id", userID).Str("org_id", orgID).Msg("no more subscribers for session, removed from map")
	}
}

// Publish sends a notification to the subscribers it is addressed to in this process, and, when
// Redis is configured, to subscribers of other processes as well
func (sm *Manager) Publish(userID, orgID string, notification Notification) error {
	if userID == "" && orgID == "" {
		log.Debug().Msg("graphsubscriptions: notification has neither user nor owner, nothing to route to")
		return nil
	}

	sm.dispatchLocal(userID, orgID, notification)

	if sm.redisClient == nil {
		log.Debug().Str("user_id", userID).Str("org_id", orgID).Msg("graphsubscriptions: redis not configured on this manager, notification only delivered to subscribers in this process")
		return nil
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("org_id", orgID).Msg("graphsubscriptions: failed to marshal notification for redis publish")
		return nil
	}

	envelope, err := json.Marshal(redisEnvelope{
		OriginID: sm.instanceID,
		UserID:   userID,
		OrgID:    orgID,
		Payload:  payload,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("org_id", orgID).Msg("graphsubscriptions: failed to marshal redis notification envelope")
		return nil
	}

	_, err = sm.redisClient.Publish(context.Background(), redisChannelName, envelope).Result()
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("org_id", orgID).Msg("graphsubscriptions: failed to publish notification to redis")
		return nil
	}

	return nil
}

// dispatchLocal sends a notification to the subscribers it is addressed to within this process.
// A notification naming a user goes only to that user's sessions in the owning org, otherwise it
// fans out to every session currently in the org
func (sm *Manager) dispatchLocal(userID, orgID string, notification Notification) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Debug().Str("instance_id", sm.instanceID).Str("user_id", userID).Str("org_id", orgID).Int("total_sessions", len(sm.subscribers)).Msg("graphsubscriptions: dispatchLocal called")

	if userID != "" {
		sm.send(sm.subscribers[subscriberKey{userID: userID, orgID: orgID}], notification)
		return
	}

	// org-wide notifications name no user, so every session in the org receives it. This scan is
	// bounded by the sessions open on this process and only runs for org-wide notifications
	for key, channels := range sm.subscribers {
		if key.orgID == orgID {
			sm.send(channels, notification)
		}
	}
}

// send delivers to each channel without blocking, dropping when a subscriber's buffer is full
func (sm *Manager) send(channels []chan Notification, notification Notification) {
	for i, ch := range channels {
		select {
		case ch <- notification:
			// Successfully sent
		default:
			// buffer is full, skip rather than block the publisher
			log.Info().Int("subscriber_index", i).Msg("channel full, unable to send notification to subscriber")
		}
	}
}
