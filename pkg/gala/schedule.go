package gala

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/theopenlane/core/pkg/logx"
)

const (
	// DefaultMinInterval is the shortest allowed scheduling interval
	DefaultMinInterval = 20 * time.Minute
	// DefaultMaxInterval is the longest allowed scheduling interval
	DefaultMaxInterval = 24 * time.Hour
	// DefaultBackoffFactor is the multiplier applied on idle or error ticks
	DefaultBackoffFactor = 2.0
	// DefaultHighDriftThreshold is the delta above which the interval snaps to minimum
	DefaultHighDriftThreshold = 200
	// intervalHalving is the divisor used to halve the interval on positive drift
	intervalHalving = 2

	// FullFetchMinInterval is the minimum interval for operations that always fetch all records
	FullFetchMinInterval = time.Hour
	// FullHighDriftThreshold is the delta above which the interval snaps to minimum
	FullHighDriftThreshold = 1000
)

// ScheduleOption configures a Schedule
type ScheduleOption func(*Schedule)

// WithMinInterval sets the shortest allowed interval between runs
func WithMinInterval(d time.Duration) ScheduleOption {
	return func(s *Schedule) { s.MinInterval = d }
}

// WithMaxInterval sets the longest allowed interval between runs
func WithMaxInterval(d time.Duration) ScheduleOption {
	return func(s *Schedule) { s.MaxInterval = d }
}

// WithBackoffFactor sets the multiplier applied when backing off
func WithBackoffFactor(f float64) ScheduleOption {
	return func(s *Schedule) { s.BackoffFactor = f }
}

// WithHighDriftThreshold sets the delta count above which the interval resets to minimum
func WithHighDriftThreshold(n int) ScheduleOption {
	return func(s *Schedule) { s.HighDriftThreshold = n }
}

// Schedule defines the adaptive scheduling policy for recurring work
type Schedule struct {
	// MinInterval is the shortest allowed interval between runs
	MinInterval time.Duration `json:"min_interval"`
	// MaxInterval is the longest allowed interval between runs
	MaxInterval time.Duration `json:"max_interval"`
	// BackoffFactor is the multiplier applied when backing off (idle or error)
	BackoffFactor float64 `json:"backoff_factor"`
	// HighDriftThreshold is the delta count above which the interval resets to MinInterval
	HighDriftThreshold int `json:"high_drift_threshold"`
}

// ScheduleSpec declares the adaptive re-emit loop for a scheduled listener definition
type ScheduleSpec[T any] struct {
	// Schedule controls adaptive interval computation
	Schedule Schedule
	// Handle is the handler invoked each cycle, returning the delta for scheduling
	Handle func(context.Context, T) (int, error)
	// State extracts the ScheduleState from the envelope
	State func(T) ScheduleState
	// Wrap builds a new envelope carrying the updated ScheduleState
	Wrap func(T, ScheduleState) T
	// PrepareEmit optionally enriches the context and headers before re-emitting
	PrepareEmit func(context.Context, T) (context.Context, Headers)
	// Override optionally returns a per-envelope schedule that overrides Schedule;
	// returning nil falls back to Schedule
	Override func(T) *Schedule
}

// scheduleHandler builds the self-sustaining loop handler for a scheduled definition: it
// processes one cycle, computes the next adaptive interval, and re-emits the successor
func scheduleHandler[T any](g *Gala, definition Definition[T]) Handler[T] {
	spec := definition.Schedule

	return func(ctx HandlerContext, payload T) error {
		delta, execErr := spec.Handle(ctx.Context, payload)

		if execErr != nil {
			if definition.Cancel != nil && definition.Cancel(ctx.Context, payload, execErr) {
				return river.JobCancel(execErr)
			}

			state := spec.State(payload)
			logx.FromContext(ctx.Context).Warn().Err(execErr).Int("error_streak", state.ErrorStreak+1).Msg("scheduled listener cycle failed, scheduling retry with backoff")
		}

		effectiveSchedule := spec.Schedule
		if spec.Override != nil {
			if override := spec.Override(payload); override != nil {
				effectiveSchedule = *override
			}
		}

		next := effectiveSchedule.Next(spec.State(payload), delta, execErr)
		scheduledAt := next.NextScheduledAt()

		emitCtx := ctx.Context
		headers := Headers{}

		if spec.PrepareEmit != nil {
			emitCtx, headers = spec.PrepareEmit(ctx.Context, payload)
		}

		headers.ScheduledAt = &scheduledAt
		// the successor of a running unique job would be skipped as its own duplicate
		headers.SkipUniqueKey = true

		_, emitErr := g.Emit(emitCtx, definition.Topic.Name, spec.Wrap(payload, next), WithHeaders(headers))

		if execErr != nil {
			if emitErr != nil {
				logx.FromContext(ctx.Context).Error().Err(emitErr).Msg("scheduled listener re-emit failed, loop will not continue")
			}

			return river.JobCancel(execErr)
		}

		return emitErr
	}
}

// ScheduleState carries adaptive scheduling state across dispatch cycles
type ScheduleState struct {
	// Interval is the current scheduling interval
	Interval time.Duration `json:"interval"`
	// IdleStreak is the number of consecutive runs with zero delta
	IdleStreak int `json:"idle_streak"`
	// ErrorStreak is the number of consecutive runs that returned an error
	ErrorStreak int `json:"error_streak"`
}

// NewFullFetchSchedule creates a Schedule suited for operations that always fetch all records
// and cannot do incremental syncs, using FullFetchMinInterval as the minimum
func NewFullFetchSchedule(opts ...ScheduleOption) *Schedule {
	s := NewSchedule(append(
		[]ScheduleOption{
			WithMinInterval(FullFetchMinInterval),
			WithHighDriftThreshold(FullHighDriftThreshold),
		}, opts...)...)
	return &s
}

// NewSchedule creates a Schedule with defaults and applies any provided options
func NewSchedule(opts ...ScheduleOption) Schedule {
	s := Schedule{
		MinInterval:        DefaultMinInterval,
		MaxInterval:        DefaultMaxInterval,
		BackoffFactor:      DefaultBackoffFactor,
		HighDriftThreshold: DefaultHighDriftThreshold,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// Next computes the next scheduling state from the current state and run outcome.
// A non-nil error signals a failed run; delta is the number of records that changed
func (s Schedule) Next(state ScheduleState, delta int, err error) ScheduleState {
	s = s.withDefaults()

	interval := max(state.Interval, s.MinInterval)

	switch {
	case err != nil:
		return ScheduleState{
			Interval:    s.clamp(time.Duration(float64(interval) * s.BackoffFactor)),
			ErrorStreak: state.ErrorStreak + 1,
		}
	case delta >= s.HighDriftThreshold:
		return ScheduleState{
			Interval: s.MinInterval,
		}
	case delta > 0:
		return ScheduleState{
			Interval: max(interval/intervalHalving, s.MinInterval),
		}
	default:
		return ScheduleState{
			Interval:   s.clamp(time.Duration(float64(interval) * s.BackoffFactor)),
			IdleStreak: state.IdleStreak + 1,
		}
	}
}

// NextScheduledAt returns the wall-clock time for the next run based on the computed state
func (s ScheduleState) NextScheduledAt() time.Time {
	return time.Now().Add(s.Interval)
}

// clamp restricts an interval to [MinInterval, MaxInterval]
func (s Schedule) clamp(d time.Duration) time.Duration {
	switch {
	case d < s.MinInterval:
		return s.MinInterval
	case d > s.MaxInterval:
		return s.MaxInterval
	default:
		return d
	}
}

// withDefaults fills zero-valued fields with package defaults
func (s Schedule) withDefaults() Schedule {
	if s.MinInterval <= 0 {
		s.MinInterval = DefaultMinInterval
	}

	if s.MaxInterval <= 0 {
		s.MaxInterval = DefaultMaxInterval
	}

	if s.BackoffFactor <= 0 {
		s.BackoffFactor = DefaultBackoffFactor
	}

	if s.HighDriftThreshold <= 0 {
		s.HighDriftThreshold = DefaultHighDriftThreshold
	}

	return s
}
