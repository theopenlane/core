package gala

import (
	"errors"
	"testing"
	"time"
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
	state := ScheduleState{Interval: 10 * time.Minute, IdleStreak: 2}

	next := s.Next(state, 0, nil)

	if next.Interval != 40*time.Minute {
		t.Fatalf("expected 40m, got %v", next.Interval)
	}

	if next.IdleStreak != 3 {
		t.Fatalf("expected idle streak 3, got %d", next.IdleStreak)
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
