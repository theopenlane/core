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

func InitEventPool(client interface{}) *soiree.EventPool {
	return soiree.NewEventPool(soiree.WithPool(soiree.NewPondPool(soiree.WithMaxWorkers(100), soiree.WithName("ent_event_pool"))), soiree.WithErrorHandler(func(event soiree.Event, err error) error { // nolint: mnd
		log.Printf("Error encountered during event '%s': %v, with payload: %v", event.Topic(), err, event.Payload())
		return nil
	}), soiree.WithClient(client))
}

// RegisterListeners registers listeners for events
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

// handleCustomerCreate handles the creation of a customer in Stripe
func handleCustomerCreate(event soiree.Event) error {
	props := event.Properties()
	// TODO: add funcs for default unmarshalling of props and send all props to stripe
	orgsettingID, exists := props["ID"]
	if !exists {
		log.Info().Msg("organizationSetting ID not found in event properties")
		return nil
	}

	// if stripe ID is being added to the properties, we enumerate through any other fields in the update and ensure they match
	stripeID, exists := props["stripe_id"]
	if exists && stripeID != "" {
		stripes := stripeID.(string)

		stripecustomer, err := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.Get(stripes, nil)
		if err != nil {
			log.Err(err).Msg("Failed to retrieve Stripe customer by ID")
			return err
		}

		billingEmail, exists := props["billing_email"]
		if exists && billingEmail != "" {
			email := billingEmail.(string)
			if stripecustomer.Email != email {
				_, err := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.Update(stripes, &stripe.CustomerParams{
					Email: &email,
				})
				if err != nil {
					log.Err(err).Msg("failed to update Stripe customer email")
					return err
				}
			}
		}

		billingPhone, exists := props["billing_phone"]
		if exists && billingPhone != "" {
			phone := billingPhone.(string)
			if stripecustomer.Phone != phone {
				_, err := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.Update(stripes, &stripe.CustomerParams{
					Phone: &phone,
				})
				if err != nil {
					log.Err(err).Msg("failed to update Stripe customer phone")
					return err
				}
			}
		}

		// TODO: split out address fields in ent schema
		billingAddress, exists := props["billing_address"]
		if exists && billingAddress != "" {
			address := billingAddress.(string)
			if stripecustomer.Address.Line1 != address {
				_, err := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.Update(stripes, &stripe.CustomerParams{
					Address: &stripe.AddressParams{
						Line1: &address,
					},
				})
				if err != nil {
					log.Err(err).Msg("failed to update Stripe customer address")
					return err
				}
			}
		}
	}

	billingEmail, exists := props["billing_email"]
	if exists && billingEmail != "" {
		email := billingEmail.(string)

		i := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.List(&stripe.CustomerListParams{Email: &email})

		if !i.Next() {
			customer, err := event.Client().(*entgen.Client).EntitlementManager.Client.Customers.New(&stripe.CustomerParams{Email: &email})
			if err != nil {
				log.Err(err).Msg("Failed to create Stripe customer")
				return err
			}

			log.Info().Msgf("Created Stripe customer with ID: %s", customer.ID)

			if err := updateOrganizationSettingWithCustomerID(event.Context(), orgsettingID.(string), customer.ID, event.Client()); err != nil {
				log.Err(err).Msg("Failed to update OrganizationSetting with Stripe customer ID")
				return err
			}

			log.Info().Msgf("Updated OrganizationSetting with Stripe customer ID: %s", customer.ID)
		}

		subs, err := event.Client().(*entgen.Client).EntitlementManager.ListOrCreateStripeSubscriptions(i.Customer().ID)
		if err != nil {
			log.Err(err).Msg("failed to list or create Stripe subscriptions")
			return err
		}

		log.Info().Msgf("Created stripe subscription with ID %s", subs.ID)

		if err := updateOrganizationSettingWithCustomerID(event.Context(), orgsettingID.(string), i.Customer().ID, event.Client()); err != nil {
			log.Err(err).Msg("Failed to update OrganizationSetting with Stripe customer ID")
			return err
		}

		log.Info().Msgf("Updated OrganizationSetting with Stripe customer ID: %s", i.Customer().ID)
	}

	return nil
}

// updateOrganizationSettingWithCustomerID updates an OrganizationSetting with a Stripe customer ID
func updateOrganizationSettingWithCustomerID(ctx context.Context, orgID, customerID string, client interface{}) error {
	if _, err := client.(*entgen.Client).OrganizationSetting.UpdateOneID(orgID).SetStripeID(customerID).Save(ctx); err != nil {
		log.Err(err).Msgf("Failed to update OrganizationSetting ID %s with Stripe customer ID %s", orgID, customerID)

		return err
	}

	log.Info().Msgf("Updated OrganizationSetting ID %s with Stripe customer ID %s", orgID, customerID)

	return nil
}
