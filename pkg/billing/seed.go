package billing

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
)

type PlansWrapper struct {
	Plans    []entitlements.Plan
	EntPlans []*ent.EntitlementPlan
}

// Seed creates the seed data for the billing service
func Seed(entclient *ent.Client, stripe *entitlements.StripeClient) error {
	products := stripe.GetProducts()
	planswrapper := PlansWrapper{}

	ctx := ent.NewContext(context.Background(), entclient)
	orgGetCtx := privacy.DecisionContext(ctx, privacy.Allow)

	for _, product := range products {
		for _, price := range product.Prices {
			entclient.EntitlementPlan.Create().
				SetName(product.Name).
				SetDescription(product.Description).
				SetStripeProductID(product.ID).
				SetStripePriceID(price.ID).
				SetVersion(price.Interval).
				SaveX(orgGetCtx)
		}
	}

	seedFeatures(orgGetCtx, entclient, planswrapper.Plans)

	return nil
}

func seedFeatures(ctx context.Context, entclient *ent.Client, plans []entitlements.Plan) {
	for _, plan := range plans {
		for _, feature := range plan.Features {
			entclient.Feature.Create().
				SetName(feature.Name).
				SetStripeFeatureID(feature.ID).
				SetEnabled(true).
				SaveX(ctx)
		}
	}
}
