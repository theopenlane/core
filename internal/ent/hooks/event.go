package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/models"
)

// Eventer is a wrapper struct for having a soiree as well as a list of listeners
type Eventer struct {
	Emitter   *soiree.EventPool
	Listeners []soiree.Listener
	Topics    map[string]any
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
func WithEventerTopics(topics map[string]any) EventerOpts {
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
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return nil, err
			}

			if reflect.TypeOf(retVal).Kind() == reflect.Int {
				zerolog.Ctx(ctx).Debug().Interface("value", retVal).Msg("mutation returned an int, skipping event emission")
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

				zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("mutation_id", eventID.ID)
				})

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
	},
		emitEventOn(),
	)
}

// emitEventOn is a function that returns a function that checks if an event should be emitted
// based on the mutation type and operation and fields that were updated
func emitEventOn() func(context.Context, entgen.Mutation) bool {
	return func(_ context.Context, m entgen.Mutation) bool {
		switch m.Type() {
		case entgen.TypeOrgSubscription:
			if m.Op().Is(ent.OpCreate) {
				return true
			}
		case entgen.TypeOrganizationSetting:
			if m.Op().Is(ent.OpUpdateOne) || m.Op().Is(ent.OpUpdate) {
				_, billingSetOK := m.Field("billing_email")
				_, phoneSetOK := m.Field("billing_phone")
				_, addressSetOK := m.Field("billing_address")

				if billingSetOK || phoneSetOK || addressSetOK {
					return true
				}
			}
		case entgen.TypeOrganization:
			if m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne) {
				return true
			}
		}

		return false
	}
}

// OrganizationSettingCreate and OrganizationSettingUpdateOne are the topics for the organization setting events; formatted as `type.operation`
var OrganizationSettingUpdateOne = fmt.Sprintf("%s.%s", entgen.TypeOrganizationSetting, entgen.OpUpdateOne.String())
var OrgSubscriptionCreate = fmt.Sprintf("%s.%s", entgen.TypeOrgSubscription, entgen.OpCreate.String())
var OrganizationDelete = fmt.Sprintf("%s.%s", entgen.TypeOrganization, entgen.OpDelete.String())
var OrganizationDeleteOne = fmt.Sprintf("%s.%s", entgen.TypeOrganization, entgen.OpDeleteOne.String())

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

	_, err = e.Emitter.On(OrganizationSettingUpdateOne, handleOrganizationSettingsUpdateOne)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	_, err = e.Emitter.On(OrganizationDelete, handleOrganizationDelete)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	_, err = e.Emitter.On(OrganizationDeleteOne, handleOrganizationDelete)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	return nil
}

// handleOrganizationDelete handles the deletion of an organization and deletes the customer in Stripe
func handleOrganizationDelete(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		zerolog.Ctx(event.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer deletion")

		return nil
	}

	if err := entMgr.FindAndDeactivateCustomerSubscription(event.Context(), lo.ValueOr(event.Properties(), "ID", "").(string)); err != nil {
		zerolog.Ctx(event.Context()).Error().Err(err).Msg("Failed to deactivate customer subscription")

		return err
	}

	return nil
}

// handleOrgSubscriptionCreated checks for the creation of an organization subscription and creates a customer in Stripe
func handleOrgSubscriptionCreated(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		zerolog.Ctx(event.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	// setup the context to allow the creation of a customer subscription without any restrictions
	allowCtx := privacy.DecisionContext(event.Context(), privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})

	orgCustomer := &entitlements.OrganizationCustomer{}

	orgSubs, err := client.OrgSubscription.Get(allowCtx, lo.ValueOr(event.Properties(), "ID", "").(string))
	if err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization subscription")

		return err
	}

	orgCustomer.OrganizationSubscriptionID = orgSubs.ID

	if err := updateOrgCustomerWithSubscription(allowCtx, orgSubs, client, orgCustomer); err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization from subscription")

		return nil
	}

	if err = entMgr.FindOrCreateCustomer(allowCtx, orgCustomer); err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to create customer")

		return err
	}

	if err := updateCustomerOrgSub(allowCtx, orgCustomer, client); err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to map customer to org subscription")

		return err
	}

	return nil
}

// updateCustomerOrgSub maps the customer fields to the organization subscription and updates the orgproduct and orgprice fields
func updateCustomerOrgSub(ctx context.Context, customer *entitlements.OrganizationCustomer, client any) error {
	if customer.OrganizationSubscriptionID == "" {
		zerolog.Ctx(ctx).Error().Msg("organization subscription ID is empty on customer, unable to update organization subscription")
		return ErrNoSubscriptions
	}

	entClient := client.(*entgen.Client)

	// Shuffle product and price info into OrgProduct and OrgPrice for this org/subscription
	for _, price := range customer.Prices {
		// Update OrgProduct with product info
		_, err := entClient.OrgProduct.Update().
			Where(
				// Assuming OrgProduct has fields: OrganizationID, SubscriptionID, ProductID
				// and that ProductID matches price.ProductID
				// Adjust predicates as needed for your schema
				// ...existing code...
				// Example:
				// orgproduct.OrganizationID(customer.OrganizationID),
				// orgproduct.SubscriptionID(customer.OrganizationSubscriptionID),
				// orgproduct.ProductID(price.ProductID),
			).
			SetProductName(price.ProductName).
			Save(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to update OrgProduct")
			return err
		}

		// Update OrgPrice with price info
		_, err = entClient.OrgPrice.Update().
			Where(
				// Assuming OrgPrice has fields: OrganizationID, SubscriptionID, ProductID, PriceID
				// and that PriceID matches price.PriceID
				// ...existing code...
				// Example:
				// orgprice.OrganizationID(customer.OrganizationID),
				// orgprice.SubscriptionID(customer.OrganizationSubscriptionID),
				// orgprice.ProductID(price.ProductID),
				// orgprice.PriceID(price.PriceID),
			).
			SetAmount(price.Price).
			SetCurrency(price.Currency).
			SetInterval(price.Interval).
			Save(ctx)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to update OrgPrice")
			return err
		}
	}

	return nil
}

// updateOrgCustomerWithSubscription updates the organization customer with the subscription data
// by querying the organization and organization settings
func updateOrgCustomerWithSubscription(ctx context.Context, orgSubs *entgen.OrgSubscription, client any, o *entitlements.OrganizationCustomer) error {
	if orgSubs == nil {
		return ErrNoSubscriptions
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSubs.OwnerID)).WithSetting().Only(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization by organization ID %s", orgSubs.OwnerID)

		return err
	}

	if org.Edges.Setting != nil {
		o.OrganizationSettingsID = org.Edges.Setting.ID
	} else {
		zerolog.Ctx(ctx).Debug().Msgf("Organization setting is nil for organization ID %s", orgSubs.OwnerID)
	}

	o.OrganizationID = org.ID
	o.OrganizationName = org.Name
	o.OrganizationSettingsID = org.Edges.Setting.ID
	o.PersonalOrg = org.PersonalOrg
	o.Email = org.Edges.Setting.BillingEmail

	return nil
}

// handleOrganizationSettingsUpdateOne handles the update of an organization setting and updates the customer in Stripe
// the event is only emitted if the billing settings change; so we proceed to update the customer in stripe
// based on the current organization settings
func handleOrganizationSettingsUpdateOne(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		zerolog.Ctx(event.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	orgCust, err := fetchOrganizationCustomerByOrgSettingID(event.Context(), lo.ValueOr(event.Properties(), "ID", "").(string), client)
	if err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization ID by organization setting ID")

		return err
	}

	if orgCust.StripeCustomerID != "" {
		params := entitlements.GetUpdatedFields(event.Properties(), orgCust)
		if params != nil {
			if _, err := entMgr.UpdateCustomer(event.Context(), orgCust.StripeCustomerID, params); err != nil {
				zerolog.Ctx(event.Context()).Err(err).Msg("Failed to update customer")

				return err
			}
		}
	}

	return nil
}

// fetchOrganizationCustomerByOrgSettingID fetches the organization customer data based on the organization setting ID
func fetchOrganizationCustomerByOrgSettingID(ctx context.Context, orgSettingID string, client any) (*entitlements.OrganizationCustomer, error) {
	orgSetting, err := client.(*entgen.Client).OrganizationSetting.Get(ctx, orgSettingID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization setting ID %s", orgSettingID)

		return nil, err
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSetting.OrganizationID)).WithOrgSubscriptions().Only(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization by organization setting ID %s after 3 attempts", orgSettingID)

		return nil, err
	}

	personalOrg := org.PersonalOrg

	if len(org.Edges.OrgSubscriptions) > 1 {
		zerolog.Ctx(ctx).Warn().Str("organization_id", org.ID).Msg("organization has multiple subscriptions")

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
