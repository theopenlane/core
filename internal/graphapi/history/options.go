package graphapihistory

import (
	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql/handler"
	ent "github.com/theopenlane/core/internal/ent/historygenerated"
)

// WithTransactions adds the transactioner to the ent db client
func WithTransactions(h *handler.Server, d *ent.Client) {
	// setup transactional db client
	h.AroundOperations(injectClient(d))

	h.Use(entgql.Transactioner{TxOpener: d})
}
