package gala

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/riverqueue/river"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// defaultMinInterval is the shortest allowed scheduling interval
	defaultMinInterval = 20 * time.Minute
	// defaultMaxInterval is the longest allowed scheduling interval
	defaultMaxInterval = 24 * time.Hour
	// defaultBackoffFactor is the multiplier applied on idle or error ticks
	defaultBackoffFactor = 2.0
	// defaultHighDriftThreshold is the delta above which the interval snaps to minimum
	defaultHighDriftThreshold = 200
	// intervalHalving is the divisor used to halve the interval on positive drift
	intervalHalving = 2
	// defaultMaxErrorStreak stops a loop after this many consecutive failed cycles
	defaultMaxErrorStreak = 5
	// fullFetchMinInterval is the minimum interval for operations that always fetch all records
	fullFetchMinInterval = time.Hour
	// FullHighDriftThreshold is the delta above which a full-fetch schedule snaps to minimum
	FullHighDriftThreshold = 1000
	// UnlimitedErrorStreak disables error-streak exhaustion
	UnlimitedErrorStreak = -1
)

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
	// MaxErrorStreak stops the loop after this many consecutive failed cycles; UnlimitedErrorStreak disables exhaustion
	MaxErrorStreak int `json:"max_error_streak"`
}

// ScheduleSpec declares the adaptive re-emit loop for a scheduled listener definition
type ScheduleSpec[T any] struct {
	// Schedule controls adaptive interval computation
	Schedule Schedule
	// Handle is the handler invoked each cycle
	Handle func(context.Context, T) (int, error)
	// State extracts the ScheduleState from the envelope
	State func(T) ScheduleState
	// Wrap builds a new envelope carrying the updated ScheduleState
	Wrap func(T, ScheduleState) T
	// PrepareEmit optionally enriches the context and headers before re-emitting
	PrepareEmit func(context.Context, T) (context.Context, Headers)
	// Override optionally returns a per-envelope schedule that overrides Schedule
	Override func(T) *Schedule
}

// scheduleHandler builds the self-sustaining loop handler for a scheduled definition
func scheduleHandler[T any](g *Gala, definition Definition[T]) Handler[T] {
	spec := definition.Schedule

	return func(ctx HandlerContext, payload T) error {
		delta, execErr := spec.Handle(ctx.Context, payload)
		state := spec.State(payload)

		effectiveSchedule := spec.Schedule
		if spec.Override != nil {
			if override := spec.Override(payload); override != nil {
				effectiveSchedule = *override
			}
		}

		if execErr != nil {
			if definition.Cancel != nil && definition.Cancel(ctx.Context, payload, execErr) {
				logx.FromContext(ctx.Context).Error().Err(execErr).Msg("scheduled listener cycle failed, canceling loop")

				return river.JobCancel(execErr)
			}

			streak := state.ErrorStreak + 1
			if effectiveSchedule.exhausted(streak) {
				if definition.OnExhausted != nil {
					definition.OnExhausted(ctx.Context, payload, execErr)
				}

				logx.FromContext(ctx.Context).Error().Err(execErr).Int("error_streak", streak).Msg("scheduled listener exhausted error budget, stopping loop")

				return river.JobCancel(execErr)
			}

			logx.FromContext(ctx.Context).Warn().Err(execErr).Int("error_streak", streak).Msg("scheduled listener cycle failed, scheduling retry with backoff")
		}

		next := effectiveSchedule.Next(state, delta, execErr)
		next.Cycle = state.Cycle + 1
		next.Incarnation = state.Incarnation
		if next.Incarnation == "" {
			next.Incarnation = string(ctx.Envelope.ID)
			if next.Incarnation == "" {
				next.Incarnation = string(NewEventID())
			}
		}

		scheduledAt := next.NextScheduledAt()
		wrapped := spec.Wrap(payload, next)

		emitCtx := ctx.Context
		headers := Headers{}

		if spec.PrepareEmit != nil {
			emitCtx, headers = spec.PrepareEmit(ctx.Context, payload)
		}

		// Prevent a detached running cycle from emitting its successor
		if err := ctx.Context.Err(); err != nil {
			return river.JobCancel(err)
		}

		headers.ScheduledAt = &scheduledAt

		// a per-cycle key dedups crash-retry re-emissions without colliding with the running predecessor
		if definition.Topic.UniqueKey != nil {
			headers.UniqueKey = strings.Join([]string{definition.Topic.UniqueKey(wrapped), "incarnation", next.Incarnation, "cycle", strconv.Itoa(next.Cycle)}, uniqueKeySeparator)
			headers.UniqueOnce = true
		} else {
			headers.SkipUniqueKey = true
		}

		_, emitErr := g.EmitWithHeaders(emitCtx, definition.Topic.Name, wrapped, headers)
		if emitErr != nil {
			logx.FromContext(ctx.Context).Error().Err(emitErr).Msg("scheduled listener re-emit failed, predecessor will retry")

			return errors.Join(execErr, emitErr)
		}

		if execErr != nil {
			return river.JobCancel(execErr)
		}

		return nil
	}
}

// ScheduleState carries adaptive scheduling state across dispatch cycles
type ScheduleState struct {
	// Incarnation identifies one logical schedule chain across every cycle
	Incarnation string `json:"incarnation,omitempty"`
	// Interval is the current scheduling interval
	Interval time.Duration `json:"interval"`
	// IdleStreak is the number of consecutive runs with zero delta
	IdleStreak int `json:"idle_streak"`
	// ErrorStreak is the number of consecutive runs that returned an error
	ErrorStreak int `json:"error_streak"`
	// Cycle is the monotonic cycle counter
	Cycle int `json:"cycle"`
}

// NewFullFetchSchedule creates a Schedule suited for operations that always fetch all records
// and cannot do incremental syncs, using an hour as the minimum interval
func NewFullFetchSchedule() *Schedule {
	return &Schedule{
		MinInterval:        fullFetchMinInterval,
		HighDriftThreshold: FullHighDriftThreshold,
	}
}

// Next computes the next scheduling state from the current state and run outcome.
// A non-nil error signals a failed run; delta is the number of records that changed
func (s Schedule) Next(state ScheduleState, delta int, err error) ScheduleState {
	s = s.withDefaults()

	interval := max(state.Interval, s.MinInterval)

	switch {
	case err != nil:
		return ScheduleState{
			Incarnation: state.Incarnation,
			Interval:    s.clamp(time.Duration(float64(interval) * s.BackoffFactor)),
			ErrorStreak: state.ErrorStreak + 1,
		}
	case delta >= s.HighDriftThreshold:
		return ScheduleState{
			Incarnation: state.Incarnation,
			Interval:    s.MinInterval,
		}
	case delta > 0:
		return ScheduleState{
			Incarnation: state.Incarnation,
			Interval:    max(interval/intervalHalving, s.MinInterval),
		}
	default:
		return ScheduleState{
			Incarnation: state.Incarnation,
			Interval:    s.clamp(time.Duration(float64(interval) * s.BackoffFactor)),
			IdleStreak:  state.IdleStreak + 1,
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
		s.MinInterval = defaultMinInterval
	}

	if s.MaxInterval <= 0 {
		s.MaxInterval = defaultMaxInterval
	}

	if s.BackoffFactor <= 0 {
		s.BackoffFactor = defaultBackoffFactor
	}

	if s.HighDriftThreshold <= 0 {
		s.HighDriftThreshold = defaultHighDriftThreshold
	}

	if s.MaxErrorStreak == 0 {
		s.MaxErrorStreak = defaultMaxErrorStreak
	}

	return s
}

// exhausted reports whether a consecutive-error streak has reached the stop threshold
func (s Schedule) exhausted(streak int) bool {
	limit := s.withDefaults().MaxErrorStreak

	return limit > 0 && streak >= limit
}
