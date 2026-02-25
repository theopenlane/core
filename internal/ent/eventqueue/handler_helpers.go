package eventqueue

import (
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// ClientFromHandler resolves the ent client from the Gala injector and seeds it into
// HandlerContext.Context so that ent interceptors relying on generated.FromContext work
// correctly when the context was reconstructed from a durable snapshot.
// The restored snapshot carries auth claims and log fields but not the ent client,
// which is a live runtime dependency. Without it, interceptors such as InterceptorModules
// cannot resolve the FGA authz client and incorrectly report features as disabled.
func ClientFromHandler(ctx gala.HandlerContext) (gala.HandlerContext, *generated.Client, bool) {
	client, err := do.Invoke[*generated.Client](ctx.Injector)
	if err != nil || client == nil {
		return ctx, nil, false
	}

	ctx.Context = generated.NewContext(ctx.Context, client)

	return ctx, client, true
}
