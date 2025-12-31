package reconciler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	internalentitlements "github.com/theopenlane/core/internal/entitlements"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
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
	OrgID   string
	OrgName string
	Action  string
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
	ErrMissingPrice          = fmt.Errorf("missing price for customer")
	ErrMultipleCustomers     = fmt.Errorf("multiple customers found for organization")
	ErrMissingCustomer       = fmt.Errorf("missing customer for organization")
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
func (r *Reconciler) Reconcile(ctx context.Context, orgIDs []string) (*ReconcileResult, error) {
	if r.db == nil {
		return nil, ErrMissingDBClient
	}

	where := []predicate.Organization{
		organization.And(
			organization.DeletedAtIsNil(),
			organization.IDNEQ("01101101011010010111010001100010"),
			organization.PersonalOrg(false),
		),
	}

	if len(orgIDs) > 0 {
		where = append(where, organization.IDIn(orgIDs...))
	}

	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)

	orgs, err := r.db.Organization.Query().
		WithOrgSubscriptions().
		WithSetting().
		Where(
			where...,
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
				rows = append(rows, actionRow{OrgID: org.ID, OrgName: org.DisplayName, Action: action})
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

	if cust.Prices == nil && r.db != nil && r.db.EntConfig != nil {
		// Mutations that create orgs no longer populate prices inline; seed the defaults here so
		// the reconciler mirrors the legacy PopulatePricesForOrganizationCustomer behavior.
		modules := &r.db.EntConfig.Modules
		useSandbox := modules.UseSandbox

		if modules.DevMode {
			cust.Prices = internalentitlements.AllMonthlyPrices(useSandbox)
		} else {
			cust.Prices = internalentitlements.TrialMonthlyPrices(useSandbox)
		}
	}

	// Set metadata for personal organizations
	if org.PersonalOrg {
		cust.Metadata = map[string]string{
			"personal_org": "true",
		}
	}

	if err := r.stripe.FindOrCreateCustomer(ctx, cust); err != nil {
		if !errors.Is(err, entitlements.ErrNoSubscriptions) && !errors.Is(err, entitlements.ErrNoSubscriptionItems) {
			return fmt.Errorf("stripe customer: %w", err)
		}
	}

	// make sure the customer id is set on the org
	if cust.StripeCustomerID != "" && (org.StripeCustomerID == nil || *org.StripeCustomerID != cust.StripeCustomerID) {
		if err := r.db.Organization.UpdateOneID(cust.OrganizationID).
			SetStripeCustomerID(cust.StripeCustomerID).
			Exec(internalCtx); err != nil {

			// to anyone looking at this later and wondering why test logs have lots of error logs about
			// pq: duplicate key value violates unique constraint - this is due to the mocked calls
			// and can safely be ignored in tests
			return fmt.Errorf("update organization stripe customer id: %w", err)
		}
	}

	if cust.StripeSubscriptionID == "" {
		if err := r.createSubscription(ctx, cust); err != nil {
			return fmt.Errorf("create subscription: %w", err)
		}
	}

	if sub.StripeSubscriptionID == "" {
		if err := r.updateSubscription(ctx, cust); err != nil {
			return err
		}
	}

	return nil
}

var trialdays int64 = 30

// createSubscription creates a new subscription in Stripe for the organization customer with the trial settings
func (r *Reconciler) createSubscription(ctx context.Context, cust *entitlements.OrganizationCustomer) error {
	customers, err := r.stripe.SearchCustomers(ctx, fmt.Sprintf("name: '%s'", cust.OrganizationID))
	if err != nil {
		return err
	}

	if len(customers) == 0 {
		return ErrMissingCustomer
	}

	if len(customers) != 1 {
		return ErrMultipleCustomers
	}

	if len(cust.Prices) == 0 && r.db != nil && r.db.EntConfig != nil {
		// Ensure a deterministic subscription payload even when Stripe did not return prices.
		// This keeps the reconciler idempotent across retries.
		modules := &r.db.EntConfig.Modules
		if modules.DevMode {
			cust.Prices = internalentitlements.AllMonthlyPrices(modules.UseSandbox)
		} else {
			cust.Prices = internalentitlements.TrialMonthlyPrices(modules.UseSandbox)
		}
	}
	if len(cust.Prices) == 0 {
		return ErrMissingPrice
	}

	cust.Status = string(stripe.SubscriptionStatusTrialing)
	cust.TrialEnd = time.Now().AddDate(0, 0, int(trialdays)).Unix()

	_, err = r.stripe.CreateSubscriptionWithPrices(ctx, customers[0], cust)
	if err != nil {
		return err
	}

	// update subscription in db
	if err := r.updateSubscription(ctx, cust); err != nil {
		return err
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
		SetActive(active)

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
	if r.stripe == nil {
		return "", ErrMissingStripeClient
	}

	var sub *ent.OrgSubscription
	if len(org.Edges.OrgSubscriptions) > 0 {
		sub = org.Edges.OrgSubscriptions[0]
	}

	customerMissing := sub == nil
	subscriptionMissing := sub == nil

	if !customerMissing && org != nil && org.StripeCustomerID != nil && *org.StripeCustomerID != "" {
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
	case !customerMissing && org.StripeCustomerID == nil && subscriptionMissing:
		var prices []entitlements.Price
		modules := &r.db.EntConfig.Modules
		if modules.DevMode {
			prices = internalentitlements.AllMonthlyPrices(modules.UseSandbox)
		} else {
			prices = internalentitlements.TrialMonthlyPrices(modules.UseSandbox)
		}
		if len(prices) == 0 {
			return "", ErrMissingPrice
		}

		return fmt.Sprintf("set stripe customer ID on organization and create stripe subscription with prices %v", prices), nil
	case !customerMissing && org.StripeCustomerID == nil && !subscriptionMissing:
		return "set stripe customer ID on organization", nil
	case subscriptionMissing:
		return "create stripe subscription", nil
	default:
		return "", nil
	}
}

// orgModuleConfig controls which modules are selected when creating default module records - small functional options wrapper
type orgModuleConfig struct {
	trial      bool
	allModules bool
}

// OrgModuleOption sets fields on orgModuleConfig
type OrgModuleOption func(*orgModuleConfig)

// WithTrial sets the trial flag to true, allowing trial modules to be included
func WithTrial() OrgModuleOption {
	return func(c *orgModuleConfig) {
		c.trial = true
	}
}

// WithAllModules enables all modules regardless of trial status, useful for local development
func WithAllModules() OrgModuleOption {
	return func(c *orgModuleConfig) {
		c.allModules = true
	}
}

func CreateDefaultOrgModulesProductsPrices(ctx context.Context, db *ent.Client, orgSubs *ent.OrgSubscription, orgID string, opts ...OrgModuleOption) ([]string, error) {
	cfg := orgModuleConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	modulesCreated := make([]string, 0)

	// the catalog contains config for which things should be in a trial
	if db.EntConfig == nil {
		return nil, fmt.Errorf("ent config is nil") //nolint:err113
	}

	for moduleName, mod := range gencatalog.GetModules(db.EntConfig.Modules.UseSandbox) {
		if !cfg.allModules && (!cfg.trial || !mod.IncludeWithTrial) {
			continue
		}

		// Find the first price with "month" interval
		// we want to create, by default, a monthly recurring price rather than a one-time or annual
		var monthlyPrice *catalog.Price

		for _, price := range mod.Billing.Prices {
			if price.Interval == "month" {
				monthlyPrice = &price
				break
			}
		}

		if monthlyPrice == nil {
			continue // skip if no monthly price
		}

		newCtx := contextx.With(ctx, auth.OrganizationCreationContextKey{})
		newCtx = contextx.With(newCtx, auth.OrgSubscriptionContextKey{})

		// we set the price purely for reference; it will not be used for billing - we care mostly about the association of subscription to module
		orgMod, err := db.OrgModule.Create().
			SetModule(models.OrgModule(moduleName)).
			SetSubscriptionID(orgSubs.ID).
			SetStatus(string(stripe.SubscriptionStatusTrialing)).
			SetOwnerID(orgID).
			SetModuleLookupKey(mod.LookupKey).
			SetStripePriceID(monthlyPrice.PriceID).
			SetActive(true).
			SetPrice(models.Price{Amount: float64(monthlyPrice.UnitAmount), Interval: monthlyPrice.Interval}).
			Save(newCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OrgModule for %s: %w", moduleName, err)
		}

		logx.FromContext(ctx).Debug().Msgf("created OrgModule for %s with ID %s", moduleName, orgMod.ID)

		// the product and price entries are somewhat redundant but creating them for reference and future extensibility
		orgProduct, err := db.OrgProduct.Create().
			SetModule(moduleName).
			SetOwnerID(orgID).
			SetModule(orgMod.ID).
			SetSubscriptionID(orgSubs.ID).
			SetActive(true).
			Save(newCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OrgProduct for %s: %w", moduleName, err)
		}

		logx.FromContext(ctx).Debug().Msgf("created OrgProduct for %s with ID %s", moduleName, orgProduct.ID)

		// we care mostly about which price ID we used in stripe, so we create the local reference for the price because it's the resource which dictates most of the billing toggles in stripe
		// we don't actually care that it's active or not, but it's relevant to set because we could end up with many prices on a product, and many products on a module
		orgPrice, err := db.OrgPrice.Create().
			SetProductID(orgProduct.ID).
			SetPrice(models.Price{Amount: float64(monthlyPrice.UnitAmount), Interval: monthlyPrice.Interval}).
			SetOwnerID(orgID).
			SetSubscriptionID(orgSubs.ID).
			SetStripePriceID(monthlyPrice.PriceID).
			SetActive(true).
			Save(newCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OrgPrice for module %s: %w", moduleName, err)
		}

		logx.FromContext(ctx).Debug().Msgf("created OrgPrice for %s with Stripe Price ID %s", moduleName, monthlyPrice.PriceID)

		// update the org modules with the price ID
		if _, err := db.OrgModule.UpdateOne(orgMod).SetPriceID(orgPrice.ID).Save(newCtx); err != nil {
			return nil, fmt.Errorf("failed to update OrgModule with price ID for module %s: %w", moduleName, err)
		}

		modulesCreated = append(modulesCreated, moduleName)
	}

	return modulesCreated, nil
}
