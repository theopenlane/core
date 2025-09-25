package reconciler

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/models"
)

// Reconciler reconciles organization subscriptions with Stripe
type Reconciler struct {
	db     *ent.Client
	stripe *entitlements.StripeClient

	dryRun bool
	writer io.Writer
}

// actionRow represents a single reconciliation action to be performed
type actionRow struct {
	OrgID  string
	Action string
}

// Option configures the Reconciler using the functional options pattern
type Option func(*Reconciler)

// WithDB sets the ent client
func WithDB(db *ent.Client) Option {
	return func(r *Reconciler) {
		r.db = db
	}
}

// WithStripeClient sets the stripe client
func WithStripeClient(sc *entitlements.StripeClient) Option {
	return func(r *Reconciler) {
		r.stripe = sc
	}
}

// WithDryRun enables dry run mode. If writer is nil output is sent to os.Stdout
func WithDryRun(writer io.Writer) Option {
	return func(r *Reconciler) {
		r.dryRun = true
		r.writer = writer
	}
}

var (
	ErrMissingStripeClient   = fmt.Errorf("missing stripe client")
	ErrMissingDBClient       = fmt.Errorf("missing database client")
	ErrMissingSubscriptionID = fmt.Errorf("missing organization subscription ID")
	ErrMultiplePrices        = fmt.Errorf("multiple prices found for customer")
)

// New creates a new Reconciler instance with the provided options
func New(opts ...Option) (*Reconciler, error) {
	r := &Reconciler{}
	for _, opt := range opts {
		opt(r)
	}

	if r.stripe == nil {
		return nil, ErrMissingStripeClient
	}

	if r.dryRun && r.writer == nil {
		r.writer = os.Stdout
	}

	return r, nil
}

// ReconcileResult contains the results of a reconcile operation with actions that need to be taken
type ReconcileResult struct {
	Actions []actionRow
}

// Reconcile iterates through all organizations and ensures customer records and subscriptions exist in Stripe
func (r *Reconciler) Reconcile(ctx context.Context) (*ReconcileResult, error) {
	if r.db == nil {
		return nil, ErrMissingDBClient
	}

	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)
	orgs, err := r.db.Organization.Query().
		WithOrgSubscriptions().
		WithSetting().
		Where(
			organization.And(
				organization.DeletedAtIsNil(),
				organization.Not(organization.ID("01101101011010010111010001100010")),
			),
		).
		All(internalCtx)
	if err != nil {
		return nil, fmt.Errorf("query organizations: %w", err)
	}

	var rows []actionRow
	for _, org := range orgs {
		if r.dryRun {
			action, err := r.analyzeOrg(ctx, org)
			if err != nil {
				return nil, err
			}

			if action != "" {
				rows = append(rows, actionRow{OrgID: org.ID, Action: action})
			}

			continue
		}

		if err := r.reconcileOrg(ctx, org); err != nil {
			return nil, err
		}
	}

	return &ReconcileResult{Actions: rows}, nil
}

// reconcileOrg ensures the organization has a customer and subscription in Stripe, creating them if missing
func (r *Reconciler) reconcileOrg(ctx context.Context, org *ent.Organization) error {
	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)

	var sub *ent.OrgSubscription
	if len(org.Edges.OrgSubscriptions) > 0 {
		sub = org.Edges.OrgSubscriptions[0]
	} else {
		var err error
		sub, err = r.db.OrgSubscription.Create().SetOwnerID(org.ID).Save(internalCtx)
		if err != nil {
			return fmt.Errorf("create subscription: %w", err)
		}
	}

	cust := &entitlements.OrganizationCustomer{
		OrganizationID:             org.ID,
		OrganizationSettingsID:     org.Edges.Setting.ID,
		OrganizationSubscriptionID: sub.ID,
		OrganizationName:           org.Name,
		ContactInfo: entitlements.ContactInfo{
			Email:      org.Edges.Setting.BillingEmail,
			Phone:      org.Edges.Setting.BillingPhone,
			Line1:      &org.Edges.Setting.BillingAddress.Line1,
			Line2:      &org.Edges.Setting.BillingAddress.Line2,
			City:       &org.Edges.Setting.BillingAddress.City,
			State:      &org.Edges.Setting.BillingAddress.State,
			Country:    &org.Edges.Setting.BillingAddress.Country,
			PostalCode: &org.Edges.Setting.BillingAddress.PostalCode,
		},
	}

	// Set metadata for personal organizations
	if org.PersonalOrg {
		cust.Metadata = map[string]string{
			"personal_org": "true",
		}
	}

	if err := r.stripe.FindOrCreateCustomer(ctx, cust); err != nil {
		return fmt.Errorf("stripe customer: %w", err)
	}

	if sub.StripeSubscriptionID == "" {
		if err := r.updateSubscription(ctx, cust); err != nil {
			return err
		}
	}

	return nil
}

// updateSubscription updates the organization subscription in the database with current Stripe data
func (r *Reconciler) updateSubscription(ctx context.Context, c *entitlements.OrganizationCustomer) error {
	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)
	if c.OrganizationSubscriptionID == "" {
		return ErrMissingSubscriptionID
	}

	if len(c.Prices) > 1 {
		return ErrMultiplePrices
	}

	price := models.Price{}
	if len(c.Prices) == 1 {
		price = models.Price{
			Amount:   c.Prices[0].Price,
			Currency: c.Prices[0].Currency,
			Interval: c.Prices[0].Interval,
		}
	}

	trialExpiresAt := time.Unix(0, 0)
	if c.Status == string(stripe.SubscriptionStatusTrialing) {
		trialExpiresAt = time.Unix(c.TrialEnd, 0)
	}

	expiresAt := time.Unix(0, 0)
	if c.EndDate > 0 {
		expiresAt = time.Unix(c.EndDate, 0)
	}

	active := c.Status == string(stripe.SubscriptionStatusActive) || c.Status == string(stripe.SubscriptionStatusTrialing)

	update := r.db.OrgSubscription.UpdateOneID(c.OrganizationSubscriptionID).
		SetStripeSubscriptionID(c.StripeSubscriptionID).
		SetStripeSubscriptionStatus(c.Subscription.Status).
		SetActive(active).
		SetFeatures(c.FeatureNames).
		SetFeatureLookupKeys(c.Features).
		SetProductPrice(price)

	if c.Status == string(stripe.SubscriptionStatusTrialing) {
		update.SetTrialExpiresAt(trialExpiresAt)
	} else {
		update.SetExpiresAt(expiresAt)
	}

	if err := update.Exec(internalCtx); err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	log.Ctx(ctx).Info().Str("org", c.OrganizationID).Msg("reconciled subscription")

	return nil
}

// analyzeOrg checks the organization subscription status and returns a description of the action needed
func (r *Reconciler) analyzeOrg(ctx context.Context, org *ent.Organization) (string, error) {
	var sub *ent.OrgSubscription
	if len(org.Edges.OrgSubscriptions) > 0 {
		sub = org.Edges.OrgSubscriptions[0]
	}

	customerMissing := sub == nil
	subscriptionMissing := sub == nil

	if !customerMissing {
		if _, err := r.stripe.GetCustomerByStripeID(ctx, *org.StripeCustomerID); err != nil {
			customerMissing = true
		}
	}

	if !subscriptionMissing {
		if _, err := r.stripe.GetSubscriptionByID(ctx, sub.StripeSubscriptionID); err != nil {
			subscriptionMissing = true
		}
	}

	switch {
	case customerMissing && subscriptionMissing:
		return "create stripe customer & subscription", nil
	case customerMissing:
		return "create stripe customer", nil
	case subscriptionMissing:
		return "create stripe subscription", nil
	default:
		return "", nil
	}
}

