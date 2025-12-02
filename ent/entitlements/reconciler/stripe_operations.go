package reconciler

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/shared/catalog"
	"github.com/theopenlane/shared/entitlements"
)

const (
	defaultStripePageLimit = 100
)

var stripeSubsStatuses = []stripe.SubscriptionStatus{
	stripe.SubscriptionStatusActive,
	stripe.SubscriptionStatusTrialing,
	stripe.SubscriptionStatusPastDue,
	stripe.SubscriptionStatusUnpaid,
}

// IsSubscriptionActiveOrTrialing checks if subscription is in active or trialing status
func IsSubscriptionActiveOrTrialing(status stripe.SubscriptionStatus) bool {
	return status == stripe.SubscriptionStatusActive || status == stripe.SubscriptionStatusTrialing
}

// ShouldUpdateCancelBehavior determines if a subscription needs cancel behavior update from pause to cancel
func ShouldUpdateCancelBehavior(sub *stripe.Subscription) bool {
	return sub.TrialSettings != nil &&
		sub.TrialSettings.EndBehavior != nil &&
		sub.TrialSettings.EndBehavior.MissingPaymentMethod == stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause
}

// CancelBehaviorResult contains the results of cancel behavior update operations
type CancelBehaviorResult struct {
	Actions []actionRow
}

// UpdateSubscriptionsCancelBehavior updates all active/trialing subscriptions with pause trial end behavior to cancel behavior
func (r *Reconciler) UpdateSubscriptionsCancelBehavior(ctx context.Context, orgIDs []string) (*CancelBehaviorResult, error) {
	if r.dryRun {
		log.Info().Msg("analyzing subscription cancel behavior from pause to cancel (DRY RUN)")
	} else {
		log.Info().Msg("updating subscription cancel behavior from pause to cancel")
	}

	var (
		updateCount int
		rows        []actionRow
	)

	for subs := range stripeSubsStatuses {
		it := r.stripe.Client.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
			ListParams: stripe.ListParams{
				Limit:  stripe.Int64(defaultStripePageLimit),
				Expand: []*string{stripe.String("data.schedule")},
			},
			Status: stripe.String(string(stripeSubsStatuses[subs])),
		})

		for sub, err := range it {
			if err != nil {
				return nil, fmt.Errorf("listing subscriptions: %w", err)
			}

			// If orgIDs filter is provided, skip subscriptions not in the list
			if len(orgIDs) > 0 && !slices.Contains(orgIDs, sub.ID) {
				continue
			}

			// Additional safety check for subscription status
			if !IsSubscriptionActiveOrTrialing(sub.Status) {
				continue
			}

			if !ShouldUpdateCancelBehavior(sub) {
				continue
			}

			if r.dryRun {
				orgName := ""
				if sub.Metadata != nil {
					orgName = sub.Metadata["organization_name"]
				}

				rows = append(rows, actionRow{OrgID: sub.ID, OrgName: orgName, Action: "update cancel behavior from pause to cancel"})

				continue
			}

			updateParams := &stripe.SubscriptionUpdateParams{
				TrialSettings: &stripe.SubscriptionUpdateTrialSettingsParams{
					EndBehavior: &stripe.SubscriptionUpdateTrialSettingsEndBehaviorParams{
						MissingPaymentMethod: stripe.String(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodCancel),
					},
				},
			}

			if _, err := r.stripe.UpdateSubscription(ctx, sub.ID, updateParams); err != nil {
				log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to update subscription")
				continue
			}

			updateCount++

			log.Info().Str("subscription_id", sub.ID).Msg("updated subscription cancel behavior")
		}
	}

	if r.dryRun {
		log.Info().Int("subscriptions_found", len(rows)).Msg("completed analyzing subscription cancel behaviors (DRY RUN)")
	} else {
		log.Info().Int("updated_count", updateCount).Msg("completed updating subscription cancel behaviors")
	}

	return &CancelBehaviorResult{Actions: rows}, nil
}

// ShouldCreateSchedule determines if a subscription needs a subscription schedule created
func ShouldCreateSchedule(sub *stripe.Subscription) bool {
	return sub.Schedule == nil || sub.Schedule.ID == ""
}

// GenerateScheduleActionDescription generates an action description for creating a subscription schedule and returns action text and customer ID
func GenerateScheduleActionDescription(sub *stripe.Subscription) (string, string) {
	customerID := ""

	action := fmt.Sprintf("create subscription schedule for %s", sub.ID)
	if sub.Customer != nil && sub.Customer.ID != "" {
		customerID = sub.Customer.ID
		action += fmt.Sprintf(" and update customer metadata for %s", customerID)
	}

	return action, customerID
}

// ScheduleResult contains the results of subscription schedule creation operations
type ScheduleResult struct {
	Actions []actionRow
}

// CreateMissingSubscriptionSchedules creates subscription schedules for active/trialing subscriptions that don't have them
func (r *Reconciler) CreateMissingSubscriptionSchedules(ctx context.Context, orgIDs []string) (*ScheduleResult, error) {
	if r.dryRun {
		log.Info().Msg("analyzing subscription schedules for subscriptions without schedules (DRY RUN)")
	} else {
		log.Info().Msg("creating subscription schedules for subscriptions without schedules")
	}

	var (
		createCount int
		rows        []actionRow
	)

	for subs := range stripeSubsStatuses {
		it := r.stripe.Client.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
			ListParams: stripe.ListParams{
				Limit:  stripe.Int64(defaultStripePageLimit),
				Expand: []*string{stripe.String("data.schedule")},
			},
			Status: stripe.String(string(stripeSubsStatuses[subs])),
		})
		for sub, err := range it {
			if err != nil {
				return nil, fmt.Errorf("listing subscriptions: %w", err)
			}

			// If orgIDs filter is provided, skip subscriptions not in the list
			if len(orgIDs) > 0 && !slices.Contains(orgIDs, sub.ID) {
				continue
			}

			// Additional safety check for subscription status
			if !IsSubscriptionActiveOrTrialing(sub.Status) {
				continue
			}

			if !ShouldCreateSchedule(sub) {
				continue
			}

			action, customerID := GenerateScheduleActionDescription(sub)

			if r.dryRun {
				orgName := ""
				if sub.Metadata != nil {
					orgName = sub.Metadata["organization_name"]
				}

				rows = append(rows, actionRow{OrgID: sub.ID, OrgName: orgName, Action: action})

				continue
			}

			// Create subscription schedule from existing subscription
			schedule, err := r.stripe.CreateSubscriptionScheduleFromSubs(ctx, sub.ID)
			if err != nil {
				log.Error().Err(err).Str("subscription_id", sub.ID).Msg("failed to create subscription schedule")
				continue
			}

			createCount++

			log.Info().Str("subscription_id", sub.ID).Str("schedule_id", schedule.ID).Msg("created subscription schedule")

			// Update customer metadata with schedule ID
			if customerID != "" {
				updateParams := r.stripe.UpdateCustomerWithOptions(
					&stripe.CustomerUpdateParams{},
					entitlements.WithUpdateCustomerMetadata(map[string]string{
						"subscription_schedule_id": schedule.ID,
					}),
				)

				if _, err := r.stripe.UpdateCustomer(ctx, customerID, updateParams); err != nil {
					log.Error().Err(err).Str("customer_id", customerID).Msg("failed to update customer metadata")
				}
			}
		}
	}

	if r.dryRun {
		log.Info().Int("schedules_needed", len(rows)).Msg("completed analyzing subscription schedules (DRY RUN)")
	} else {
		log.Info().Int("created_count", createCount).Msg("completed creating subscription schedules")
	}

	return &ScheduleResult{Actions: rows}, nil
}

// BuildValidProductsMap builds a map of valid product IDs from the provided catalog for lookup operations
func BuildValidProductsMap(cat *catalog.Catalog) map[string]bool {
	validProducts := make(map[string]bool)

	for _, module := range cat.Modules {
		if module.ProductID != "" {
			validProducts[module.ProductID] = true
		}
	}

	for _, addon := range cat.Addons {
		if addon.ProductID != "" {
			validProducts[addon.ProductID] = true
		}
	}

	return validProducts
}

// FindMissingProductsInSubscription checks a subscription for products not in the valid products map and returns a report
func FindMissingProductsInSubscription(sub *stripe.Subscription, validProducts map[string]bool) []SubscriptionProductReport {
	var report []SubscriptionProductReport

	// Skip inactive subscriptions
	if !entitlements.IsSubscriptionActive(sub.Status) {
		return report
	}

	customerID := ""
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	orgID := ""
	if sub.Metadata != nil {
		orgID = sub.Metadata["organization_id"]
	}

	// Check each subscription item for invalid products
	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.Product != nil {
			productID := item.Price.Product.ID
			if !validProducts[productID] {
				report = append(report, SubscriptionProductReport{
					SubscriptionID: sub.ID,
					CustomerID:     customerID,
					ProductID:      productID,
					ProductName:    item.Price.Product.Name,
					Status:         string(sub.Status),
					OrganizationID: orgID,
				})
			}
		}
	}

	return report
}

// ReportSubscriptionsWithMissingProducts generates a comprehensive report of active subscriptions with products not found in the catalog
func (r *Reconciler) ReportSubscriptionsWithMissingProducts(ctx context.Context, catalogPath string) ([]SubscriptionProductReport, error) {
	log.Info().Msg("generating report of subscriptions with products not in catalog")

	// Load catalog
	cat, err := catalog.LoadCatalog(catalogPath)
	if err != nil {
		return nil, fmt.Errorf("loading catalog: %w", err)
	}

	validProducts := BuildValidProductsMap(cat)

	var report []SubscriptionProductReport

	it := r.stripe.Client.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(defaultStripePageLimit),
			// Cannot expand too deep due to Stripe limits
		},
	})

	for sub, err := range it {
		if err != nil {
			return nil, fmt.Errorf("listing subscriptions: %w", err)
		}

		// Skip inactive subscriptions
		if !entitlements.IsSubscriptionActive(sub.Status) {
			continue
		}

		// For each subscription item, check if product is in catalog
		for _, item := range sub.Items.Data {
			if item.Price != nil && item.Price.ID != "" {
				// Fetch price with product data
				priceObj, err := r.stripe.Client.V1Prices.Retrieve(ctx, item.Price.ID, &stripe.PriceRetrieveParams{
					Params: stripe.Params{
						Expand: []*string{stripe.String("product")},
					},
				})
				if err != nil || priceObj == nil || priceObj.Product == nil {
					continue
				}

				productID := priceObj.Product.ID
				if !validProducts[productID] {
					customerID := ""
					if sub.Customer != nil {
						customerID = sub.Customer.ID
					}

					orgID := ""
					if sub.Metadata != nil {
						orgID = sub.Metadata["organization_id"]
					}

					report = append(report, SubscriptionProductReport{
						SubscriptionID: sub.ID,
						CustomerID:     customerID,
						ProductID:      productID,
						ProductName:    priceObj.Product.Name,
						Status:         string(sub.Status),
						OrganizationID: orgID,
					})
				}
			}
		}
	}

	log.Info().Int("missing_products_count", len(report)).Msg("completed product catalog report")

	return report, nil
}

// SubscriptionProductReport represents a subscription with products not found in the catalog
type SubscriptionProductReport struct {
	SubscriptionID string
	CustomerID     string
	ProductID      string
	ProductName    string
	Status         string
	OrganizationID string
}

// ShouldUpdateCustomerPersonalOrgMetadata determines if a customer needs personal_org metadata update to mark it as a personal organization
func ShouldUpdateCustomerPersonalOrgMetadata(customer *stripe.Customer) bool {
	return customer.Metadata == nil || customer.Metadata["personal_org"] != "true"
}

// StripeSystemMismatchReport represents data mismatches between Stripe customers and the internal system
type StripeSystemMismatchReport struct {
	CustomerID         string `json:"customer_id"`
	OrganizationID     string `json:"organization_id"`
	OrganizationName   string `json:"organization_name,omitempty"`
	MismatchType       string `json:"mismatch_type"`
	Description        string `json:"description"`
	StripeData         string `json:"stripe_data,omitempty"`
	InternalData       string `json:"internal_data,omitempty"`
	SubscriptionIssues []any  `json:"subscription_issues,omitempty"`
}

// AnalyzeStripeSystemMismatches performs comprehensive analysis of Stripe customers vs internal system data, optionally executing cleanup actions
func (r *Reconciler) AnalyzeStripeSystemMismatches(ctx context.Context, action string) ([]StripeSystemMismatchReport, error) {
	if r.db == nil {
		return nil, ErrMissingDBClient
	}

	if action == "cleanup-orphaned-customers" {
		_, err := r.CleanupOrphanedStripeCustomers(ctx)
		if err != nil {
			return nil, err
		}
		// Convert to correct return type - action was performed so return empty report
		var reports []StripeSystemMismatchReport

		return reports, nil
	}

	log.Info().Msg("analyzing Stripe customers vs internal system for mismatches")

	var report []StripeSystemMismatchReport

	// Get all organizations from internal system for lookup
	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)

	orgs, err := r.db.Organization.Query().
		WithOrgSubscriptions().
		WithSetting().
		Where(
			organization.And(
				organization.DeletedAtIsNil(),
				organization.Not(organization.ID("01101101011010010111010001100010")),
				organization.PersonalOrg(false),
			),
		).
		All(internalCtx)
	if err != nil {
		return nil, fmt.Errorf("query organizations: %w", err)
	}

	// Build lookup maps for internal data
	orgMap := make(map[string]bool)
	customerIDMap := make(map[string]string)

	for _, org := range orgs {
		orgMap[org.ID] = true
		if org.StripeCustomerID != nil {
			customerIDMap[*org.StripeCustomerID] = org.ID
		}
	}

	// List all Stripe customers
	customerIt := r.stripe.Client.V1Customers.List(ctx, &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(defaultStripePageLimit),
		}})

	for customer, err := range customerIt {
		if err != nil {
			return nil, fmt.Errorf("listing Stripe customers: %w", err)
		}

		// Extract organization ID from metadata
		orgID := ""
		orgName := ""

		if customer.Metadata != nil {
			orgID = customer.Metadata["organization_id"]
			orgName = customer.Metadata["organization_name"]
		}

		if orgID == "" {
			// Customer has no organization ID metadata
			report = append(report, StripeSystemMismatchReport{
				CustomerID:   customer.ID,
				MismatchType: "missing_organization_metadata",
				Description:  "Stripe customer has no organization_id in metadata",
				StripeData:   fmt.Sprintf("Customer: %s, Email: %s", customer.ID, customer.Email),
			})

			continue
		}

		// check if it's a personal org customer
		org, err := r.db.Organization.Query().Where(organization.ID(orgID)).Select(organization.FieldPersonalOrg).Only(internalCtx)
		if err != nil {
			if !generated.IsNotFound(err) {
				log.Error().Err(err).Str("org_id", orgID).Msg("failed to check if organization is personal")
				continue
			}
		}

		if org != nil && org.PersonalOrg {
			continue
		}

		// Check if organization exists in internal system
		if !orgMap[orgID] {
			report = append(report, StripeSystemMismatchReport{
				CustomerID:       customer.ID,
				OrganizationID:   orgID,
				OrganizationName: orgName,
				MismatchType:     "organization_not_found",
				Description:      "Organization ID from Stripe metadata not found in internal system",
				StripeData:       fmt.Sprintf("Customer: %s, OrgID: %s", customer.ID, orgID),
			})

			continue
		}

		// Check if internal organization points to this customer
		if expectedOrgID, exists := customerIDMap[customer.ID]; !exists || expectedOrgID != orgID {
			report = append(report, StripeSystemMismatchReport{
				CustomerID:       customer.ID,
				OrganizationID:   orgID,
				OrganizationName: orgName,
				MismatchType:     "customer_id_mismatch",
				Description:      "Internal organization points to different Stripe customer ID",
				StripeData:       fmt.Sprintf("Customer: %s", customer.ID),
				InternalData:     fmt.Sprintf("Expected OrgID: %s", expectedOrgID),
			})
		}
	}

	log.Info().Int("mismatch_count", len(report)).Msg("completed Stripe system mismatch analysis")

	return report, nil
}

// CleanupResult contains the results of cleanup operations
type CleanupResult struct {
	Actions []actionRow
}

// CleanupOrphanedStripeCustomers removes Stripe customers that don't exist in the internal system or have invalid references
func (r *Reconciler) CleanupOrphanedStripeCustomers(ctx context.Context) (*CleanupResult, error) {
	if r.db == nil {
		return nil, ErrMissingDBClient
	}

	if r.dryRun {
		log.Info().Msg("analyzing orphaned Stripe customers for cleanup (DRY RUN)")
	} else {
		log.Info().Msg("cleaning up orphaned Stripe customers")
	}

	// Get all organizations from internal system for lookup
	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)

	orgs, err := r.db.Organization.Query().
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

	// Build lookup map for internal organizations
	orgMap := make(map[string]bool)
	customerIDMap := make(map[string]bool)

	for _, org := range orgs {
		orgMap[org.ID] = true
		if org.StripeCustomerID != nil {
			customerIDMap[*org.StripeCustomerID] = true
		}
	}

	var (
		deleteCount int
		rows        []actionRow
	)

	// List all Stripe customers
	customerIt := r.stripe.Client.V1Customers.List(ctx, &stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(defaultStripePageLimit),
		}})

	for customer, err := range customerIt {
		if err != nil {
			return nil, fmt.Errorf("listing Stripe customers: %w", err)
		}

		shouldDelete := false
		deleteReason := ""

		// Check if customer has organization metadata
		orgID := ""
		orgName := ""

		if customer.Metadata != nil {
			orgID = customer.Metadata["organization_id"]
			orgName = customer.Metadata["organization_name"]
		}

		switch {
		case orgID == "":
			shouldDelete = true
			deleteReason = "no organization_id metadata"
		case !orgMap[orgID]:
			shouldDelete = true
			deleteReason = fmt.Sprintf("organization %s not found in internal system", orgID)
		case !customerIDMap[customer.ID]:
			shouldDelete = true
			deleteReason = "customer ID not referenced by any internal organization"
		}

		if !shouldDelete {
			continue
		}

		action := fmt.Sprintf("delete orphaned customer %s (%s)", customer.ID, deleteReason)

		if r.dryRun {
			rows = append(rows, actionRow{OrgID: orgID, OrgName: orgName, Action: action})
			continue
		}

		// Check if customer has active subscriptions before deleting
		subscriptionIt := r.stripe.Client.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
			Customer: stripe.String(customer.ID),
			Status:   stripe.String("active"),
			ListParams: stripe.ListParams{
				Limit: stripe.Int64(1),
			},
		})

		hasActiveSubscriptions := false

		for _, err := range subscriptionIt {
			if err != nil {
				log.Error().Err(err).Str("customer_id", customer.ID).Msg("failed to check customer subscriptions")
				break
			}

			hasActiveSubscriptions = true

			break
		}

		if hasActiveSubscriptions {
			log.Warn().Str("customer_id", customer.ID).Msg("skipping customer deletion - has active subscriptions")
			continue
		}

		// Delete the customer
		if _, err := r.stripe.Client.V1Customers.Delete(ctx, customer.ID, &stripe.CustomerDeleteParams{}); err != nil {
			log.Error().Err(err).Str("customer_id", customer.ID).Msg("failed to delete customer")
			continue
		}

		deleteCount++

		log.Info().Str("customer_id", customer.ID).Str("reason", deleteReason).Msg("deleted orphaned customer")
	}

	if r.dryRun {
		log.Info().Int("customers_to_delete", len(rows)).Msg("completed analyzing orphaned customers (DRY RUN)")
	} else {
		log.Info().Int("deleted_count", deleteCount).Msg("completed cleaning up orphaned customers")
	}

	return &CleanupResult{Actions: rows}, nil
}

// MetadataUpdateResult contains the results of customer metadata update operations
type MetadataUpdateResult struct {
	Actions []actionRow
}

// UpdatePersonalOrgMetadata updates Stripe customer metadata to mark personal organizations with the personal_org flag
func (r *Reconciler) UpdatePersonalOrgMetadata(ctx context.Context) (*MetadataUpdateResult, error) {
	if r.db == nil {
		return nil, ErrMissingDBClient
	}

	if r.dryRun {
		log.Info().Msg("analyzing personal organization metadata (DRY RUN)")
	} else {
		log.Info().Msg("updating personal organization metadata")
	}

	// Add internal context for administrative operations
	internalCtx := rule.WithInternalContext(ctx)

	orgs, err := r.db.Organization.Query().
		WithOrgSubscriptions().
		WithSetting().
		Where(
			organization.And(
				organization.PersonalOrg(true),
				organization.DeletedAtIsNil(),
			),
		).
		All(internalCtx)
	if err != nil {
		return nil, fmt.Errorf("query personal organizations: %w", err)
	}

	var (
		updateCount int
		rows        []actionRow
	)

	for _, org := range orgs {
		if org.StripeCustomerID == nil || *org.StripeCustomerID == "" {
			continue
		}

		// Retrieve current customer to check existing metadata
		customer, err := r.stripe.GetCustomerByStripeID(ctx, *org.StripeCustomerID)
		if err != nil {
			log.Error().Err(err).Str("customer_id", *org.StripeCustomerID).Msg("failed to get customer")
			continue
		}

		if !ShouldUpdateCustomerPersonalOrgMetadata(customer) {
			continue
		}

		if r.dryRun {
			action := fmt.Sprintf("add personal_org metadata to customer %s", *org.StripeCustomerID)
			rows = append(rows, actionRow{OrgID: org.ID, OrgName: org.DisplayName, Action: action})

			continue
		}

		// Update metadata
		metadata := make(map[string]string)
		if customer.Metadata != nil {
			maps.Copy(metadata, customer.Metadata)
		}

		metadata["personal_org"] = "true"

		updateParams := r.stripe.UpdateCustomerWithOptions(
			&stripe.CustomerUpdateParams{},
			entitlements.WithUpdateCustomerMetadata(metadata),
		)

		if _, err := r.stripe.UpdateCustomer(ctx, *org.StripeCustomerID, updateParams); err != nil {
			log.Error().Err(err).Str("customer_id", *org.StripeCustomerID).Msg("failed to update customer metadata")
			continue
		}

		updateCount++

		log.Info().Str("org_id", org.ID).Str("customer_id", *org.StripeCustomerID).Msg("updated personal org metadata")
	}

	if r.dryRun {
		log.Info().Int("customers_needing_update", len(rows)).Msg("completed analyzing personal org metadata (DRY RUN)")
	} else {
		log.Info().Int("updated_count", updateCount).Msg("completed updating personal org metadata")
	}

	return &MetadataUpdateResult{Actions: rows}, nil
}
