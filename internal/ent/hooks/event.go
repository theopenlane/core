package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// Eventer is a wrapper struct for having a soiree as well as a list of listeners
type Eventer struct {
	Emitter   *soiree.EventPool
	Listeners []soiree.Listener
	Topics    map[string]interface{}
}

// EventID is used to marshall and unmarshall the ID out of a ent mutation
type EventID struct {
	ID string `json:"id,omitempty"`
}

// EventerOpts is a functional options wrapper
type EventerOpts (func(*Eventer))

// NewEventer creates a new Eventer with the provided options
func NewEventer(opts ...EventerOpts) *Eventer {
	e := &Eventer{}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithEventerEmitter sets the emitter for the Eventer if there's an existing soiree pool that needs to be passed in
func WithEventerEmitter(emitter *soiree.EventPool) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

// WithEventerTopics sets the topics for the Eventer
func WithEventerTopics(topics map[string]interface{}) EventerOpts {
	return func(e *Eventer) {
		e.Topics = topics
	}
}

// WithEventerListeners takes a single topic and appends an array of listeners to the Eventer
func WithEventerListeners(topic string, listeners []soiree.Listener) EventerOpts {
	return func(e *Eventer) {
		for _, listener := range listeners {
			_, err := e.Emitter.On(topic, listener)
			if err != nil {
				log.Panic().Msg("Failed to add listener")
			}
		}
	}
}

// NewEventerPool initializes a new Eventer and takes a client to be used as the client for the soiree pool
func NewEventerPool(client interface{}) *Eventer {
	pool := soiree.NewEventPool(
		soiree.WithPool(
			soiree.NewPondPool(
				soiree.WithMaxWorkers(100), // nolint:mnd
				soiree.WithName("ent_event_pool"))),
		soiree.WithClient(client))

	return NewEventer(WithEventerEmitter(pool))
}

// parseEventID parses the event ID from the return value of an ent mutation
func parseEventID(retVal ent.Value) (*EventID, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		log.Err(err).Msg("Failed to marshal return value")
		return nil, fmt.Errorf("failed to fetch organization from subscription: %w", err)
	}

	event := EventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		log.Err(err).Msg("Failed to unmarshal return value")
		return nil, err
	}

	return &event, nil
}

// EmitEventHook emits an event to the event pool when a mutation is performed
func EmitEventHook(e *Eventer) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if reflect.TypeOf(retVal).Kind() == reflect.Int {
				log.Debug().Interface("value", retVal).Msg("mutation returned an int, skipping event emission")
				// TODO: determine if we need to emit events for mutations that return an int
				return retVal, err
			}

			emit := func() {
				eventID, err := parseEventID(retVal)
				if err != nil {
					log.Err(err).Msg("Failed to parse event ID")
					return
				}

				name := fmt.Sprintf("%s.%s", mutation.Type(), mutation.Op())
				event := soiree.NewBaseEvent(name, mutation)

				event.Properties().Set("ID", eventID.ID)

				for _, field := range mutation.Fields() {
					value, exists := mutation.Field(field)
					if exists {
						event.Properties().Set(field, value)
					}
				}

				event.SetContext(context.WithoutCancel(ctx))
				event.SetClient(e.Emitter.GetClient())
				e.Emitter.Emit(event.Topic(), event)
			}

			if tx := transactionFromContext(ctx); tx != nil {
				tx.OnCommit(func(next entgen.Committer) entgen.Committer {
					return entgen.CommitFunc(func(ctx context.Context, tx *entgen.Tx) error {
						err := next.Commit(ctx, tx)
						if err == nil {
							defer emit()
						}

						return err
					})
				})
			} else {
				defer emit()
			}

			return retVal, err
		})
	}
}

// OrganizationSettingCreate and OrganizationSettingUpdateOne are the topics for the organization setting events
var OrganizationSettingUpdateOne = "OrganizationSetting.OpUpdateOne"
var OrgSubscriptionCreate = "OrgSubscription.OpCreate"

// RegisterGlobalHooks registers global event hooks for the entdb client and expects a pointer to an Eventer
func RegisterGlobalHooks(client *entgen.Client, e *Eventer) {
	client.Use(EmitEventHook(e))
}

// RegisterListeners is currently used to globally register what listeners get applied on the entdb client
func RegisterListeners(e *Eventer) error {
	_, err := e.Emitter.On(OrgSubscriptionCreate, handleOrgSubscriptionCreated)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	// TODO(MKA): after ensuring we overwrite fields from stripe into our system and viceversa, uncomment this event listener
	//	_, err = e.Emitter.On(OrganizationSettingUpdateOne, handleOrganizationSettingsUpdateOne)
	//	if err != nil {
	//		log.Error().Err(ErrFailedToRegisterListener)
	//		return err
	//	}

	return nil
}

func handleOrgSubscriptionCreated(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		log.Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	ctx := privacy.DecisionContext(event.Context(), privacy.Allow)
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})

	orgCustomer := &entitlements.OrganizationCustomer{}
	orgSubs, err := client.OrgSubscription.Get(allowCtx, lo.ValueOr(event.Properties(), "ID", "").(string))
	if err != nil {
		log.Err(err).Msg("Failed to fetch organization subscription")

		return err
	}

	orgCustomer.OrganizationSubscriptionID = orgSubs.ID

	orgCust, err := getOrgFromSubs(allowCtx, orgSubs, client, orgCustomer)
	if err != nil {
		log.Err(err).Msg("Failed to fetch organization from subscription")

		return nil
	}

	orgCust, err = entMgr.FindOrCreateCustomer(allowCtx, orgCust)
	if err != nil {
		log.Err(err).Msg("Failed to create customer")

		return err
	}

	_, err = mapCustomerToOrgSubs(allowCtx, orgCust, client)
	if err != nil {
		log.Err(err).Msg("Failed to map customer to org subscription")

		return err
	}

	return nil
}

func mapCustomerToOrgSubs(ctx context.Context, customer *entitlements.OrganizationCustomer, client interface{}) (*entgen.OrgSubscription, error) {
	productName := ""
	productPrice := models.Price{}

	if len(customer.Prices) != 1 {
		return nil, fmt.Errorf("unable to determine product name and price, there are %v prices", len(customer.Prices))
	}

	productName = customer.Prices[0].ProductName
	productPrice.Amount = customer.Prices[0].Price
	productPrice.Currency = customer.Prices[0].Currency
	productPrice.Interval = customer.Prices[0].Interval

	expiresAt := time.Unix(0, 0)
	if customer.Subscription.EndDate != 0 {
		expiresAt = time.Unix(customer.Subscription.EndDate, 0)
	}

	active := false
	if customer.Subscription.Status == "active" || customer.Subscription.Status == "trialing" {
		active = true
	}

	if customer.OrganizationSubscriptionID == "" {
		return nil, fmt.Errorf("organization subscription ID is empty")
	}

	newOrgSubs, err := client.(*entgen.Client).OrgSubscription.UpdateOneID(customer.OrganizationSubscriptionID).
		SetStripeSubscriptionID(customer.StripeSubscriptionID).
		SetStripeCustomerID(customer.StripeCustomerID).
		SetStripeSubscriptionStatus(customer.Subscription.Status).
		SetActive(active).
		SetProductTier(productName).
		SetFeatures(customer.FeatureNames).
		SetFeatureLookupKeys(customer.Features).
		SetStripeProductTierID(customer.Subscription.ProductID).
		SetProductPrice(productPrice).
		SetExpiresAt(expiresAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return newOrgSubs, nil
}

func getOrgFromSubs(ctx context.Context, orgSubs *entgen.OrgSubscription, client interface{}, o *entitlements.OrganizationCustomer) (*entitlements.OrganizationCustomer, error) {
	if orgSubs == nil {
		return nil, fmt.Errorf("organization subscription is nil")
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSubs.OwnerID)).WithSetting().Only(ctx)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization by organization ID %s", orgSubs.OwnerID)

		return nil, err
	}

	if org.Edges.Setting != nil {
		o.OrganizationSettingsID = org.Edges.Setting.ID
	} else {
		log.Warn().Msgf("Organization setting is nil for organization ID %s", orgSubs.OwnerID)
	}

	o.OrganizationID = org.ID
	o.OrganizationName = org.Name
	o.OrganizationSettingsID = org.Edges.Setting.ID
	o.PersonalOrg = org.PersonalOrg
	o.Email = org.Edges.Setting.BillingEmail

	return o, nil
}

func getOrgSubs(ctx context.Context, orgsubsID string, client interface{}) (*entgen.OrgSubscription, error) {
	if orgsubsID == "" {
		return nil, fmt.Errorf("org subs ID is empty")
	}

	orgSubs, err := client.(*entgen.Client).OrgSubscription.Get(ctx, orgsubsID)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization subscriptions by organization ID %s", orgsubsID)

		return nil, err
	}

	return orgSubs, nil
}

// handleOrganizationSettingsUpdateOne checks for updates to the organization settings and updates the customer in Stripe accordingly
func handleOrganizationSettingsUpdateOne(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		log.Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	orgCust, err := fetchOrganizationIDbyOrgSettingID(event.Context(), lo.ValueOr(event.Properties(), "ID", "").(string), client)
	if err != nil {
		log.Err(err).Msg("Failed to fetch organization ID by organization setting ID")

		return err
	}

	// TODO(MKA): ensure all the fields which can be updated in stripe by the customer override what we store, and vice versa
	if params, hasUpdate := entitlements.CheckForBillingUpdate(event.Properties(), orgCust); hasUpdate {
		if _, err := entMgr.UpdateCustomer(orgCust.StripeCustomerID, params); err != nil {
			log.Err(err).Msg("Failed to update customer")

			return err
		}
	}

	return nil
}

// fetchOrganizationIDbyOrgSettingID fetches the organization ID by the organization setting ID
func fetchOrganizationIDbyOrgSettingID(ctx context.Context, orgsettingID string, client interface{}) (*entitlements.OrganizationCustomer, error) {
	orgSetting, err := client.(*entgen.Client).OrganizationSetting.Get(ctx, orgsettingID)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization setting ID %s", orgsettingID)

		return nil, err
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSetting.OrganizationID)).WithOrgSubscriptions().Only(ctx)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization by organization setting ID %s after 3 attempts", orgsettingID)

		return nil, err
	}

	personalOrg := org.PersonalOrg

	if len(org.Edges.OrgSubscriptions) > 1 {
		log.Warn().Str("organization_id", org.ID).Msg("organization has multiple subscriptions")

		return nil, ErrTooManySubscriptions
	}

	stripeCustomerID := ""
	if len(org.Edges.OrgSubscriptions) > 0 {
		stripeCustomerID = org.Edges.OrgSubscriptions[0].StripeCustomerID
	}

	return &entitlements.OrganizationCustomer{
		OrganizationID:         org.ID,
		OrganizationName:       org.Name,
		StripeCustomerID:       stripeCustomerID,
		OrganizationSettingsID: orgSetting.ID,
		PersonalOrg:            personalOrg,
		ContactInfo: entitlements.ContactInfo{
			Email:      orgSetting.BillingEmail,
			Phone:      orgSetting.BillingPhone,
			Line1:      &orgSetting.BillingAddress.Line1,
			Line2:      &orgSetting.BillingAddress.Line2,
			City:       &orgSetting.BillingAddress.City,
			State:      &orgSetting.BillingAddress.State,
			Country:    &orgSetting.BillingAddress.Country,
			PostalCode: &orgSetting.BillingAddress.PostalCode,
		},
	}, nil
}
