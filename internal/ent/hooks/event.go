package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v81"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type Eventer struct {
	Emitter   *soiree.EventPool
	Event     soiree.BaseEvent
	EventID   string `json:"id,omitempty"`
	HookFuncs []EventHookFunc
	EntClient *entgen.Client
}

type EventerOpts (func(*Eventer))

type EventHookFunc func(ctx context.Context, m ent.Mutation) error

func NewEventer(opts ...EventerOpts) *Eventer {
	e := &Eventer{}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func WithEventerEmitter(emitter *soiree.EventPool) EventerOpts {
	return func(e *Eventer) {
		e.Emitter = emitter
	}
}

// InitEventPool initializes an event pool with a client and a error handler
func NewEventerPool(client interface{}) *Eventer {
	pool := soiree.NewEventPool(
		soiree.WithPool(
			soiree.NewPondPool(
				soiree.WithMaxWorkers(100), // noling:mnd
				soiree.WithName("ent_event_pool"))),
		soiree.WithClient(client))

	return NewEventer(WithEventerEmitter(pool))
}

func (e *Eventer) AddHookFunc(fn EventHookFunc) {
	e.HookFuncs = append(e.HookFuncs, fn)
}

type EventID struct {
	ID string `json:"id,omitempty"`
}

func parseEventID(retVal ent.Value) (*EventID, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		log.Err(err).Msg("Failed to marshal return value")
		return nil, err
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

				log.Debug().Msg("base event created with topic name" + name)

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
				log.Debug().Msg("event emitted")
			}

			if tx := entgen.TxFromContext(ctx); tx != nil {
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

func RegisterGlobalHooks(client *entgen.Client, e *Eventer) {
	client.Use(EmitEventHook(e))
}

// OrganizationSettingCreated and OrganizationSettingUpdated are the topics for OrganizationSetting creation and update events
var OrganizationSettingCreated = "OrganizationSetting.OpCreate"
var OrganizationSettingUpdated = "OrganizationSetting.OpUpdateOne"

type EntitlementEvent struct {
	OpType          string
	listener        []soiree.Listener
	client          interface{}
	customerwrapper *entitlements.OrganizationCustomer
}

func mapEntitlementEvent(e *EntitlementEvent) soiree.Listener {
	return func(event soiree.Event) error {
		client := event.Client().(*entgen.Client)

		if client.EntitlementManager == nil {
			log.Info().Msg("EntitlementManager not found on client, skipping customer creation")

			return nil
		}

		props := event.Properties()

		org, err := fetchOrganizationIDbyOrgSettingID(event.Context(), props["ID"].(string), client)
		if err != nil {
			log.Err(err).Msg("Failed to fetch organization by organization setting ID")
			return err
		}

		e.customerwrapper = &entitlements.OrganizationCustomer{
			OrganizationID:   org.ID,
			StripeCustomerID: props["stripe_id"].(string),
			BillingEmail:     props["billing_email"].(string),
			BillingPhone:     props["billing_phone"].(string),
			OrganizationName: props["billing_address"].(string),
		}

		if e.OpType == OrganizationSettingCreated {
			if err := handleCustomerCreate(event); err != nil {
				log.Err(err).Msg("Failed to handle customer creation")
				return
			}

			return nil
		}

		if e.OpType == OrganizationSettingUpdated {
			if err := handleCustomerUpdate(event); err != nil {
				log.Err(err).Msg("Failed to handle customer update")
				return
			}

			return nil
		}

		return nil
	}
}

// handleCustomerCreate handles the creation of a customer in Stripe when an OrganizationSetting is created or updated
func handleCustomerCreate(event soiree.Event) error {
	client := event.Client().(*entgen.Client)

	if client.EntitlementManager == nil {
		log.Info().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	props := event.Properties()
	// TODO: add funcs for default unmarshalling of props and send all props to stripe
	orgSettingID, exists := props["ID"]
	if !exists {
		log.Info().Msg("organizationSetting ID not found in event properties")
		return nil
	}

	// if stripe ID is being added to the properties, we enumerate through any other fields in the update and ensure they match
	stripeID, exists := props["stripe_id"]
	if exists && stripeID != "" {
		stripes := stripeID.(string)

		stripeCustomer, err := client.EntitlementManager.Client.Customers.Get(stripes, nil)
		if err != nil {
			log.Err(err).Msg("Failed to retrieve Stripe customer by ID")
			return err
		}

		// check for updates to billing information and update the Stripe customer if necessary
		if params, hasUpdate := checkForBillingUpdate(props, stripeCustomer); hasUpdate {
			if _, err := client.EntitlementManager.Client.Customers.Update(stripes, params); err != nil {
				log.Err(err).Interface("params", params).Msg("failed to update Stripe customer with new billing information")

				return err
			}
		}
	}

	billingEmail, exists := props["billing_email"]
	if exists && billingEmail != "" {
		email := billingEmail.(string)

		i := client.EntitlementManager.Client.Customers.List(&stripe.CustomerListParams{Email: &email})

		var customerID string

		if i.Next() {
			customerID = i.Customer().ID
		} else {
			// if there is no customer with the email, create one
			customer, err := client.EntitlementManager.CreateCustomer(email)
			if err != nil {
				log.Err(err).Msg("Failed to create Stripe customer")
				return err
			}

			customerID = customer.ID

			log.Debug().Msgf("Created Stripe customer with ID: %s", customer.ID)

			if err := updateOrganizationSettingWithCustomerID(event.Context(), orgSettingID.(string), customer.ID, event.Client()); err != nil {
				log.Err(err).Msg("Failed to update OrganizationSetting with Stripe customer ID")
				return err
			}

			log.Debug().Msgf("Updated OrganizationSetting with Stripe customer ID: %s", customer.ID)
		}

		// TODO create ent db records / corresponding feature / plan records
		subs, err := client.EntitlementManager.ListOrCreateSubscriptions(customerID)
		if err != nil {
			log.Err(err).Msg("failed to list or create Stripe subscriptions")
			return err
		}

		checkout, err := client.EntitlementManager.CreateBillingPortalUpdateSession(subs.ID, customerID)
		if err != nil {
			log.Err(err).Msg("failed to create billing portal update session")
			return err
		}

		log.Debug().Msgf("Created billing portal update session with URL %s", checkout.URL)

		if err := updateOrganizationSettingWithCustomerID(event.Context(), orgSettingID.(string), customerID, event.Client()); err != nil {
			log.Err(err).Msg("Failed to update OrganizationSetting with Stripe customer ID")
			return err
		}

		log.Debug().Msgf("Updated OrganizationSetting with Stripe customer ID: %s", customerID)
	}

	return nil
}

// updateOrganizationSettingWithCustomerID updates an OrganizationSetting with a Stripe customer ID
func updateOrganizationSettingWithCustomerID(ctx context.Context, orgsettingID, customerID string, client interface{}) error {
	if _, err := client.(*entgen.Client).OrganizationSetting.UpdateOneID(orgsettingID).SetStripeID(customerID).Save(ctx); err != nil {
		log.Err(err).Msgf("Failed to update OrganizationSetting ID %s with Stripe customer ID %s", orgsettingID, customerID)

		return err
	}

	return nil
}

// fetchOrganizationIDbyOrgSettingID fetches the organization ID by the organization setting ID
func fetchOrganizationIDbyOrgSettingID(ctx context.Context, orgsettingID string, client interface{}) (*entitlements.OrganizationCustomer, error) {
	orgSetting, err := client.(*entgen.Client).OrganizationSetting.Get(ctx, orgsettingID)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization setting by ID %s", orgsettingID)
		return nil, err
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSetting.OrganizationID)).Only(ctx)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization by organization setting ID %s", orgsettingID)
		return nil, err
	}

	return &entitlements.OrganizationCustomer{
		OrganizationID:         org.ID,
		StripeCustomerID:       orgSetting.StripeID,
		BillingEmail:           orgSetting.BillingEmail,
		BillingPhone:           orgSetting.BillingPhone,
		OrganizationName:       org.Name,
		OrganizationSettingsID: orgSetting.ID,
	}, nil
}

// checkForBillingUpdate checks for updates to billing information in the properties and returns a stripe.CustomerParams object with the updated information
// and a boolean indicating whether there are updates
func checkForBillingUpdate(props map[string]interface{}, stripeCustomer *stripe.Customer) (params *stripe.CustomerParams, hasUpdate bool) {
	params = &stripe.CustomerParams{}

	billingEmail, exists := props["billing_email"]
	if exists && billingEmail != "" {
		email := billingEmail.(string)
		if stripeCustomer.Email != email {
			params.Email = &email
			hasUpdate = true
		}
	}

	billingPhone, exists := props["billing_phone"]
	if exists && billingPhone != "" {
		phone := billingPhone.(string)
		if stripeCustomer.Phone != phone {
			params.Phone = &phone
			hasUpdate = true
		}
	}

	// TODO: split out address fields in ent schema
	billingAddress, exists := props["billing_address"]
	if exists && billingAddress != "" {
		address := billingAddress.(string)
		if stripeCustomer.Address.Line1 != address {
			params.Address = &stripe.AddressParams{
				Line1: &address,
			}
			hasUpdate = true
		}
	}

	return
}

func HasFields(m ent.Mutation, err error, fields ...string) bool {
	_, ok := lo.Find(fields, func(field string) bool {
		_, ok := m.Field(field)
		return ok
	})

	return ok
}
