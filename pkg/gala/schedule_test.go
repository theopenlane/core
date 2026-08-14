package gala

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/riverqueue/river"
)

func TestNewScheduleDefaults(t *testing.T) {
	s := NewSchedule()

	if s.MinInterval != DefaultMinInterval {
		t.Fatalf("expected MinInterval %v, got %v", DefaultMinInterval, s.MinInterval)
	}

	if s.MaxInterval != DefaultMaxInterval {
		t.Fatalf("expected MaxInterval %v, got %v", DefaultMaxInterval, s.MaxInterval)
	}

	if s.BackoffFactor != DefaultBackoffFactor {
		t.Fatalf("expected BackoffFactor %v, got %v", DefaultBackoffFactor, s.BackoffFactor)
	}

	if s.HighDriftThreshold != DefaultHighDriftThreshold {
		t.Fatalf("expected HighDriftThreshold %v, got %v", DefaultHighDriftThreshold, s.HighDriftThreshold)
	}
}

func TestNewScheduleWithOptions(t *testing.T) {
	s := NewSchedule(
		WithMinInterval(1*time.Minute),
		WithMaxInterval(10*time.Minute),
		WithBackoffFactor(3.0),
		WithHighDriftThreshold(50),
	)

	if s.MinInterval != 1*time.Minute {
		t.Fatalf("expected MinInterval 1m, got %v", s.MinInterval)
	}

	if s.MaxInterval != 10*time.Minute {
		t.Fatalf("expected MaxInterval 10m, got %v", s.MaxInterval)
	}

	if s.BackoffFactor != 3.0 {
		t.Fatalf("expected BackoffFactor 3.0, got %v", s.BackoffFactor)
	}

	if s.HighDriftThreshold != 50 {
		t.Fatalf("expected HighDriftThreshold 50, got %v", s.HighDriftThreshold)
	}
}

func TestScheduleNextZeroStateSetsMinInterval(t *testing.T) {
	s := NewSchedule()
	next := s.Next(ScheduleState{}, 0, nil)

	// zero interval floors to MinInterval, then idle backoff applies
	expected := time.Duration(float64(DefaultMinInterval) * DefaultBackoffFactor)
	if next.Interval != expected {
		t.Fatalf("expected %v, got %v", expected, next.Interval)
	}

	if next.IdleStreak != 1 {
		t.Fatalf("expected idle streak 1, got %d", next.IdleStreak)
	}
}

func TestScheduleNextHighDriftSnapsToMin(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 30 * time.Minute, IdleStreak: 5}

	next := s.Next(state, 150, nil)

	if next.Interval != DefaultMinInterval {
		t.Fatalf("expected %v, got %v", DefaultMinInterval, next.Interval)
	}

	if next.IdleStreak != 0 {
		t.Fatalf("expected idle streak reset to 0, got %d", next.IdleStreak)
	}

	if next.ErrorStreak != 0 {
		t.Fatalf("expected error streak 0, got %d", next.ErrorStreak)
	}
}

func TestScheduleNextLowDriftHalvesInterval(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 20 * time.Minute}

	next := s.Next(state, 20, nil)

	if next.Interval != 20*time.Minute {
		t.Fatalf("expected 20m, got %v", next.Interval)
	}
}

func TestScheduleNextLowDriftFloorsAtMin(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 6 * time.Minute}

	next := s.Next(state, 1, nil)

	if next.Interval != DefaultMinInterval {
		t.Fatalf("expected %v, got %v", DefaultMinInterval, next.Interval)
	}
}

func TestScheduleNextIdleBacksOff(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Incarnation: "chain-1", Interval: 10 * time.Minute, IdleStreak: 2}

	next := s.Next(state, 0, nil)

	if next.Interval != 40*time.Minute {
		t.Fatalf("expected 40m, got %v", next.Interval)
	}

	if next.IdleStreak != 3 {
		t.Fatalf("expected idle streak 3, got %d", next.IdleStreak)
	}

	if next.Incarnation != state.Incarnation {
		t.Fatalf("expected schedule incarnation %q, got %q", state.Incarnation, next.Incarnation)
	}
}

func TestScheduleNextIdleCapsAtMax(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 18 * time.Hour}

	next := s.Next(state, 0, nil)

	if next.Interval != DefaultMaxInterval {
		t.Fatalf("expected %v, got %v", DefaultMaxInterval, next.Interval)
	}
}

func TestScheduleNextErrorBacksOff(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 10 * time.Minute, ErrorStreak: 1}

	next := s.Next(state, 0, errors.New("upstream unavailable"))

	if next.Interval != 40*time.Minute {
		t.Fatalf("expected 40m, got %v", next.Interval)
	}

	if next.ErrorStreak != 2 {
		t.Fatalf("expected error streak 2, got %d", next.ErrorStreak)
	}

	if next.IdleStreak != 0 {
		t.Fatalf("expected idle streak reset to 0, got %d", next.IdleStreak)
	}
}

func TestScheduleNextErrorCapsAtMax(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 18 * time.Hour, ErrorStreak: 3}

	next := s.Next(state, 0, errors.New("still down"))

	if next.Interval != DefaultMaxInterval {
		t.Fatalf("expected %v, got %v", DefaultMaxInterval, next.Interval)
	}
}

func TestScheduleNextSuccessAfterErrorsResetsStreak(t *testing.T) {
	s := NewSchedule()
	state := ScheduleState{Interval: 30 * time.Minute, ErrorStreak: 5}

	next := s.Next(state, 200, nil)

	if next.Interval != DefaultMinInterval {
		t.Fatalf("expected %v, got %v", DefaultMinInterval, next.Interval)
	}

	if next.ErrorStreak != 0 {
		t.Fatalf("expected error streak reset to 0, got %d", next.ErrorStreak)
	}
}

func TestScheduleNextCustomConfig(t *testing.T) {
	s := NewSchedule(
		WithMinInterval(1*time.Minute),
		WithMaxInterval(10*time.Minute),
		WithBackoffFactor(3.0),
		WithHighDriftThreshold(50),
	)

	state := ScheduleState{Interval: 2 * time.Minute}

	// idle: 2m * 3 = 6m
	next := s.Next(state, 0, nil)
	if next.Interval != 6*time.Minute {
		t.Fatalf("expected 6m, got %v", next.Interval)
	}

	// high drift at custom threshold
	next = s.Next(state, 50, nil)
	if next.Interval != 1*time.Minute {
		t.Fatalf("expected 1m, got %v", next.Interval)
	}
}

func TestScheduleStateNextScheduledAt(t *testing.T) {
	state := ScheduleState{Interval: 15 * time.Minute}

	before := time.Now().Add(15 * time.Minute)
	scheduled := state.NextScheduledAt()
	after := time.Now().Add(15 * time.Minute)

	if scheduled.Before(before) || scheduled.After(after) {
		t.Fatalf("NextScheduledAt %v not in expected range [%v, %v]", scheduled, before, after)
	}
}

func TestScheduleWithDefaultsFillsZeroValues(t *testing.T) {
	s := Schedule{}
	filled := s.withDefaults()

	if filled.MinInterval != DefaultMinInterval {
		t.Fatalf("expected MinInterval %v, got %v", DefaultMinInterval, filled.MinInterval)
	}

	if filled.MaxInterval != DefaultMaxInterval {
		t.Fatalf("expected MaxInterval %v, got %v", DefaultMaxInterval, filled.MaxInterval)
	}

	if filled.BackoffFactor != DefaultBackoffFactor {
		t.Fatalf("expected BackoffFactor %v, got %v", DefaultBackoffFactor, filled.BackoffFactor)
	}

	if filled.HighDriftThreshold != DefaultHighDriftThreshold {
		t.Fatalf("expected HighDriftThreshold %v, got %v", DefaultHighDriftThreshold, filled.HighDriftThreshold)
	}
}

func TestScheduleHandlerMarksSuccessorUniqueOnce(t *testing.T) {
	dispatcher := &runtimeTestDispatcher{}
	runtime := newTestGala(t, dispatcher)
	topic := Topic[runtimeTestPayload]{
		Name:      TopicName("gala.test.schedule.unique_once"),
		Kind:      JobKindSystem,
		UniqueKey: func(payload runtimeTestPayload) string { return payload.Message },
	}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	var successorState ScheduleState
	definition := Definition[runtimeTestPayload]{
		Topic: topic,
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Schedule: NewSchedule(WithMinInterval(time.Millisecond)),
			Handle:   func(context.Context, runtimeTestPayload) (int, error) { return 0, nil },
			State:    func(runtimeTestPayload) ScheduleState { return ScheduleState{} },
			Wrap: func(payload runtimeTestPayload, state ScheduleState) runtimeTestPayload {
				successorState = state

				return payload
			},
		},
	}

	handler := scheduleHandler(runtime, definition)
	handlerCtx := HandlerContext{
		Context:  context.Background(),
		Envelope: Envelope{ID: EventID("schedule-incarnation")},
	}
	err := handler(handlerCtx, runtimeTestPayload{Message: "loop"})
	if err != nil {
		t.Fatalf("unexpected schedule error: %v", err)
	}
	if err := handler(handlerCtx, runtimeTestPayload{Message: "loop"}); err != nil {
		t.Fatalf("unexpected schedule retry error: %v", err)
	}

	if len(dispatcher.envelopes) != 2 {
		t.Fatalf("expected two successor dispatch attempts, got %d", len(dispatcher.envelopes))
	}
	if !dispatcher.envelopes[0].Headers.UniqueOnce {
		t.Fatal("expected successor unique key to include terminal states")
	}
	if successorState.Incarnation != "schedule-incarnation" {
		t.Fatalf("expected envelope id to seed schedule incarnation, got %q", successorState.Incarnation)
	}
	wantKey := "loop:incarnation:schedule-incarnation:cycle:1"
	if got := dispatcher.envelopes[0].Headers.UniqueKey; got != wantKey {
		t.Fatalf("successor unique key = %q, want %q", got, wantKey)
	}
	if dispatcher.envelopes[1].Headers.UniqueKey != dispatcher.envelopes[0].Headers.UniqueKey {
		t.Fatalf("retry derived a different successor key: %q != %q", dispatcher.envelopes[1].Headers.UniqueKey, dispatcher.envelopes[0].Headers.UniqueKey)
	}
}

func TestScheduleHandlerReturnsEmitErrorWhenExecutionAndEmitFail(t *testing.T) {
	emitErr := errors.New("successor insert failed")
	dispatcher := &runtimeTestDispatcher{err: emitErr}
	runtime := newTestGala(t, dispatcher)
	topic := Topic[runtimeTestPayload]{
		Name:      TopicName("gala.test.schedule.emit_error"),
		Kind:      JobKindSystem,
		UniqueKey: func(payload runtimeTestPayload) string { return payload.Message },
	}
	if err := registerTopic(runtime.registry, topic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	execErr := errors.New("cycle failed")
	definition := Definition[runtimeTestPayload]{
		Topic: topic,
		Schedule: &ScheduleSpec[runtimeTestPayload]{
			Schedule: NewSchedule(WithMinInterval(time.Millisecond)),
			Handle:   func(context.Context, runtimeTestPayload) (int, error) { return 0, execErr },
			State:    func(runtimeTestPayload) ScheduleState { return ScheduleState{} },
			Wrap:     func(payload runtimeTestPayload, _ ScheduleState) runtimeTestPayload { return payload },
		},
	}

	err := scheduleHandler(runtime, definition)(HandlerContext{Context: context.Background()}, runtimeTestPayload{Message: "loop"})
	if !errors.Is(err, emitErr) {
		t.Fatalf("expected emit error for River retry, got %v", err)
	}
	if !errors.Is(err, execErr) {
		t.Fatalf("expected joined execution error, got %v", err)
	}
	var cancelErr *river.JobCancelError
	if errors.As(err, &cancelErr) {
		t.Fatalf("expected emit failure not to cancel the predecessor, got %v", err)
	}
}

func TestScheduleStateIncarnationJSONCompatibility(t *testing.T) {
	const legacy = `{"interval":60000000000,"idle_streak":2,"error_streak":1,"cycle":7}`

	var state ScheduleState
	if err := json.Unmarshal([]byte(legacy), &state); err != nil {
		t.Fatalf("legacy schedule state failed to decode: %v", err)
	}
	if state.Incarnation != "" || state.Cycle != 7 {
		t.Fatalf("unexpected legacy schedule state: %#v", state)
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("schedule state failed to encode: %v", err)
	}
	if strings.Contains(string(encoded), "incarnation") {
		t.Fatalf("zero incarnation should remain omitted, got %s", encoded)
	}

	state.Incarnation = "chain-1"
	encoded, err = json.Marshal(state)
	if err != nil {
		t.Fatalf("schedule state with incarnation failed to encode: %v", err)
	}
	if !strings.Contains(string(encoded), `"incarnation":"chain-1"`) {
		t.Fatalf("incarnation missing from encoded state: %s", encoded)
	}
}
