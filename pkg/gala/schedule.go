package gala

import "time"

const (
	// DefaultMinInterval is the shortest allowed scheduling interval
	DefaultMinInterval = 5 * time.Minute
	// DefaultMaxInterval is the longest allowed scheduling interval
	DefaultMaxInterval = 24 * time.Hour
	// DefaultBackoffFactor is the multiplier applied on idle or error ticks
	DefaultBackoffFactor = 2.0
	// DefaultHighDriftThreshold is the delta above which the interval snaps to minimum
	DefaultHighDriftThreshold = 100
	// intervalHalving is the divisor used to halve the interval on positive drift
	intervalHalving = 2
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

// ScheduleState carries adaptive scheduling state across dispatch cycles
type ScheduleState struct {
	// Interval is the current scheduling interval
	Interval time.Duration `json:"interval"`
	// IdleStreak is the number of consecutive runs with zero delta
	IdleStreak int `json:"idle_streak"`
	// ErrorStreak is the number of consecutive runs that returned an error
	ErrorStreak int `json:"error_streak"`
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
