package handlers

import (
	"context"
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/ent/entdb"
)

// StatusReply returns server status
type StatusReply struct {
	Status map[string]string `json:"status"`
}

// CheckFunc is a function that can be used to check the status of a service
type CheckFunc func(ctx context.Context) error

type Checks struct {
	checks map[string]CheckFunc
}

// AddReadinessCheck will accept a function to be ran during calls to /readyz
// These functions should accept a context and only return an error. When adding
// a readiness check a name is also provided, this name will be used when returning
// the state of all the checks
func (h *Handler) AddReadinessCheck(name string, f CheckFunc) {
	// if this is null, create the struct before trying to add
	if h.ReadyChecks.checks == nil {
		h.ReadyChecks.checks = map[string]CheckFunc{}
	}

	h.ReadyChecks.checks[name] = f
}

func (c *Checks) ReadyHandler(ctx echo.Context) error {
	if entdb.IsShuttingDown() {
		return ctx.JSON(http.StatusServiceUnavailable, echo.Map{"status": "shutting down"})
	}

	failed := false
	status := map[string]string{}

	for name, check := range c.checks {
		if err := check(ctx.Request().Context()); err != nil {
			failed = true
			status[name] = err.Error()
		} else {
			status[name] = "OK"
		}
	}

	if failed {
		return ctx.JSON(http.StatusServiceUnavailable, status)
	}

	out := &StatusReply{
		Status: status,
	}

	return ctx.JSON(http.StatusOK, out)
}
