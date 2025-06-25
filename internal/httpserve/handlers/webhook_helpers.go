package handlers

import (
	"context"

	"github.com/stripe/stripe-go/v82"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgprice"
	"github.com/theopenlane/core/internal/ent/generated/orgproduct"
	"github.com/theopenlane/core/pkg/middleware/transaction"

	em "github.com/theopenlane/core/internal/entmapping"
)

// syncSubscriptionItemsWithStripe ensures OrgProduct, OrgPrice, and OrgModule
// records exist and are updated based on the given Stripe subscription data.
func syncSubscriptionItemsWithStripe(ctx context.Context, sub *stripe.Subscription) error {
	orgSub, err := getOrgSubscription(ctx, sub)
	if err != nil {
		return err
	}

	for _, item := range sub.Items.Data {
		if item.Price == nil || item.Price.Product == nil {
			continue
		}

		prod, err := upsertOrgProduct(ctx, orgSub, item.Price.Product)
		if err != nil {
			return err
		}

		price, err := upsertOrgPrice(ctx, orgSub, prod, item.Price)
		if err != nil {
			return err
		}

		if _, err = upsertOrgModule(ctx, orgSub, prod, price, item); err != nil {
			return err
		}
	}

	return nil
}

func upsertOrgProduct(ctx context.Context, orgSub *ent.OrgSubscription, p *stripe.Product) (*ent.OrgProduct, error) {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	existing, err := tx.OrgProduct.Query().
		Where(orgproduct.StripeProductID(p.ID), orgproduct.SubscriptionID(orgSub.ID)).
		Only(allowCtx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if ent.IsNotFound(err) {
		builder := tx.OrgProduct.Create().
			SetOwnerID(orgSub.OwnerID).
			SetSubscriptionID(orgSub.ID)
		em.ApplyStripeProduct(builder, p)
		return builder.Save(allowCtx)
	}

	builder := tx.OrgProduct.UpdateOne(existing)
	em.ApplyStripeProduct(builder, p)
	_, err = builder.Save(allowCtx)
	return existing, err
}

func upsertOrgPrice(ctx context.Context, orgSub *ent.OrgSubscription, prod *ent.OrgProduct, price *stripe.Price) (*ent.OrgPrice, error) {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	existing, err := tx.OrgPrice.Query().
		Where(orgprice.StripePriceID(price.ID), orgprice.SubscriptionID(orgSub.ID)).
		Only(allowCtx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if ent.IsNotFound(err) {
		builder := tx.OrgPrice.Create().
			SetOwnerID(orgSub.OwnerID).
			SetSubscriptionID(orgSub.ID).
			SetProductID(prod.ID)
		em.ApplyStripePrice(builder, price)
		return builder.Save(allowCtx)
	}

	builder := tx.OrgPrice.UpdateOne(existing)
	em.ApplyStripePrice(builder, price)
	builder.SetProductID(prod.ID)
	_, err = builder.Save(allowCtx)
	return existing, err
}

func upsertOrgModule(ctx context.Context, orgSub *ent.OrgSubscription, _ *ent.OrgProduct, price *ent.OrgPrice, item *stripe.SubscriptionItem) (*ent.OrgModule, error) {
	if item.Price == nil {
		return nil, nil
	}

	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	existing, err := tx.OrgModule.Query().
		Where(orgmodule.StripePriceID(item.Price.ID), orgmodule.SubscriptionID(orgSub.ID)).
		Only(allowCtx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if ent.IsNotFound(err) {
		builder := tx.OrgModule.Create().
			SetOwnerID(orgSub.OwnerID).
			SetSubscriptionID(orgSub.ID).
			SetPriceID(price.ID)
		em.ApplyStripeSubscriptionItem(builder, item)
		return builder.Save(allowCtx)
	}

	builder := tx.OrgModule.UpdateOne(existing)
	em.ApplyStripeSubscriptionItem(builder, item)
	builder.SetPriceID(price.ID)
	_, err = builder.Save(allowCtx)
	return existing, err
}
