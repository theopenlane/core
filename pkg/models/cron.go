package models

import (
	"database/sql/driver"
	"fmt"
	"io"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	// MaxRunsInBetween defines how much time each job must have between runs
	// Maybe make this configurable or maybe we need to take this down to like
	// 5/10 minutes
	MaxRunsInBetween = 30 * time.Minute

	nextNCronExecutions = 5
)

// Cron defines the syntax for the job execution
type Cron string

// Validate checks a cron to make sure it is valid .
// It also limits concurrent runs to 30 minutes interval of the last run
// so it parses the cron - look at next few executions and check the elapsed time
func (c Cron) Validate() error {
	if c.String() == "" {
		return nil
	}

	cron, err := cronexpr.Parse(c.String())
	if err != nil {
		return fmt.Errorf("invalid cron syntax: %w", err) // nolint:err113
	}

	// compute the next 5 execution times to cover cases like
	// 0,20,40 * * * * where the user can request to run in the 20th and 40th minute
	// that would break the 30 minute check
	currentTime := time.Now()
	executions := cron.NextN(currentTime, nextNCronExecutions)

	for i := 1; i < len(executions); i++ {
		interval := executions[i].Sub(executions[i-1])
		if interval < MaxRunsInBetween {
			return fmt.Errorf("cron runs too frequently: %s between runs, must be at least 30 minutes", interval) // nolint:err113
		}
	}

	return nil
}

// Next returns the next scheduled time after `from` based on the cron expression.
func (c Cron) Next(from time.Time) time.Time {
	cron, err := cronexpr.Parse(c.String())
	if err != nil {
		return time.Time{}
	}

	next := cron.Next(from)
	if next.IsZero() {
		return time.Time{}
	}

	return next
}

// String returns a string representation of the cron
func (c Cron) String() string { return string(c) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (c Cron) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, fmt.Sprintf("%q", c.String()))
}

// Value returns human readable cron from the database
func (c Cron) Value() (driver.Value, error) { return string(c), nil }

func (c *Cron) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported Scan type for Cron: %T", value) //nolint:err113
	}

	*c = Cron(str)

	return nil
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (c *Cron) UnmarshalGQL(v any) error { return unmarshalGQLJSON(v, c) }
