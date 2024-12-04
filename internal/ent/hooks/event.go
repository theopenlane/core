package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type Eventer struct {
	Emitter   *soiree.EventPool
	EventID   string `json:"id,omitempty"`
	Listeners []soiree.Listener
}

type EventID struct {
	ID string `json:"id,omitempty"`
}

type EventerOpts (func(*Eventer))

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

func RegisterListeners(e *Eventer) error {
	for _, event := range []string{"OrganizationSetting.OpCreate", "OrganizationSetting.OpUpdateOne"} {
		_, err := e.Emitter.On(event, handleOrganizationSettingEvents, soiree.WithRetry(3))
		if err != nil {
			log.Error().Err(ErrFailedToRegisterListener)
			return err
		}

		log.Debug().Msg("Listener registered for " + event)

	}

	return nil
}

// handleCustomerCreate handles the creation of a customer in Stripe when an OrganizationSetting is created or updated
func handleOrganizationSettingEvents(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		log.Info().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	props := event.Properties()

	orgSettingID, exists := props["ID"]
	if !exists {
		log.Info().Msg("organizationSetting ID not found in event properties")
		return nil
	}

	orgCust, err := fetchOrganizationIDbyOrgSettingID(event.Context(), orgSettingID.(string), client)
	if err != nil {
		log.Err(err).Msg("Failed to fetch organization ID by organization setting ID")

		return err
	}

	if orgCust.StripeCustomerID == "" {
		customer, err := entMgr.FindorCreateCustomer(event.Context(), orgCust)
		if err != nil {
			log.Err(err).Msg("Failed to create customer")

			return err
		}

		if err := updateOrganizationSettingWithCustomerID(event.Context(), orgCust.OrganizationSettingsID, customer.StripeCustomerID, client); err != nil {
			log.Err(err).Msg("Failed to update organization setting with customer ID")

			return err
		}

		return nil
	} else {
		if params, hasUpdate := checkForBillingUpdate(event.Properties(), orgCust); hasUpdate {
			if _, err := entMgr.UpdateCustomer(orgCust.StripeCustomerID, params); err != nil {
				log.Err(err).Msg("Failed to update customer")

				return err
			}
		}
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
	var orgSetting *entgen.OrganizationSetting

	var err error

	// Retry fetching organization setting
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second) // nolint:mnd

		orgSetting, err = client.(*entgen.Client).OrganizationSetting.Get(ctx, orgsettingID)
		if err == nil {
			break
		}

		log.Warn().Msgf("Retrying fetch organization setting by ID %s, attempt %d", orgsettingID, i+1)

	}

	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization setting by ID %s after 3 attempts", orgsettingID)
		return nil, err
	}

	var org *entgen.Organization

	// Retry fetching organization
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second) // nolint:mnd

		org, err = client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSetting.OrganizationID)).Only(ctx)
		if err == nil {
			break
		}

		log.Warn().Msgf("Retrying fetch organization by organization setting ID %s, attempt %d", orgsettingID, i+1)

	}

	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization by organization setting ID %s after 3 attempts", orgsettingID)
		return nil, err
	}

	//if !org.PersonalOrg {
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
func checkForBillingUpdate(props map[string]interface{}, stripeCustomer *entitlements.OrganizationCustomer) (params *stripe.CustomerParams, hasUpdate bool) {
	params = &stripe.CustomerParams{}

	billingEmail, exists := props["billing_email"]
	if exists && billingEmail != "" {
		email := billingEmail.(string)
		if stripeCustomer.BillingEmail != email {
			params.Email = &email
			hasUpdate = true
		}
	}

	billingPhone, exists := props["billing_phone"]
	if exists && billingPhone != "" {
		phone := billingPhone.(string)
		if stripeCustomer.BillingPhone != phone {
			params.Phone = &phone
			hasUpdate = true
		}
	}

	return params, hasUpdate
}
