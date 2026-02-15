//go:build integration

package gala

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// benchPayload is the standard payload used in benchmark tests.
type benchPayload struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

// benchMetrics collects performance metrics during benchmark runs.
type benchMetrics struct {
	emitted      atomic.Int64
	dispatched   atomic.Int64
	succeeded    atomic.Int64
	failed       atomic.Int64
	retried      atomic.Int64
	panicked     atomic.Int64
	totalLatency atomic.Int64
}

func (m *benchMetrics) report(t testing.TB) {
	t.Helper()

	emitted := m.emitted.Load()
	dispatched := m.dispatched.Load()
	succeeded := m.succeeded.Load()
	failed := m.failed.Load()
	retried := m.retried.Load()
	panicked := m.panicked.Load()
	totalLatency := m.totalLatency.Load()

	t.Logf("Metrics:")
	t.Logf("  Emitted:    %d", emitted)
	t.Logf("  Dispatched: %d", dispatched)
	t.Logf("  Succeeded:  %d", succeeded)
	t.Logf("  Failed:     %d", failed)
	t.Logf("  Retried:    %d", retried)
	t.Logf("  Panicked:   %d", panicked)

	if dispatched > 0 {
		avgLatency := time.Duration(totalLatency / dispatched)
		t.Logf("  Avg Latency: %v", avgLatency)
	}
}

// soireeMetrics collects performance metrics for soiree benchmark runs.
type soireeMetrics struct {
	emitted      atomic.Int64
	dispatched   atomic.Int64
	succeeded    atomic.Int64
	failed       atomic.Int64
	retried      atomic.Int64
	panicked     atomic.Int64
	totalLatency atomic.Int64
}

func (m *soireeMetrics) report(t testing.TB) {
	t.Helper()

	emitted := m.emitted.Load()
	dispatched := m.dispatched.Load()
	succeeded := m.succeeded.Load()
	failed := m.failed.Load()
	retried := m.retried.Load()
	panicked := m.panicked.Load()
	totalLatency := m.totalLatency.Load()

	t.Logf("Soiree Metrics:")
	t.Logf("  Emitted:    %d", emitted)
	t.Logf("  Dispatched: %d", dispatched)
	t.Logf("  Succeeded:  %d", succeeded)
	t.Logf("  Failed:     %d", failed)
	t.Logf("  Retried:    %d", retried)
	t.Logf("  Panicked:   %d", panicked)

	if dispatched > 0 {
		avgLatency := time.Duration(totalLatency / dispatched)
		t.Logf("  Avg Latency: %v", avgLatency)
	}
}

// newBenchSoiree creates a soiree EventBus for benchmarking.
func newBenchSoiree(tb testing.TB, workers int) *soiree.EventBus {
	tb.Helper()

	if workers <= 0 {
		workers = 100
	}

	bus := soiree.New(
		soiree.Workers(workers),
	)

	tb.Cleanup(func() {
		bus.Close()
	})

	return bus
}

// newRiverBackedGala creates a gala instance backed by a real PostgreSQL/River setup.
// This is a convenience wrapper around NewTestGala for benchmark tests.
// Uses aggressive fetch settings for benchmarking throughput.
func newRiverBackedGala(t *testing.T, workerCount int, optimized bool) *TestGalaFixture {
	t.Helper()

	opts := []TestGalaOption{
		WithTestQueueName("gala_benchmark_test"),
		WithTestWorkerCount(workerCount),
		WithTestMaxRetries(3),
	}

	if optimized {
		opts = append(opts,
			WithTestFetchCooldown(time.Millisecond),        // 1ms = River's minimum
			WithTestFetchPollInterval(time.Millisecond),    // 1ms fallback polling
		)
	}

	return NewTestGala(t, opts...)
}

// benchScenario defines parameters for a benchmark scenario
type benchScenario struct {
	name        string
	numEvents   int
	numWorkers  int
	numTopics   int
	numEmitters int // 1 = sequential, >1 = concurrent goroutines
}

// TestIntegrationGalaVsSoireeScenarios runs comparative benchmarks across scenarios.
// Scenarios are designed to isolate variables:
//   - Sequential vs concurrent emission
//   - Single topic vs multi-topic (contention)
//   - Worker scaling
//   - River fetch optimization (default vs aggressive)
func TestIntegrationGalaVsSoireeScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	scenarios := []benchScenario{
		{
			name:        "sequential_multi_topic",
			numEvents:   500,
			numWorkers:  20,
			numTopics:   5,
			numEmitters: 1,
		},
		{
			name:        "sequential_single_topic",
			numEvents:   500,
			numWorkers:  20,
			numTopics:   1,
			numEmitters: 1,
		},
		{
			name:        "concurrent_multi_topic",
			numEvents:   1000,
			numWorkers:  50,
			numTopics:   5,
			numEmitters: 50,
		},
		{
			name:        "concurrent_single_topic",
			numEvents:   1000,
			numWorkers:  50,
			numTopics:   1,
			numEmitters: 50,
		},
	}

	// Run with optimized fetch settings
	t.Run("optimized_fetch", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				runGalaScenario(t, scenario, true)
				runSoireeScenario(t, scenario)
			})
		}
	})

	// Run with default fetch settings for comparison
	t.Run("default_fetch", func(t *testing.T) {
		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				runGalaScenario(t, scenario, false)
				// Skip soiree - results are identical
			})
		}
	})
}

func runGalaScenario(t *testing.T, scenario benchScenario, optimized bool) {
	t.Helper()

	fixture := newRiverBackedGala(t, scenario.numWorkers, optimized)
	runtime := fixture.Gala

	var dispatched atomic.Int64

	// Register topics and listeners
	topicNames := make([]TopicName, scenario.numTopics)
	for i := range scenario.numTopics {
		topic := Topic[benchPayload]{Name: TopicName(fmt.Sprintf("bench.gala.%s.%d", scenario.name, i))}
		topicNames[i] = topic.Name

		if err := RegisterTopic(runtime.Registry(), Registration[benchPayload]{
			Topic: topic,
			Codec: JSONCodec[benchPayload]{},
		}); err != nil {
			t.Fatalf("failed to register topic: %v", err)
		}

		if _, err := AttachListener(runtime.Registry(), Definition[benchPayload]{
			Topic: topic,
			Name:  fmt.Sprintf("bench.gala.%s.%d.listener", scenario.name, i),
			Handle: func(_ HandlerContext, _ benchPayload) error {
				dispatched.Add(1)
				return nil
			},
		}); err != nil {
			t.Fatalf("failed to register listener: %v", err)
		}
	}

	var emitted atomic.Int64
	startTime := time.Now()

	if scenario.numEmitters == 1 {
		// Sequential emission from single goroutine
		for i := range scenario.numEvents {
			topicName := topicNames[i%scenario.numTopics]
			receipt := runtime.EmitWithHeaders(context.Background(), topicName, benchPayload{
				ID:        NewEventID().String(),
				Timestamp: time.Now().UnixNano(),
				Data:      fmt.Sprintf("event-%d", i),
			}, Headers{})
			if receipt.Err == nil {
				emitted.Add(1)
			}
		}
	} else {
		// Concurrent emission from multiple goroutines
		var wg sync.WaitGroup
		eventsPerEmitter := scenario.numEvents / scenario.numEmitters
		startBarrier := make(chan struct{})

		for emitterID := range scenario.numEmitters {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				<-startBarrier

				for i := range eventsPerEmitter {
					topicName := topicNames[(id*eventsPerEmitter+i)%scenario.numTopics]
					receipt := runtime.EmitWithHeaders(context.Background(), topicName, benchPayload{
						ID:        NewEventID().String(),
						Timestamp: time.Now().UnixNano(),
						Data:      fmt.Sprintf("event-%d-%d", id, i),
					}, Headers{})
					if receipt.Err == nil {
						emitted.Add(1)
					}
				}
			}(emitterID)
		}

		close(startBarrier)
		wg.Wait()
	}

	// Wait for processing
	expectedEvents := scenario.numEvents
	if scenario.numEmitters > 1 {
		expectedEvents = (scenario.numEvents / scenario.numEmitters) * scenario.numEmitters
	}

	deadline := time.Now().Add(30 * time.Second)
	for dispatched.Load() < int64(expectedEvents) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	elapsed := time.Since(startTime)

	t.Logf("Gala Results:")
	t.Logf("  Events:     %d", expectedEvents)
	t.Logf("  Workers:    %d", scenario.numWorkers)
	t.Logf("  Topics:     %d", scenario.numTopics)
	t.Logf("  Emitters:   %d", scenario.numEmitters)
	t.Logf("  Duration:   %v", elapsed)
	t.Logf("  Emitted:    %d", emitted.Load())
	t.Logf("  Dispatched: %d", dispatched.Load())
	t.Logf("  Throughput: %.2f events/sec", float64(expectedEvents)/elapsed.Seconds())
}

func runSoireeScenario(t *testing.T, scenario benchScenario) {
	t.Helper()

	bus := newBenchSoiree(t, scenario.numWorkers)

	var dispatched atomic.Int64

	// Register topics and listeners
	topicNames := make([]string, scenario.numTopics)
	for i := range scenario.numTopics {
		topicName := fmt.Sprintf("bench.soiree.%s.%d", scenario.name, i)
		topicNames[i] = topicName

		if _, err := bus.On(topicName, func(_ *soiree.EventContext) error {
			dispatched.Add(1)
			return nil
		}); err != nil {
			t.Fatalf("failed to register listener: %v", err)
		}
	}

	var emitted atomic.Int64
	startTime := time.Now()

	if scenario.numEmitters == 1 {
		// Sequential emission from single goroutine
		for i := range scenario.numEvents {
			topicName := topicNames[i%scenario.numTopics]
			_ = bus.Emit(topicName, benchPayload{
				ID:        NewEventID().String(),
				Timestamp: time.Now().UnixNano(),
				Data:      fmt.Sprintf("event-%d", i),
			})
			emitted.Add(1)
		}
	} else {
		// Concurrent emission from multiple goroutines
		var wg sync.WaitGroup
		eventsPerEmitter := scenario.numEvents / scenario.numEmitters
		startBarrier := make(chan struct{})

		for emitterID := range scenario.numEmitters {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				<-startBarrier

				for i := range eventsPerEmitter {
					topicName := topicNames[(id*eventsPerEmitter+i)%scenario.numTopics]
					_ = bus.Emit(topicName, benchPayload{
						ID:        NewEventID().String(),
						Timestamp: time.Now().UnixNano(),
						Data:      fmt.Sprintf("event-%d-%d", id, i),
					})
					emitted.Add(1)
				}
			}(emitterID)
		}

		close(startBarrier)
		wg.Wait()
	}

	bus.WaitForIdle()
	elapsed := time.Since(startTime)

	expectedEvents := scenario.numEvents
	if scenario.numEmitters > 1 {
		expectedEvents = (scenario.numEvents / scenario.numEmitters) * scenario.numEmitters
	}

	t.Logf("Soiree Results:")
	t.Logf("  Events:     %d", expectedEvents)
	t.Logf("  Workers:    %d", scenario.numWorkers)
	t.Logf("  Topics:     %d", scenario.numTopics)
	t.Logf("  Emitters:   %d", scenario.numEmitters)
	t.Logf("  Duration:   %v", elapsed)
	t.Logf("  Emitted:    %d", emitted.Load())
	t.Logf("  Dispatched: %d", dispatched.Load())
	t.Logf("  Throughput: %.2f events/sec", float64(expectedEvents)/elapsed.Seconds())
}

// TestIntegrationGalaRetryBehavior validates retry mechanics with River.
func TestIntegrationGalaRetryBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	const (
		numEvents      = 20
		failUntilRetry = 2
	)

	fixture := newRiverBackedGala(t, 10, true)
	runtime := fixture.Gala

	metrics := &benchMetrics{}
	attemptTracker := &sync.Map{}

	topic := Topic[benchPayload]{Name: "integration.retry.gala"}

	if err := RegisterTopic(runtime.Registry(), Registration[benchPayload]{
		Topic: topic,
		Codec: JSONCodec[benchPayload]{},
	}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if _, err := AttachListener(runtime.Registry(), Definition[benchPayload]{
		Topic: topic,
		Name:  "integration.retry.gala.listener",
		Handle: func(_ HandlerContext, payload benchPayload) error {
			metrics.dispatched.Add(1)

			countPtr, _ := attemptTracker.LoadOrStore(payload.ID, new(atomic.Int32))
			attempts := countPtr.(*atomic.Int32).Add(1)

			if attempts <= failUntilRetry {
				metrics.failed.Add(1)
				metrics.retried.Add(1)
				return fmt.Errorf("intentional failure attempt %d", attempts)
			}

			metrics.succeeded.Add(1)
			return nil
		},
	}); err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	// Start workers
	if err := runtime.StartWorkers(context.Background()); err != nil {
		t.Fatalf("failed to start workers: %v", err)
	}

	startTime := time.Now()

	// Emit events
	for i := range numEvents {
		receipt := runtime.EmitWithHeaders(context.Background(), topic.Name, benchPayload{
			ID:        fmt.Sprintf("retry-event-%d", i),
			Timestamp: time.Now().UnixNano(),
		}, Headers{})

		if receipt.Err != nil {
			t.Errorf("emit failed: %v", receipt.Err)
		}
	}

	// Wait for all events to be processed (including retries)
	expectedSuccesses := int64(numEvents)
	deadline := time.Now().Add(60 * time.Second)
	for metrics.succeeded.Load() < expectedSuccesses && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}

	elapsed := time.Since(startTime)

	t.Logf("Gala (River-backed) Retry Results:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Events:   %d", numEvents)
	metrics.report(t)

	if metrics.succeeded.Load() != expectedSuccesses {
		t.Errorf("expected %d succeeded, got %d", expectedSuccesses, metrics.succeeded.Load())
	}

	expectedRetries := int64(numEvents * failUntilRetry)
	if metrics.retried.Load() != expectedRetries {
		t.Errorf("expected %d retries, got %d", expectedRetries, metrics.retried.Load())
	}
}

// TestCompareRetryBehavior compares retry mechanics between gala (in-memory) and soiree.
// This test uses in-memory gala dispatch for faster comparison.
func TestCompareRetryBehavior(t *testing.T) {
	const (
		numEvents      = 100
		failUntilRetry = 2
	)

	t.Run("gala_inmemory", func(t *testing.T) {
		metrics := &benchMetrics{}
		runtime := newBenchGala(t, nil)

		topic := Topic[benchPayload]{Name: "compare.retry.gala"}
		registerBenchTopic(t, runtime, topic)

		attemptTracker := &sync.Map{}

		if _, err := AttachListener(runtime.Registry(), Definition[benchPayload]{
			Topic: topic,
			Name:  "compare.retry.gala.listener",
			Handle: func(_ HandlerContext, payload benchPayload) error {
				metrics.dispatched.Add(1)

				countPtr, _ := attemptTracker.LoadOrStore(payload.ID, new(atomic.Int32))
				attempts := countPtr.(*atomic.Int32).Add(1)

				if attempts <= failUntilRetry {
					metrics.failed.Add(1)
					metrics.retried.Add(1)
					return fmt.Errorf("intentional failure attempt %d", attempts)
				}

				metrics.succeeded.Add(1)
				return nil
			},
		}); err != nil {
			t.Fatalf("failed to register listener: %v", err)
		}

		startTime := time.Now()

		for i := range numEvents {
			payload := benchPayload{
				ID:        fmt.Sprintf("event-%d", i),
				Timestamp: time.Now().UnixNano(),
			}

			encodedPayload, err := runtime.Registry().EncodePayload(topic.Name, payload)
			if err != nil {
				t.Fatalf("failed to encode payload: %v", err)
			}

			envelope := Envelope{
				ID:      NewEventID(),
				Topic:   topic.Name,
				Payload: encodedPayload,
			}

			for attempt := 1; attempt <= failUntilRetry+1; attempt++ {
				err := runtime.DispatchEnvelope(context.Background(), envelope)
				if err == nil {
					break
				}

				var listenerErr ListenerError
				if !errors.As(err, &listenerErr) {
					t.Errorf("expected ListenerError, got %T", err)
					break
				}
			}
		}

		elapsed := time.Since(startTime)
		t.Logf("Gala (in-memory) Retry Results (duration: %v):", elapsed)
		metrics.report(t)
	})

	t.Run("soiree", func(t *testing.T) {
		metrics := &soireeMetrics{}
		bus := soiree.New(
			soiree.Workers(10),
			soiree.Retry(failUntilRetry+1, func() backoff.BackOff {
				return &backoff.ZeroBackOff{}
			}),
		)
		defer bus.Close()

		topicName := "compare.retry.soiree"
		attemptTracker := &sync.Map{}

		if _, err := bus.On(topicName, func(ctx *soiree.EventContext) error {
			metrics.dispatched.Add(1)

			payload, ok := soiree.PayloadAs[benchPayload](ctx)
			if !ok {
				return errors.New("invalid payload type")
			}

			countPtr, _ := attemptTracker.LoadOrStore(payload.ID, new(atomic.Int32))
			attempts := countPtr.(*atomic.Int32).Add(1)

			if attempts <= failUntilRetry {
				metrics.failed.Add(1)
				metrics.retried.Add(1)
				return fmt.Errorf("intentional failure attempt %d", attempts)
			}

			metrics.succeeded.Add(1)
			return nil
		}); err != nil {
			t.Fatalf("failed to register listener: %v", err)
		}

		startTime := time.Now()

		for i := range numEvents {
			payload := benchPayload{
				ID:        fmt.Sprintf("event-%d", i),
				Timestamp: time.Now().UnixNano(),
			}

			errCh := bus.Emit(topicName, payload)
			<-errCh
		}

		bus.WaitForIdle()
		elapsed := time.Since(startTime)
		t.Logf("Soiree Retry Results (duration: %v):", elapsed)
		metrics.report(t)
	})
}

// newBenchGala creates a gala instance for benchmarking (in-memory, no River).
func newBenchGala(tb testing.TB, dispatcher Dispatcher) *Gala {
	tb.Helper()

	g := &Gala{}
	if err := g.initialize(dispatcher); err != nil {
		tb.Fatalf("failed to build gala runtime: %v", err)
	}

	return g
}

// registerBenchTopic registers a topic for benchmarking.
func registerBenchTopic(tb testing.TB, runtime *Gala, topic Topic[benchPayload]) {
	tb.Helper()

	if err := RegisterTopic(runtime.Registry(), Registration[benchPayload]{
		Topic: topic,
		Codec: JSONCodec[benchPayload]{},
	}); err != nil {
		tb.Fatalf("failed to register topic: %v", err)
	}
}

func (id EventID) String() string {
	return string(id)
}
