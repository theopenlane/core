package transaction

import (
	"context"
	"database/sql"
	"errors"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/logx"
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
}

var entClientContextKey = contextx.NewKey[*ent.Tx]()

// FromContext returns a TX Client stored inside a context, or nil if there isn't one
func FromContext(ctx context.Context) *ent.Tx {
	c, _ := entClientContextKey.Get(ctx)
	return c
}

// NewContext returns a new context with the given TX Client attached
func NewContext(parent context.Context, c *ent.Tx) context.Context {
	return entClientContextKey.Set(parent, c)
}

// Middleware returns a middleware function for transactions on REST endpoints
func (d *Client) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqCtx := c.Request().Context()
		client, err := d.EntDBClient.Tx(reqCtx)
		if err != nil {
			logx.ErrorEvent(reqCtx, err).Msg(transactionStartErr)

			return ErrProcessingRequest
		}

		// add to context
		ctx := NewContext(reqCtx, client)

		c.SetRequest(c.Request().WithContext(ctx))

		if err := next(c); err != nil {
			logx.FromContext(ctx).
				Info().
				Err(err).
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Msg("rolling back transaction in middleware")

			if rbErr := client.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				logx.ErrorEvent(ctx, rbErr).Msg(rollbackErr)
			}

			return err
		}

		logx.FromContext(ctx).Debug().Msg("committing transaction in middleware")

		if err := client.Commit(); err != nil {
			logx.ErrorEvent(ctx, err).Msg(transactionCommitErr)

			return ErrProcessingRequest
		}

		return nil
	}
}
