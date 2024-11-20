package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// EventID is a struct to hold the ID of an event
type EventID struct {
	ID string `json:"id,omitempty"`
}

// EmitEventHook emits an event to the event pool when a mutation is performed
func EmitEventHook(pool *soiree.EventPool) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			// if the return value is an int, it's a count of the number of rows affected
			// and we don't want to emit an event for that
			if reflect.TypeOf(retVal).Kind() == reflect.Int {
				log.Debug().Interface("value", retVal).Msg("mutation returned an int, skipping event emission")

				return retVal, err
			}

			op := mutation.Op()
			typ := mutation.Type()
			fields := mutation.Fields()

			out, err := json.Marshal(retVal)
			if err != nil {
				log.Err(err).Msg("Failed to marshal return value")
			}

			other := EventID{}
			if err := json.Unmarshal(out, &other); err != nil {
				log.Err(err).Msg("Failed to unmarshal return value")
			}

			// create a new event unfilitered for every mutation
			// this pushes events into the pool but they are only actioned against if there is a listener
			event := soiree.NewBaseEvent(fmt.Sprintf("%s.%s", typ, op), mutation)
			event.Properties().Set("ID", other.ID)

			for _, field := range fields {
				value, exists := mutation.Field(field)
				if exists {
					event.Properties().Set(field, value)
				}
			}

			event.SetContext(context.WithoutCancel(ctx))
			event.SetClient(pool.GetClient())
			pool.Emit(event.Topic(), event)

			return retVal, err
		})
	}
}

// RegisterGlobalHooks registers global hooks for the ent client with the initialized event pool
func RegisterGlobalHooks(client *entgen.Client, pool *soiree.EventPool) {
	client.Use(EmitEventHook(pool))
}

// InitEventPool initializes an event pool with a client and a error handler
func InitEventPool(client interface{}) *soiree.EventPool {
	return soiree.NewEventPool(soiree.WithPool(soiree.NewPondPool(soiree.WithMaxWorkers(100), soiree.WithName("ent_event_pool"))), soiree.WithErrorHandler(func(event soiree.Event, err error) error { // nolint: mnd
		log.Printf("Error encountered during event '%s': %v, with payload: %v", event.Topic(), err, event.Payload())
		return nil
	}), soiree.WithClient(client))
}

// TODO: template out listeners for all mutation types and create hooks that allow functions to be easily called from within the codebase to perform actions on those events
// RegisterListeners registers listeners for events on the event pool and is registered with the dbclient
func RegisterListeners(pool *soiree.EventPool) {
	_, err := pool.On("OrganizationSetting.OpCreate", handleCustomerCreate)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
	}

	for _, event := range []string{"OrganizationSetting.OpUpdate", "OrganizationSetting.OpUpdateOne"} {
		_, err = pool.On(event, handleCustomerCreate)
		if err != nil {
			log.Error().Err(ErrFailedToRegisterListener)
		}
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

		if err := updateOrganizationSettingWithCustomerID(event.Context(), orgSettingID.(string), i.Customer().ID, event.Client()); err != nil {
			log.Err(err).Msg("Failed to update OrganizationSetting with Stripe customer ID")
			return err
		}

		log.Debug().Msgf("Updated OrganizationSetting with Stripe customer ID: %s", i.Customer().ID)
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
