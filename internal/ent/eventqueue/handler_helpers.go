package eventqueue

import (
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// ClientFromHandler resolves the ent client from Gala listener dependencies
func ClientFromHandler(ctx gala.HandlerContext) (*generated.Client, bool) {
	client, err := do.Invoke[*generated.Client](ctx.Injector)
	if err != nil || client == nil {
		return nil, false
	}

	return client, true
}
