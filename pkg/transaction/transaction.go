package transaction

import (
	"context"
	"errors"
	"net/http"

	echo "github.com/theopenlane/echox"
	"go.uber.org/zap"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

const (
	rollbackErr          = "error rolling back transaction"
	transactionStartErr  = "error starting transaction"
	transactionCommitErr = "error committing transaction"
)

var (
	// ErrProcessingRequest is returned when the request cannot be processed
	ErrProcessingRequest = errors.New("error processing request, please try again")
)

type Client struct {
	EntDBClient *ent.Client
	Logger      *zap.SugaredLogger
}

type entClientCtxKey struct{}

// FromContext returns a TX Client stored inside a context, or nil if there isn't one
func FromContext(ctx context.Context) *ent.Tx {
	c, _ := ctx.Value(entClientCtxKey{}).(*ent.Tx)
	return c
}

// NewContext returns a new context with the given TX Client attached
func NewContext(parent context.Context, c *ent.Tx) context.Context {
	return context.WithValue(parent, entClientCtxKey{}, c)
}

// Middleware returns a middleware function for transactions on REST endpoints
func (d *Client) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		client, err := d.EntDBClient.Tx(c.Request().Context())
		if err != nil {
			d.Logger.Errorw(transactionStartErr, "error", err)

			return c.JSON(http.StatusInternalServerError, ErrProcessingRequest)
		}

		// add to context
		ctx := NewContext(c.Request().Context(), client)

		c.SetRequest(c.Request().WithContext(ctx))

		if err := next(c); err != nil {
			d.Logger.Debug("rolling back transaction in middleware")

			if err := client.Rollback(); err != nil {
				d.Logger.Errorw(rollbackErr, "error", err)

				return c.JSON(http.StatusInternalServerError, ErrProcessingRequest)
			}

			return err
		}

		d.Logger.Debug("committing transaction in middleware")

		if err := client.Commit(); err != nil {
			d.Logger.Errorw(transactionCommitErr, "error", err)

			return c.JSON(http.StatusInternalServerError, ErrProcessingRequest)
		}

		return nil
	}
}
