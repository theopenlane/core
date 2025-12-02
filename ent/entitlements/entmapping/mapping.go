package entmapping

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/models"
	catalog "github.com/theopenlane/ent/entitlements"
	ent "github.com/theopenlane/ent/generated"
)

// StripePriceToOrgPrice converts a stripe.Price to an OrgPrice
func StripePriceToOrgPrice(p *stripe.Price) *ent.OrgPrice {
	if p == nil {
		return nil
	}

	interval := ""
	if p.Recurring != nil {
		interval = string(p.Recurring.Interval)
	}

	currency := string(p.Currency)

	price := models.Price{
		Amount:   float64(p.UnitAmount) / 100.0, //nolint:mnd
		Interval: interval,
		Currency: currency,
	}

	status := "inactive"
	if p.Active {
		status = "active"
	}

	productID := ""
	if p.Product != nil {
		productID = p.Product.ID
	}

	return &ent.OrgPrice{
		StripePriceID: p.ID,
		Price:         price,
		Status:        status,
		Active:        p.Active,
		ProductID:     productID,
	}
}

// StripeProductToOrgProduct converts a stripe.Product to an OrgProduct
func StripeProductToOrgProduct(p *stripe.Product) *ent.OrgProduct {
	if p == nil {
		return nil
	}

	module := p.Metadata["module"]
	if module == "" {
		module = p.Name
	}

	status := "inactive"
	if p.Active {
		status = "active"
	}

	var priceID string
	if p.DefaultPrice != nil {
		priceID = p.DefaultPrice.ID
	}

	return &ent.OrgProduct{
		Module:          module,
		StripeProductID: p.ID,
		Status:          status,
		Active:          p.Active,
		PriceID:         priceID,
	}
}

// StripeSubscriptionToOrgSubscription converts a stripe.Subscription and
// OrganizationCustomer info to an OrgSubscription.
func StripeSubscriptionToOrgSubscription(sub *stripe.Subscription) *ent.OrgSubscription {
	if sub == nil {
		return nil
	}

	status := string(sub.Status)

	orgSub := &ent.OrgSubscription{
		StripeSubscriptionID:     sub.ID,
		StripeSubscriptionStatus: status,
		Active:                   entitlements.IsSubscriptionActive(sub.Status),
		TrialExpiresAt:           timePtr(time.Unix(sub.TrialEnd, 0)),
		DaysUntilDue:             int64ToStringPtr(sub.DaysUntilDue),
	}

	return orgSub
}

// StripeSubscriptionItemToOrgModule converts a stripe.SubscriptionItem to an OrgModule
func StripeSubscriptionItemToOrgModule(item *stripe.SubscriptionItem) *ent.OrgModule {
	if item == nil || item.Price == nil {
		return nil
	}

	interval := ""
	if item.Price.Recurring != nil {
		interval = string(item.Price.Recurring.Interval)
	}

	price := models.Price{
		Amount:   float64(item.Price.UnitAmount) / 100.0, //nolint:mnd
		Interval: interval,
		Currency: string(item.Price.Currency),
	}

	moduleKey := ""
	if item.Price.Product != nil && item.Price.Product.Metadata != nil {
		moduleKey = item.Price.Product.Metadata["module"]
	}

	visibility := ""
	if v, ok := item.Price.Metadata["visibility"]; ok {
		visibility = v
	}

	status := "inactive"
	if item.Price.Active {
		status = "active"
	}

	return &ent.OrgModule{
		Module:          models.OrgModule(moduleKey),
		Price:           price,
		StripePriceID:   item.Price.ID,
		Status:          status,
		Visibility:      visibility,
		ModuleLookupKey: item.Price.LookupKey,
	}
}

func int64ToStringPtr(i int64) *string {
	s := fmt.Sprintf("%d", i)
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// OrgPriceSetter defines the methods needed to set OrgPrice fields on ent builders.
// It uses a generic type parameter so that the concrete builder type is returned
// from each setter, enabling call chaining.
type OrgPriceSetter[T any] interface {
	SetStripePriceID(string) T
	SetPrice(models.Price) T
	SetStatus(string) T
	SetActive(bool) T
	SetProductID(string) T
}

// ApplyStripePrice sets fields on the provided ent builder from the stripe.Price
func ApplyStripePrice[T OrgPriceSetter[T]](b T, p *stripe.Price) T {
	if p == nil {
		return b
	}

	interval := ""
	if p.Recurring != nil {
		interval = string(p.Recurring.Interval)
	}

	price := models.Price{
		Amount:   float64(p.UnitAmount) / 100.0, // nolint:mnd
		Interval: interval,
		Currency: string(p.Currency),
	}

	status := "inactive"
	if p.Active {
		status = "active"
	}

	productID := ""
	if p.Product != nil {
		productID = p.Product.ID
	}

	b.SetStripePriceID(p.ID)
	b.SetPrice(price)
	b.SetStatus(status)
	b.SetActive(p.Active)
	b.SetProductID(productID)

	return b
}

// OrgProductSetter defines the methods needed to set OrgProduct fields on ent builders
type OrgProductSetter[T any] interface {
	SetModule(string) T
	SetStripeProductID(string) T
	SetStatus(string) T
	SetActive(bool) T
	SetPriceID(string) T
}

// ApplyStripeProduct sets fields on the provided ent builder from the stripe.Product
func ApplyStripeProduct[T OrgProductSetter[T]](b T, p *stripe.Product) T {
	if p == nil {
		return b
	}

	module := p.Metadata["module"]
	if module == "" {
		module = p.Name
	}

	status := "inactive"
	if p.Active {
		status = "active"
	}

	var priceID string
	if p.DefaultPrice != nil {
		priceID = p.DefaultPrice.ID
	}

	b.SetModule(module)
	b.SetStripeProductID(p.ID)
	b.SetStatus(status)
	b.SetActive(p.Active)
	b.SetPriceID(priceID)

	return b
}

// OrgSubscriptionSetter defines the methods needed to set OrgSubscription fields on ent builders
type OrgSubscriptionSetter[T any] interface {
	SetStripeSubscriptionID(string) T
	SetStripeSubscriptionStatus(string) T
	SetActive(bool) T
	SetProductPrice(models.Price) T
	SetTrialExpiresAt(time.Time) T
	SetDaysUntilDue(string) T
	SetFeatures([]string) T
	SetFeatureLookupKeys([]string) T
}

// ApplyStripeSubscription sets fields on the ent builder from the stripe.Subscription and customer info
func ApplyStripeSubscription[T OrgSubscriptionSetter[T]](b T, sub *stripe.Subscription) T {
	if sub == nil {
		return b
	}

	var price models.Price

	if sub.Items != nil && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		if item.Price != nil {
			interval := ""
			if item.Price.Recurring != nil {
				interval = string(item.Price.Recurring.Interval)
			}

			price = models.Price{
				Amount:   float64(item.Price.UnitAmount) / 100.0, // nolint:mnd
				Interval: interval,
				Currency: string(item.Price.Currency),
			}
		}
	}

	status := string(sub.Status)

	b.SetStripeSubscriptionID(sub.ID)
	b.SetStripeSubscriptionStatus(status)
	b.SetActive(entitlements.IsSubscriptionActive(sub.Status))

	b.SetProductPrice(price)
	b.SetTrialExpiresAt(time.Unix(sub.TrialEnd, 0))
	b.SetDaysUntilDue(fmt.Sprintf("%d", sub.DaysUntilDue))

	return b
}

// OrgModuleSetter defines the methods needed to set OrgModule fields on ent builders
type OrgModuleSetter[T any] interface {
	SetModule(models.OrgModule) T
	SetPrice(models.Price) T
	SetStripePriceID(string) T
	SetStatus(string) T
	SetVisibility(string) T
	SetModuleLookupKey(string) T
	SetActive(bool) T
}

// ApplyStripeSubscriptionItem sets fields on the ent builder from the stripe.SubscriptionItem
func ApplyStripeSubscriptionItem[T OrgModuleSetter[T]](ctx context.Context, b T, item *stripe.SubscriptionItem,
	client *entitlements.StripeClient, status string) T {
	if item == nil || item.Price == nil {
		return b
	}

	interval := ""
	if item.Price.Recurring != nil {
		interval = string(item.Price.Recurring.Interval)
	}

	price := models.Price{
		Amount:   float64(item.Price.UnitAmount) / 100.0, // nolint:mnd
		Interval: interval,
		Currency: string(item.Price.Currency),
	}

	productMetadata := GetProductMetadata(ctx, item.Price.Product, client)
	if moduleKey := strings.TrimSpace(productMetadata["module"]); moduleKey != "" {
		b.SetModule(models.OrgModule(moduleKey))
	}

	if visibility := strings.TrimSpace(item.Price.Metadata["visibility"]); visibility != "" {
		b.SetVisibility(visibility)
	}

	switch status {
	case string(stripe.SubscriptionStatusCanceled), string(stripe.SubscriptionStatusPastDue),
		string(stripe.SubscriptionStatusPaused):
		b.SetActive(false)
	}

	b.SetPrice(price)
	b.SetStripePriceID(item.Price.ID)
	b.SetStatus(status)
	b.SetModuleLookupKey(item.Price.LookupKey)

	return b
}

// PopulatePricesForOrganizationCustomer sets Prices on the OrganizationCustomer and returns it.
func PopulatePricesForOrganizationCustomer(o *entitlements.OrganizationCustomer, useSandbox bool) *entitlements.OrganizationCustomer {
	priceIDs := catalog.TrialMonthlyPriceIDs(useSandbox)

	o.Prices = make([]entitlements.Price, len(priceIDs))
	for i, priceID := range priceIDs {
		o.Prices[i] = entitlements.Price{ID: priceID}
	}

	return o
}

// GetProductMetadata retrieves the metadata for a Stripe product, either from the provided
// product object or by fetching the full product details using the Stripe client.
func GetProductMetadata(ctx context.Context, product *stripe.Product, client *entitlements.StripeClient) map[string]string {
	if product == nil {
		return nil
	}

	if product.Metadata != nil {
		return product.Metadata
	}

	if client == nil {
		return nil
	}

	fullProduct, err := client.FindProductByID(ctx, product.ID)
	if err != nil || fullProduct == nil {
		return nil
	}

	return fullProduct.Metadata
}
