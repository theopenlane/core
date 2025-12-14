package entmapping

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v84"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
)

// priceBuilder defines the methods needed to set OrgPrice fields on ent builders
type priceBuilder struct {
	stripePriceID string
	price         models.Price
	status        string
	active        bool
	productID     string
}

func (b *priceBuilder) SetStripePriceID(id string) *priceBuilder { b.stripePriceID = id; return b }
func (b *priceBuilder) SetPrice(p models.Price) *priceBuilder    { b.price = p; return b }
func (b *priceBuilder) SetStatus(s string) *priceBuilder         { b.status = s; return b }
func (b *priceBuilder) SetActive(a bool) *priceBuilder           { b.active = a; return b }
func (b *priceBuilder) SetProductID(id string) *priceBuilder     { b.productID = id; return b }

// productBuilder defines the methods needed to set OrgProduct fields on ent builders
type productBuilder struct {
	module          string
	stripeProductID string
	status          string
	active          bool
	priceID         string
}

func (b *productBuilder) SetModule(m string) *productBuilder { b.module = m; return b }
func (b *productBuilder) SetStripeProductID(id string) *productBuilder {
	b.stripeProductID = id
	return b
}
func (b *productBuilder) SetStatus(s string) *productBuilder   { b.status = s; return b }
func (b *productBuilder) SetActive(a bool) *productBuilder     { b.active = a; return b }
func (b *productBuilder) SetPriceID(id string) *productBuilder { b.priceID = id; return b }

// subscriptionBuilder defines the methods needed to set OrgSubscription fields on ent builders
type subscriptionBuilder struct {
	stripeSubscriptionID     string
	stripeSubscriptionStatus string
	active                   bool
	stripeCustomerID         string
	productPrice             models.Price
	trialExpiresAt           time.Time
	daysUntilDue             string
	features                 []string
	featureLookupKeys        []string
}

func (b *subscriptionBuilder) SetStripeSubscriptionID(id string) *subscriptionBuilder {
	b.stripeSubscriptionID = id
	return b
}
func (b *subscriptionBuilder) SetStripeSubscriptionStatus(s string) *subscriptionBuilder {
	b.stripeSubscriptionStatus = s
	return b
}
func (b *subscriptionBuilder) SetActive(a bool) *subscriptionBuilder { b.active = a; return b }
func (b *subscriptionBuilder) SetStripeCustomerID(id string) *subscriptionBuilder {
	b.stripeCustomerID = id
	return b
}

func (b *subscriptionBuilder) SetProductPrice(p models.Price) *subscriptionBuilder {
	b.productPrice = p
	return b
}
func (b *subscriptionBuilder) SetTrialExpiresAt(t time.Time) *subscriptionBuilder {
	b.trialExpiresAt = t
	return b
}
func (b *subscriptionBuilder) SetDaysUntilDue(s string) *subscriptionBuilder {
	b.daysUntilDue = s
	return b
}
func (b *subscriptionBuilder) SetFeatures(f []string) *subscriptionBuilder { b.features = f; return b }
func (b *subscriptionBuilder) SetFeatureLookupKeys(f []string) *subscriptionBuilder {
	b.featureLookupKeys = f
	return b
}

// moduleBuilder defines the methods needed to set OrgModule fields on ent builders
type moduleBuilder struct {
	module          models.OrgModule
	price           models.Price
	stripePriceID   string
	status          string
	visibility      string
	moduleLookupKey string
	active          bool
}

func (b *moduleBuilder) SetModule(m models.OrgModule) *moduleBuilder { b.module = m; return b }
func (b *moduleBuilder) SetPrice(p models.Price) *moduleBuilder      { b.price = p; return b }
func (b *moduleBuilder) SetStripePriceID(id string) *moduleBuilder   { b.stripePriceID = id; return b }
func (b *moduleBuilder) SetStatus(s string) *moduleBuilder           { b.status = s; return b }
func (b *moduleBuilder) SetVisibility(v string) *moduleBuilder       { b.visibility = v; return b }
func (b *moduleBuilder) SetModuleLookupKey(k string) *moduleBuilder  { b.moduleLookupKey = k; return b }
func (b *moduleBuilder) SetActive(a bool) *moduleBuilder             { b.active = a; return b }

func TestStripePriceToOrgPrice(t *testing.T) {
	p := &stripe.Price{
		ID:         "price_123",
		UnitAmount: 1500,
		Currency:   stripe.CurrencyUSD,
		Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalMonth},
		Active:     true,
		Product:    &stripe.Product{ID: "prod_123"},
	}

	got := StripePriceToOrgPrice(p)
	want := &ent.OrgPrice{
		StripePriceID: "price_123",
		Price:         models.Price{Amount: 15, Interval: "month", Currency: "usd"},
		Status:        "active",
		Active:        true,
		ProductID:     "prod_123",
	}

	require.Equal(t, want, got)
}

func TestStripePriceToOrgPrice_Nil(t *testing.T) {
	require.Nil(t, StripePriceToOrgPrice(nil))
}

func TestStripeProductToOrgProduct(t *testing.T) {
	p := &stripe.Product{
		ID:           "prod_123",
		Name:         "Example",
		Metadata:     map[string]string{"module": "example"},
		Active:       true,
		DefaultPrice: &stripe.Price{ID: "price_default"},
	}

	got := StripeProductToOrgProduct(p)
	want := &ent.OrgProduct{
		Module:          "example",
		StripeProductID: "prod_123",
		Status:          "active",
		Active:          true,
		PriceID:         "price_default",
	}

	require.Equal(t, want, got)
}

func TestStripeProductToOrgProduct_NoMetadata(t *testing.T) {
	p := &stripe.Product{ID: "prod_1", Name: "Fallback", Active: true}
	got := StripeProductToOrgProduct(p)
	require.Equal(t, "Fallback", got.Module)
}

func TestStripeSubscriptionToOrgSubscription(t *testing.T) {
	sub := &stripe.Subscription{
		ID:           "sub_123",
		Status:       stripe.SubscriptionStatusActive,
		Customer:     &stripe.Customer{ID: "cus_123"},
		TrialEnd:     1700000000,
		DaysUntilDue: 7,
		Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
			{
				Price: &stripe.Price{
					ID:         "price_123",
					UnitAmount: 2000,
					Currency:   stripe.CurrencyUSD,
					Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalYear},
					Product:    &stripe.Product{ID: "prod_123", Name: "Pro"},
				},
			},
		}},
	}

	got := StripeSubscriptionToOrgSubscription(sub)

	want := &ent.OrgSubscription{
		StripeSubscriptionID:     "sub_123",
		StripeSubscriptionStatus: "active",
		Active:                   true,
		TrialExpiresAt:           timePtr(time.Unix(1700000000, 0)),
		DaysUntilDue:             int64ToStringPtr(7),
	}

	require.Equal(t, want, got)
}

func TestStripeSubscriptionItemToOrgModule(t *testing.T) {
	item := &stripe.SubscriptionItem{
		Price: &stripe.Price{
			ID:         "price_123",
			UnitAmount: 3000,
			Currency:   stripe.CurrencyUSD,
			Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalMonth},
			Active:     true,
			LookupKey:  "lookup",
			Metadata:   map[string]string{"visibility": "public"},
			Product: &stripe.Product{
				Metadata: map[string]string{"module": "mod1"},
			},
		},
	}

	got := StripeSubscriptionItemToOrgModule(item)
	want := &ent.OrgModule{
		Module:          "mod1",
		Price:           models.Price{Amount: 30, Interval: "month", Currency: "usd"},
		StripePriceID:   "price_123",
		Status:          "active",
		Visibility:      "public",
		ModuleLookupKey: "lookup",
	}

	require.Equal(t, want, got)
}

func TestApplyStripePrice(t *testing.T) {
	p := &stripe.Price{
		ID:         "price_123",
		UnitAmount: 1500,
		Currency:   stripe.CurrencyUSD,
		Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalMonth},
		Active:     true,
		Product:    &stripe.Product{ID: "prod_123"},
	}

	b := &priceBuilder{}
	got := ApplyStripePrice(b, p)

	expected := &priceBuilder{
		stripePriceID: "price_123",
		price:         models.Price{Amount: 15, Interval: "month", Currency: "usd"},
		status:        "active",
		active:        true,
		productID:     "prod_123",
	}

	require.Equal(t, b, got)
	require.Equal(t, expected, b)
}

func TestApplyStripeProduct(t *testing.T) {
	p := &stripe.Product{
		ID:           "prod_123",
		Name:         "Example",
		Metadata:     map[string]string{"module": "example"},
		Active:       true,
		DefaultPrice: &stripe.Price{ID: "price_default"},
	}

	b := &productBuilder{}
	got := ApplyStripeProduct(b, p)

	expected := &productBuilder{
		module:          "example",
		stripeProductID: "prod_123",
		status:          "active",
		active:          true,
		priceID:         "price_default",
	}

	require.Equal(t, b, got)
	require.Equal(t, expected, b)
}

func TestApplyStripeSubscription(t *testing.T) {
	sub := &stripe.Subscription{
		ID:           "sub_123",
		Status:       stripe.SubscriptionStatusActive,
		Customer:     &stripe.Customer{ID: "cus_123"},
		TrialEnd:     1700000000,
		DaysUntilDue: 7,
		Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
			{
				Price: &stripe.Price{
					ID:         "price_123",
					UnitAmount: 2000,
					Currency:   stripe.CurrencyUSD,
					Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalYear},
					Product:    &stripe.Product{ID: "prod_123", Name: "Pro"},
				},
			},
		}},
	}

	b := &subscriptionBuilder{}
	got := ApplyStripeSubscription(b, sub)

	expected := &subscriptionBuilder{
		stripeSubscriptionID:     "sub_123",
		stripeSubscriptionStatus: "active",
		active:                   true,
		stripeCustomerID:         "",
		productPrice:             models.Price{Amount: 20, Interval: "year", Currency: "usd"},
		trialExpiresAt:           time.Unix(1700000000, 0),
		daysUntilDue:             "7",
	}

	require.Equal(t, b, got)
	require.Equal(t, expected, b)
}

func TestApplyStripeSubscriptionItem(t *testing.T) {
	item := &stripe.SubscriptionItem{
		Price: &stripe.Price{
			ID:         "price_123",
			UnitAmount: 3000,
			Currency:   stripe.CurrencyUSD,
			Recurring:  &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalMonth},
			Active:     true,
			LookupKey:  "lookup",
			Metadata:   map[string]string{"visibility": "public"},
			Product: &stripe.Product{
				Metadata: map[string]string{"module": "mod1"},
			},
		},
	}

	b := &moduleBuilder{}
	got := ApplyStripeSubscriptionItem(context.Background(), b, item, nil, "active")

	expected := &moduleBuilder{
		module:          models.OrgModule("mod1"),
		price:           models.Price{Amount: 30, Interval: "month", Currency: "usd"},
		stripePriceID:   "price_123",
		status:          "active",
		visibility:      "public",
		moduleLookupKey: "lookup",
		active:          false,
	}

	require.Equal(t, b, got)
	require.Equal(t, expected, b)
}
