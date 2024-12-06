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
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
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

// TODO [MKA]: create better methods for constructing object + event topic names; possibly update soiree to make topic management more friendly
// OrganizationSettingCreate and OrganizationSettingUpdateOne are the topics for the organization setting events
var OrganizationSettingCreate = "OrganizationSetting.OpCreate"
var OrganizationSettingUpdateOne = "OrganizationSetting.OpUpdateOne"

// RegisterGlobalHooks registers global event hooks for the entdb client and expects a pointer to an Eventer
func RegisterGlobalHooks(client *entgen.Client, e *Eventer) {
	client.Use(EmitEventHook(e))
}

// RegisterListeners is currently used to globally register what listeners get applied on the entdb client
func RegisterListeners(e *Eventer) error {
	for _, event := range []string{OrganizationSettingCreate, OrganizationSettingUpdateOne} {
		_, err := e.Emitter.On(event, handleOrganizationSettingEvents)
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

	orgCust, err := fetchOrganizationIDbyOrgSettingID(event.Context(), lo.ValueOr(event.Properties(), "ID", "").(string), client)
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
		if params, hasUpdate := entitlements.CheckForBillingUpdate(event.Properties(), orgCust); hasUpdate {
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
	time.Sleep(3 * time.Second)
	orgSetting, err := client.(*entgen.Client).OrganizationSetting.Get(ctx, orgsettingID)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization setting ID %s", orgsettingID)
		return nil, err
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSetting.OrganizationID)).Only(ctx)
	if err != nil {
		log.Err(err).Msgf("Failed to fetch organization by organization setting ID %s after 3 attempts", orgsettingID)
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
