package entitlements

import (
	"strings"

	"github.com/stripe/stripe-go/v81"
	"github.com/theopenlane/utils/rout"
)

// OrganizationCustomer is a struct which holds both internal organization infos and external stripe infos
type OrganizationCustomer struct {
	OrganizationID             string `json:"organization_id"`
	OrganizationSettingsID     string `json:"organization_settings_id"`
	StripeCustomerID           string `json:"stripe_customer_id"`
	OrganizationName           string `json:"organization_name"`
	PersonalOrg                bool   `json:"personal_org"`
	OrganizationSubscriptionID string `json:"organization_subscription_id"`
	StripeSubscriptionID       string `json:"stripe_subscription_id"`
	PaymentMethodAdded         bool   `json:"payment_method_added"`
	Features                   []string
	FeatureNames               []string
	Subscription
	ContactInfo
}

// ContactInfo holds the contact information for the organization
type ContactInfo struct {
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	City       *string `form:"city"`
	Country    *string `form:"country"`
	Line1      *string `form:"line1"`
	Line2      *string `form:"line2"`
	PostalCode *string `form:"postal_code"`
	State      *string `form:"state"`
}

func (o *OrganizationCustomer) MapToStripeCustomer() *stripe.CustomerParams {
	return &stripe.CustomerParams{
		Email: &o.Email,
		Name:  &o.OrganizationID,
		Phone: &o.Phone,
		Address: &stripe.AddressParams{
			Line1:      o.Line1,
			Line2:      o.Line2,
			City:       o.City,
			State:      o.State,
			PostalCode: o.PostalCode,
			Country:    o.Country,
		},
		Metadata: map[string]string{
			"organization_id":              o.OrganizationID,
			"organization_settings_id":     o.OrganizationSettingsID,
			"organization_name":            o.OrganizationName,
			"organization_subscription_id": o.OrganizationSubscriptionID,
		},
	}
}

// Validate checks if the OrganizationCustomer contains necessary fields
func (o *OrganizationCustomer) Validate() error {
	o.OrganizationID = strings.TrimSpace(o.OrganizationID)
	o.Email = strings.TrimSpace(o.Email)

	switch {
	case o.OrganizationID == "":
		return rout.NewMissingRequiredFieldError("organization_id")
	case o.Email == "":
		return rout.NewMissingRequiredFieldError("billing_email")
	}

	return nil
}

// Customer holds the customer information
type Customer struct {
	ID             string `json:"customer_id" yaml:"customer_id"`
	Email          string `json:"email" yaml:"email"`
	Phone          string `json:"phone" yaml:"phone"`
	Address        string `json:"address" yaml:"address"`
	Plans          []Plan `json:"plans" yaml:"plans"`
	Metadata       map[string]string
	StripeParams   *stripe.CustomerParams
	StripeCustomer []stripe.Customer
}

// Plan is the recurring context that holds the payment information
type Plan struct {
	ID                 string `json:"plan_id" yaml:"plan_id"`
	Product            string `json:"product_id" yaml:"product_id"`
	Price              string `json:"price_id" yaml:"price_id"`
	StartDate          int64  `json:"start_date" yaml:"start_date"`
	EndDate            int64  `json:"end_date" yaml:"end_date"`
	StripeParams       *stripe.SubscriptionParams
	StripeSubscription []stripe.Subscription
	Products           []Product
	Features           []Feature
	TrialEnd           int64
	Status             string
}

// Subscription is the recurring context that holds the payment information
type Subscription struct {
	ID                 string `json:"plan_id" yaml:"plan_id"`
	ProductID          string `json:"product_id" yaml:"product_id"`
	PriceID            string `json:"price_id" yaml:"price_id"`
	StartDate          int64  `json:"start_date" yaml:"start_date"`
	EndDate            int64  `json:"end_date" yaml:"end_date"`
	StripeParams       *stripe.SubscriptionParams
	StripeSubscription []stripe.Subscription
	StripeProduct      []stripe.Product
	StripeFeature      []stripe.ProductFeature
	Products           []Product
	Features           []Feature
	Prices             []Price
	TrialEnd           int64
	Status             string
	StripeCustomerID   string
	OrganizationID     string
	DaysUntilDue       int64
	Metadata           map[string]string
}

// BillingPortalSession holds the billing portal session information
type BillingPortalSession struct {
	ManageSubscription string `json:"manage_subscription"`
	PaymentMethods     string `json:"payment_methods"`
	Cancellation       string `json:"cancellation"`
	HomePage           string `json:"home_page"`
}

// Product holds what we'd more commply call a "tier"
type Product struct {
	ID            string                `json:"product_id" yaml:"product_id"`
	Name          string                `json:"name" yaml:"name"`
	Description   string                `json:"description,omitempty" yaml:"description,omitempty"`
	Features      []Feature             `json:"features" yaml:"features"`
	Prices        []Price               `json:"prices" yaml:"prices"`
	StripeParams  *stripe.ProductParams `json:"stripe_params,omitempty" yaml:"stripe_params,omitempty"`
	StripeProduct []stripe.Product      `json:"stripe_product,omitempty" yaml:"stripe_product,omitempty"`
}

// Price holds stripe price params and the associated Product
type Price struct {
	ID           string              `json:"price_id" yaml:"price_id"`
	Price        float64             `json:"price" yaml:"price"`
	Currency     string              `json:"currency" yaml:"currency"`
	ProductID    string              `json:"product_id" yaml:"-"`
	ProductName  string              `json:"product_name" yaml:"product_name"`
	Interval     string              `json:"interval" yaml:"interval"`
	StripeParams *stripe.PriceParams `json:"stripe_params,omitempty" yaml:"stripe_params,omitempty"`
	StripePrice  []stripe.Price      `json:"stripe_price,omitempty" yaml:"stripe_price,omitempty"`
}

// Checkout holds the checkout information
type Checkout struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Feature are part of a product
type Feature struct {
	ID        string `json:"id" yaml:"id"`
	Name      string `json:"name" yaml:"name"`
	Lookupkey string `json:"lookupkey" yaml:"lookupkey"`
}

type ProductFeature struct {
	FeatureID string `json:"feature_id" yaml:"feature_id"`
	ProductID string `json:"product_id"`
	Name      string `json:"name" yaml:"name"`
	Lookupkey string `json:"lookupkey" yaml:"lookupkey"`
}
