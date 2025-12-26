package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgprice"
	"github.com/theopenlane/core/internal/ent/generated/orgproduct"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/middleware/transaction"

	em "github.com/theopenlane/core/internal/entitlements/entmapping"
)

// syncSubscriptionItemsWithStripe ensures OrgProduct, OrgPrice, and OrgModule
// records exist and are updated based on the given Stripe subscription data.
func (h *Handler) syncSubscriptionItemsWithStripe(ctx context.Context, subscription *stripe.Subscription, items []*stripe.SubscriptionItem, subStatus stripe.SubscriptionStatus) error {
	orgSub, err := getOrgSubscription(ctx, subscription)
	if err != nil {
		return err
	}

	var existingModules []models.OrgModule

	for _, item := range items {
		if item.Price == nil || item.Price.Product == nil {
			continue
		}

		prod, err := upsertOrgProduct(ctx, orgSub, item.Price.Product)
		if err != nil {
			return err
		}

		logx.FromContext(ctx).Info().Str("product_id", prod.StripeProductID).Msg("org product created")

		price, err := upsertOrgPrice(ctx, orgSub, prod, item.Price)
		if err != nil {
			return err
		}

		logx.FromContext(ctx).Info().Str("price_subscription_ID", price.SubscriptionID).Msg("org price created for subscription")

		mod, err := upsertOrgModule(ctx, orgSub, price, item, h.Entitlements, string(subStatus))
		if err != nil {
			return err
		}

		existingModules = append(existingModules, mod.Module)
	}

	err = reconcileModules(ctx, orgSub, existingModules)
	if err != nil {
		return err
	}

	return nil
}

// upsertOrgProduct creates or updates an OrgProduct based on the Stripe product data
func upsertOrgProduct(ctx context.Context, orgSub *ent.OrgSubscription, p *stripe.Product) (*ent.OrgProduct, error) {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	existing, err := tx.OrgProduct.Query().Where(orgproduct.StripeProductID(p.ID), orgproduct.SubscriptionID(orgSub.ID)).Only(allowCtx)
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

// upsertOrgPrice creates or updates an OrgPrice based on the Stripe price data
func upsertOrgPrice(ctx context.Context, orgSub *ent.OrgSubscription, prod *ent.OrgProduct, price *stripe.Price) (*ent.OrgPrice, error) {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	existing, err := tx.OrgPrice.Query().Where(orgprice.StripePriceID(price.ID), orgprice.SubscriptionID(orgSub.ID)).Only(allowCtx)
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

// upsertOrgModule creates or updates an OrgModule based on the Stripe subscription item data
func upsertOrgModule(ctx context.Context, orgSub *ent.OrgSubscription, price *ent.OrgPrice, item *stripe.SubscriptionItem,
	client *entitlements.StripeClient, status string) (*ent.OrgModule, error) {
	if item.Price == nil {
		return nil, nil
	}

	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	productMetadata := em.GetProductMetadata(ctx, item.Price.Product, client)
	moduleKey := strings.TrimSpace(productMetadata["module"])

	if moduleKey == models.CatalogTrustCenterModule.String() {
		// use a fresh context to avoid inheriting the OrgSubscriptionContextKey bypass and others
		newCtx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
			SubjectID:          orgSub.CreatedBy,
			OrganizationID:     orgSub.OwnerID,
			OrganizationIDs:    []string{orgSub.OwnerID},
			AuthenticationType: auth.JWTAuthentication,
		})

		// add tx back to the context
		newCtx = transaction.NewContext(newCtx, tx)

		newCtx = privacy.DecisionContext(newCtx, privacy.Allow)

		_, err := tx.TrustCenter.Create().SetOwnerID(orgSub.OwnerID).
			Save(newCtx)
		if err != nil && !errors.Is(err, hooks.ErrNotSingularTrustCenter) {
			return nil, err
		}
	}

	// include softdeleted modules in the query
	// if the module was previously marked as deleted, bring it back
	// instead of making a new record/row
	queryCtx := context.WithValue(allowCtx, entx.SoftDeleteSkipKey{}, true)

	existing, err := tx.OrgModule.Query().Where(
		orgmodule.And(
			orgmodule.ModuleEQ(models.OrgModule(moduleKey)),
			orgmodule.OwnerID(orgSub.OwnerID),
		),
		orgmodule.SubscriptionID(orgSub.ID)).Only(queryCtx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if ent.IsNotFound(err) {
		builder := tx.OrgModule.Create().
			SetOwnerID(orgSub.OwnerID).
			SetSubscriptionID(orgSub.ID).
			SetPriceID(price.ID)

		em.ApplyStripeSubscriptionItem(ctx, builder, item, client, status)

		return builder.Save(allowCtx)
	}

	builder := tx.OrgModule.UpdateOne(existing)

	em.ApplyStripeSubscriptionItem(ctx, builder, item, client, status)

	builder.SetPriceID(price.ID)

	if !existing.DeletedAt.IsZero() {
		builder.ClearDeletedAt().ClearDeletedBy()
	}

	return builder.Save(allowCtx)
}

// reconcileModules makes sure to match the modules accessible to the org
// with what is in stripe
func reconcileModules(ctx context.Context, orgSub *ent.OrgSubscription, currentModules []models.OrgModule) error {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	_, err := tx.OrgModule.Delete().Where(
		orgmodule.And(
			orgmodule.OwnerID(orgSub.OwnerID),
			orgmodule.SubscriptionID(orgSub.ID),
			orgmodule.ModuleNotIn(currentModules...),
		),
	).Exec(allowCtx)

	return err
}

// removeAllModules drops all modules except the base one
func (h *Handler) removeAllModules(ctx context.Context, subscription *stripe.Subscription) error {
	orgSub, err := getOrgSubscription(ctx, subscription)
	if err != nil {
		return err
	}

	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})
	tx := transaction.FromContext(ctx)

	_, err = tx.OrgModule.Delete().Where(
		orgmodule.And(
			orgmodule.OwnerID(orgSub.OwnerID),
			orgmodule.SubscriptionID(orgSub.ID),
			orgmodule.ModuleNEQ(models.CatalogBaseModule),
		),
	).Exec(allowCtx)

	return err
}
