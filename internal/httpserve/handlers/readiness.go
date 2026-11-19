package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/v2/internal/entdb"
)

// readyCheckTimeout bounds each readiness check
const readyCheckTimeout = 3 * time.Second

// readyCheckOK is the status of a passing readiness check
const readyCheckOK = "OK"

// StatusResponse returns server status
type StatusResponse struct {
	Status map[string]string `json:"status"`
}

// CheckFunc is a function that can be used to check the status of a service
type CheckFunc func(ctx context.Context) error

type Checks struct {
	checks map[string]CheckFunc
}

// AddReadinessCheck will accept a function to be ran during calls to /ready
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

// ReadyHandler runs the readiness checks concurrently and reports their status
func (c *Checks) ReadyHandler(ctx echo.Context) error {
	if entdb.IsShuttingDown() {
		return ctx.JSON(http.StatusServiceUnavailable, echo.Map{"status": "shutting down"})
	}

	results := lop.Map(lo.Entries(c.checks), func(check lo.Entry[string, CheckFunc], _ int) lo.Entry[string, error] {
		checkCtx, cancel := context.WithTimeout(ctx.Request().Context(), readyCheckTimeout)
		defer cancel()

		return lo.Entry[string, error]{Key: check.Key, Value: check.Value(checkCtx)}
	})

	status := lo.Associate(results, func(result lo.Entry[string, error]) (string, string) {
		if result.Value != nil {
			return result.Key, result.Value.Error()
		}

		return result.Key, readyCheckOK
	})

	if lo.SomeBy(results, func(result lo.Entry[string, error]) bool { return result.Value != nil }) {
		return ctx.JSON(http.StatusServiceUnavailable, status)
	}

	return ctx.JSON(http.StatusOK, &StatusResponse{Status: status})
}
