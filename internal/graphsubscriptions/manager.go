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

// scope prefixes namespace the subscriber map so a user-targeted notification and an
// org-wide one cannot collide on the same key
const (
	userScopePrefix = "user:"
	orgScopePrefix  = "org:"
)

func userKey(userID string) string { return userScopePrefix + userID }

// org-scoped notifications are readable by every member of the org, matching
// interceptors.NotificationQueryFilter — keep the two in sync if that filter narrows
func orgKey(orgID string) string { return orgScopePrefix + orgID }

func subscriptionKeys(userID, orgID string) []string {
	var keys []string

	if userID != "" {
		keys = append(keys, userKey(userID))
	}

	if orgID != "" {
		keys = append(keys, orgKey(orgID))
	}

	return keys
}

// routingKey picks the single key a notification is delivered on, preferring the user when the
// notification names one so org members do not receive another member's personal notification
func routingKey(userID, orgID string) string {
	if userID != "" {
		return userKey(userID)
	}

	if orgID != "" {
		return orgKey(orgID)
	}

	return ""
}

// Manager manages all active subscriptions for real-time updates
type Manager struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Notification // map of scoped key (user:<id> / org:<id>) to notification channels

	// redisClient, when set via WithRedis, distributes published notifications to other processes over Redis
	redisClient *redis.Client
	// instanceID identifies this process's Manager so it can ignore its own messages when they're echoed back by Redis
	// since Publish already delivers to local subscribers directly
	instanceID string
}

// redisEnvelope is the payload published to the Redis notification channel
type redisEnvelope struct {
	OriginID string `json:"origin_id"`
	Key      string `json:"key"`
	// UserID is written for instances predating scoped keys so a rolling deploy keeps delivering
	// personal notifications in both directions; remove once no such instance is running
	UserID  string          `json:"user_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// envelopeKey resolves the routing key from an envelope, falling back to the pre-scoped-key
// user_id field written by older instances
func envelopeKey(env redisEnvelope) string {
	if env.Key != "" {
		return env.Key
	}

	if env.UserID != "" {
		return userKey(env.UserID)
	}

	return ""
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
				log.Debug().Str("key", env.Key).Msg("graphsubscriptions: skipping self-originated redis message, already delivered locally")
				continue
			}

			key := envelopeKey(env)
			if key == "" {
				log.Warn().Msg("graphsubscriptions: redis envelope carried no routing key, dropping")
				continue
			}

			sm.dispatchLocal(key, RawNotification{Payload: env.Payload})
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

// Subscribe registers a subscriber for notifications addressed to the user directly as well as
// those addressed to their whole organization
func (sm *Manager) Subscribe(userID, orgID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, key := range subscriptionKeys(userID, orgID) {
		sm.subscribers[key] = append(sm.subscribers[key], ch)
		log.Debug().Str("instance_id", sm.instanceID).Str("key", key).Int("subscriber_count", len(sm.subscribers[key])).Msg("graphsubscriptions: subscribed to notifications")
	}
}

// Unsubscribe removes a subscriber and closes its channel. It sweeps every key rather than only
// those derived from userID/orgID: closing while the channel is still reachable under another key
// would panic the next dispatch, so the close must not depend on the caller passing back the same
// scope it subscribed with
func (sm *Manager) Unsubscribe(userID, orgID string, ch chan Notification) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	removed := false

	for key, channels := range sm.subscribers {
		if i := slices.Index(channels, ch); i >= 0 {
			sm.subscribers[key] = slices.Delete(channels, i, i+1)
			removed = true

			log.Debug().Str("key", key).Int("remaining_subscribers", len(sm.subscribers[key])).Msg("unsubscribed from notifications")
		}

		if len(sm.subscribers[key]) == 0 {
			delete(sm.subscribers, key)
		}
	}

	if !removed {
		log.Info().Str("user_id", userID).Str("org_id", orgID).Msg("attempted to unsubscribe but no subscribers found")
		return
	}

	close(ch)
}

// Publish routes a notification to the subscribers of exactly one scope — the user when the
// notification names one, otherwise the owning org — in this process, and, when Redis is
// configured, to subscribers of other processes as well
func (sm *Manager) Publish(userID, orgID string, notification Notification) error {
	key := routingKey(userID, orgID)
	if key == "" {
		log.Debug().Msg("graphsubscriptions: notification has neither user nor owner, nothing to route to")
		return nil
	}

	sm.dispatchLocal(key, notification)

	if sm.redisClient == nil {
		log.Debug().Str("key", key).Msg("graphsubscriptions: redis not configured on this manager, notification only delivered to subscribers in this process")
		return nil
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("graphsubscriptions: failed to marshal notification for redis publish")
		return nil
	}

	envelope, err := json.Marshal(redisEnvelope{
		OriginID: sm.instanceID,
		Key:      key,
		UserID:   userID,
		Payload:  payload,
	})
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("graphsubscriptions: failed to marshal redis notification envelope")
		return nil
	}

	_, err = sm.redisClient.Publish(context.Background(), redisChannelName, envelope).Result()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("graphsubscriptions: failed to publish notification to redis")
		return nil
	}

	return nil
}

// dispatchLocal sends a notification to all subscribers registered under key within this process
func (sm *Manager) dispatchLocal(key string, notification Notification) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Debug().Str("instance_id", sm.instanceID).Str("key", key).Int("total_subscribed_keys", len(sm.subscribers)).Msg("graphsubscriptions: dispatchLocal called")

	channels, ok := sm.subscribers[key]
	if !ok {
		// No subscribers for this key in this process, which is fine
		log.Debug().Str("instance_id", sm.instanceID).Str("key", key).Msg("graphsubscriptions: no local subscribers found for key")
		return
	}

	log.Debug().Str("instance_id", sm.instanceID).Str("key", key).Int("subscriber_count", len(channels)).Msg("graphsubscriptions: found local subscribers for key, sending notification")

	// Send to all subscribers
	for i, ch := range channels {
		select {
		case ch <- notification:
			// Successfully sent
			log.Debug().Str("key", key).Int("subscriber_index", i).Msg("notification successfully sent to subscriber")
		default:
			// buffer is full — a closed channel would panic here rather than land in this branch,
			// which is why Unsubscribe must remove from every key before closing
			log.Info().Str("key", key).Int("subscriber_index", i).Msg("channel full, unable to send notification to subscriber")
		}
	}
}
