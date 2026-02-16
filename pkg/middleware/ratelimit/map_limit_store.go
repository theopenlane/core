package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MapLimitStore represents a data structure for in-memory storage of ratelimiter information
type MapLimitStore struct {
	// data is a map of key to limitValue
	data map[string]limitValue
	// mutex is a mutex for data map
	mutex sync.RWMutex
	// expirationTime is the time after which the data is considered expired
	expirationTime time.Duration
	// ticker is the ticker for periodic cleanup
	ticker *time.Ticker
	// ctx is the context for graceful shutdown
	ctx context.Context
	// cancel is the cancel function for the context
	cancel context.CancelFunc
}

// limitValue represents value of the limit counter
type limitValue struct {
	val        int64
	lastUpdate time.Time
}

// NewMapLimitStore creates new in-memory data store for internal limiter data
func NewMapLimitStore(c context.Context, expirationTime time.Duration, flushInterval time.Duration) (m *MapLimitStore) {
	ctx, cancel := context.WithCancel(c)
	ticker := time.NewTicker(flushInterval)

	m = &MapLimitStore{
		data:           make(map[string]limitValue),
		expirationTime: expirationTime,
		ticker:         ticker,
		ctx:            ctx,
		cancel:         cancel,
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				m.mutex.Lock()

				for key, val := range m.data {
					if val.lastUpdate.Before(time.Now().UTC().Add(-m.expirationTime)) {
						delete(m.data, key)
					}
				}

				m.mutex.Unlock()
			case <-ctx.Done():
				// Context cancelled, exit goroutine
				return
			}
		}
	}()

	return m
}

// Inc increments current window limit counter
func (m *MapLimitStore) Inc(key string, window time.Time) error {
	m.mutex.Lock()

	defer m.mutex.Unlock()

	data := m.data[mapKey(key, window)]
	data.val++
	data.lastUpdate = time.Now().UTC()
	m.data[mapKey(key, window)] = data

	return nil
}

// Get gets value of previous window counter and current window counter
func (m *MapLimitStore) Get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error) {
	m.mutex.RLock()

	defer m.mutex.RUnlock()

	prevValue = m.data[mapKey(key, previousWindow)].val
	currValue = m.data[mapKey(key, currentWindow)].val

	return
}

// Size returns current length of data map
func (m *MapLimitStore) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.data)
}

// mapKey creates a key for the map
func mapKey(key string, window time.Time) string {
	return fmt.Sprintf("%s_%s", key, window.Format(time.RFC3339))
}

// Close stops the background cleanup goroutine and releases resources
func (m *MapLimitStore) Close() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.ticker != nil {
		m.ticker.Stop()
	}
}
