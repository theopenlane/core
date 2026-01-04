package soiree

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

// redisEvent is a struct that represents an event stored in Redis
// it contains the topic of the event and the serialized payload
// it should be used to enqueue events in Redis and to store them in the event store (if you need persistence)
// it is serialized to JSON for storage and deserialization when retrieving events
type redisEvent struct {
	// Topic is the topic of the event
	Topic string `json:"topic"`
	// Payload is the serialized payload of the event
	Payload json.RawMessage `json:"payload"`
	// Properties captures the serialized event properties for replay/idempotency
	Properties json.RawMessage `json:"properties,omitempty"`
}

// RedisStore persists events and results in redis and acts as an event queue
type RedisStore struct {
	client  *redis.Client
	metrics *redisMetrics
}

// NewRedisStore creates a new RedisStore with default metrics
func NewRedisStore(client *redis.Client) *RedisStore {
	return NewRedisStoreWithMetrics(client, defaultRedisMetrics)
}

// NewRedisStoreWithMetrics allows injecting custom metrics (for testing)
func NewRedisStoreWithMetrics(client *redis.Client, metrics *redisMetrics) *RedisStore {
	s := &RedisStore{client: client, metrics: metrics}
	s.initQueueLength()

	return s
}

// SaveEvent enqueues and stores the event
func (s *RedisStore) SaveEvent(e Event) error {
	// Marshal the event payload to JSON
	payload, err := json.Marshal(e.Payload())
	if err != nil {
		return err
	}

	props, propsErr := marshalEventProperties(e)
	if propsErr != nil {
		return propsErr
	}

	// Create a redisEvent instance with the topic and payload
	data, err := json.Marshal(redisEvent{Topic: e.Topic(), Payload: payload, Properties: props})
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Use a Redis pipeline to enqueue the event in both the "soiree:events" and "soiree:queue" lists
	// this allows us to store the event for persistence and also enqueue it for processing
	// the RPush command appends the data to the end of the list
	// we use a pipeline to reduce the number of round trips to Redis
	pipe := s.client.Pipeline()

	pipe.RPush(ctx, "soiree:events", data)
	pipe.RPush(ctx, "soiree:queue", data)

	// we don't need the results of the pipeline commands, so we ignore them
	// we just want to ensure that both commands are executed atomically
	// if one fails, the other will not be executed
	// this ensures that we don't end up with an event in the queue without it being stored
	// caller is responsible for handling errors
	_, err = pipe.Exec(ctx)
	if err == nil {
		s.metrics.redisEventsPersisted.Inc()
		s.metrics.redisQueueLength.Inc()
	}

	return err
}

// SaveHandlerResult stores the result of a listener processing an event
func (s *RedisStore) SaveHandlerResult(e Event, handlerID string, err error) error {
	res := storedResult{Topic: e.Topic(), HandlerID: handlerID, EventID: EventID(e)}
	if err != nil {
		res.Error = err.Error()
	}

	data, marshalErr := json.Marshal(res)
	if marshalErr != nil {
		return marshalErr
	}

	ctx := context.Background()

	pipe := s.client.Pipeline()
	pipe.RPush(ctx, "soiree:results", data)

	if res.EventID != "" && handlerID != "" {
		pipe.SAdd(ctx, redisHandlerDedupKey(res.EventID), handlerID)
	}

	_, pushErr := pipe.Exec(ctx)
	if pushErr == nil {
		s.metrics.redisResultsPersisted.Inc()
	}

	return pushErr
}

// DequeueEvent pops a soiree event from the event queue - party line !
func (s *RedisStore) DequeueEvent(ctx context.Context) (Event, error) {
	// Use BLPop to block until an event is available in the queue
	// this will wait indefinitely until an event is pushed to the queue
	// it returns a slice where the first element is the key and the second is the value (the event)
	// we expect the key to be "soiree:queue" and the value to be a JSON string representing the event
	vals, err := s.client.BLPop(ctx, 0, "soiree:queue").Result()
	if err != nil {
		return nil, err
	}

	// vals is a slice where the first element is the key ("soiree:queue") and the second is the value (the event)
	// we only care about the second element, which is the JSON string representing the event
	// if there are not enough elements, return an error
	if len(vals) < 2 { //nolint:mnd
		return nil, redis.Nil
	}

	var re redisEvent
	if err := json.Unmarshal([]byte(vals[1]), &re); err != nil {
		return nil, err
	}

	var payload any
	// If the payload is not empty, unmarshal it into a generic interface
	if len(re.Payload) > 0 {
		if err = json.Unmarshal(re.Payload, &payload); err != nil {
			return nil, err
		}
	}

	var props Properties
	if len(re.Properties) > 0 {
		if err := json.Unmarshal(re.Properties, &props); err != nil {
			return nil, err
		}
	}

	// decrease queue length and increment dequeued events counter
	s.metrics.redisEventsDequeued.Inc()
	s.metrics.redisQueueLength.Dec()

	event := NewBaseEvent(re.Topic, payload)
	if props != nil {
		event.SetProperties(props)
	}

	return event, nil
}

// Events returns all persisted events
func (s *RedisStore) Events(ctx context.Context) ([]Event, error) {
	// Retrieve all events from the Redis list "soiree:events"
	// using LRange to get all elements from index 0 to -1 (which means all elements)
	// this will return a slice of JSON strings representing the events
	vals, err := s.client.LRange(ctx, "soiree:events", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(vals))
	// Iterate over the values and unmarshal them into redisEvent structs
	// then create BaseEvent instances from them
	// this allows us to return a slice of Event interface types
	// which is what the caller expects
	for _, v := range vals {
		var re redisEvent
		if err := json.Unmarshal([]byte(v), &re); err != nil {
			return nil, err
		}

		var payload any
		// If the payload is not empty, unmarshal it into a generic interface
		if len(re.Payload) > 0 {
			if err := json.Unmarshal(re.Payload, &payload); err != nil {
				return nil, err
			}
		}

		var props Properties
		if len(re.Properties) > 0 {
			if err := json.Unmarshal(re.Properties, &props); err != nil {
				return nil, err
			}
		}

		event := NewBaseEvent(re.Topic, payload)
		if props != nil {
			event.SetProperties(props)
		}

		events = append(events, event)
	}

	return events, nil
}

// Results returns all persisted listener results
func (s *RedisStore) Results(ctx context.Context) ([]storedResult, error) {
	vals, err := s.client.LRange(ctx, "soiree:results", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	results := make([]storedResult, 0, len(vals))
	for _, v := range vals {
		var r storedResult
		if err := json.Unmarshal([]byte(v), &r); err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	return results, nil
}

// HandlerSucceeded reports whether the handler has already succeeded for the given event ID.
func (s *RedisStore) HandlerSucceeded(ctx context.Context, eventID string, handlerID string) (bool, error) {
	if strings.TrimSpace(eventID) == "" || strings.TrimSpace(handlerID) == "" {
		return false, nil
	}

	return s.client.SIsMember(ctx, redisHandlerDedupKey(eventID), handlerID).Result()
}

// redisHandlerDedupKey returns the Redis key used for handler deduplication for a given event ID
func redisHandlerDedupKey(eventID string) string {
	return fmt.Sprintf("soiree:dedup:%s", eventID)
}

// marshalEventProperties marshals the event properties to JSON, ensuring at least the EventID is included
func marshalEventProperties(event Event) (json.RawMessage, error) {
	props := event.Properties()
	if props == nil {
		if id := EventID(event); id != "" {
			props = Properties{PropertyEventID: id}
		}
	}

	data, err := json.Marshal(props)
	if err == nil {
		return data, nil
	}

	id := EventID(event)
	if id == "" {
		return nil, err
	}

	fallback, fallbackErr := json.Marshal(Properties{PropertyEventID: id})
	if fallbackErr != nil {
		return nil, err
	}

	return fallback, nil
}
