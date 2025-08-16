package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"time"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v82"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/slack"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	catalog "github.com/theopenlane/core/internal/entitlements/entmapping"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/slacktemplates"
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

// SlackConfig holds configuration for Slack notifications
type SlackConfig struct {
	WebhookURL               string
	NewSubscriberMessageFile string
	NewUserMessageFile       string
}

var slackCfg SlackConfig

// SetSlackConfig sets the Slack configuration for event handlers
func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
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
			switch m.Op() {
			case ent.OpDelete, ent.OpDeleteOne, ent.OpCreate:
				return true
			}
		case entgen.TypeSubscriber:
			if m.Op().Is(ent.OpCreate) {
				return true
			}
		case entgen.TypeUser:
			if m.Op().Is(ent.OpCreate) {
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
var OrganizationCreate = fmt.Sprintf("%s.%s", entgen.TypeOrganization, entgen.OpCreate.String())
var OrganizationDeleteOne = fmt.Sprintf("%s.%s", entgen.TypeOrganization, entgen.OpDeleteOne.String())
var SubscriberCreate = fmt.Sprintf("%s.%s", entgen.TypeSubscriber, entgen.OpCreate.String())
var UserCreate = fmt.Sprintf("%s.%s", entgen.TypeUser, entgen.OpCreate.String())

// RegisterGlobalHooks registers global event hooks for the entdb client and expects a pointer to an Eventer
func RegisterGlobalHooks(client *entgen.Client, e *Eventer) {
	client.Use(EmitEventHook(e))
}

// RegisterListeners is currently used to globally register what listeners get applied on the entdb client
func RegisterListeners(e *Eventer) error {
	_, err := e.Emitter.On(OrganizationCreate, handleOrganizationCreated)
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

	_, err = e.Emitter.On(SubscriberCreate, handleSubscriberCreate)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	_, err = e.Emitter.On(UserCreate, handleUserCreate)
	if err != nil {
		log.Error().Err(ErrFailedToRegisterListener)
		return err
	}

	return nil
}

// handleSubscriberCreate sends a Slack notification when a new subscriber is created
func handleSubscriberCreate(event soiree.Event) error {
	if slackCfg.WebhookURL == "" {
		return nil
	}

	emailVal := event.Properties().GetKey("email")
	email, _ := emailVal.(string)

	var (
		t   *template.Template
		err error
		msg string
	)

	if slackCfg.NewSubscriberMessageFile != "" {
		b, err := os.ReadFile(slackCfg.NewSubscriberMessageFile)
		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to read slack template")

			return err
		}

		t, err = template.New("slack").Parse(string(b))

		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to parse slack template")

			return err
		}
	} else {
		t, err = template.ParseFS(slacktemplates.Templates, slacktemplates.SubscriberTemplateName)
		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to parse embedded slack template")

			return err
		}
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		zerolog.Ctx(event.Context()).Debug().Msg("failed to execute slack template")

		return err
	}

	msg = buf.String()

	client := slack.New(slackCfg.WebhookURL)

	payload := &slack.Payload{Text: msg}

	return client.Post(event.Context(), payload)
}

// handleUserCreate sends a Slack notification when a new user is created
func handleUserCreate(event soiree.Event) error {
	if slackCfg.WebhookURL == "" {
		return nil
	}

	emailVal := event.Properties().GetKey("email")
	email, _ := emailVal.(string)

	var (
		t   *template.Template
		err error
		msg string
	)

	if slackCfg.NewUserMessageFile != "" {
		b, err := os.ReadFile(slackCfg.NewUserMessageFile)
		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to read slack template")

			return err
		}

		t, err = template.New("slack").Parse(string(b))
		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to parse slack template")

			return err
		}
	} else {
		t, err = template.ParseFS(slacktemplates.Templates, slacktemplates.UserTemplateName)
		if err != nil {
			zerolog.Ctx(event.Context()).Debug().Msg("failed to parse embedded slack template")

			return err
		}
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		zerolog.Ctx(event.Context()).Debug().Msg("failed to execute slack template")

		return err
	}

	msg = buf.String()

	client := slack.New(slackCfg.WebhookURL)

	payload := &slack.Payload{Text: msg}

	return client.Post(event.Context(), payload)
}

// handleOrganizationDelete handles the deletion of an organization and deletes the customer in Stripe
func handleOrganizationDelete(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		zerolog.Ctx(event.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer deletion")

		return nil
	}

	// setup the context to allow the creation of a customer subscription without any restrictions
	allowCtx := privacy.DecisionContext(event.Context(), privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})

	org, err := client.Organization.Get(allowCtx, lo.ValueOr(event.Properties(), "ID", "").(string))
	if err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization")

		return err
	}

	if org.StripeCustomerID == nil {
		return nil
	}

	if err := entMgr.FindAndDeactivateCustomerSubscription(event.Context(), *org.StripeCustomerID); err != nil {
		zerolog.Ctx(event.Context()).Error().Err(err).Msg("Failed to deactivate customer subscription")

		return err
	}

	return nil
}

// handleOrganizationCreated checks for the creation of an organization subscription and creates a customer in Stripe
func handleOrganizationCreated(event soiree.Event) error {
	client := event.Client().(*entgen.Client)
	entMgr := client.EntitlementManager

	if entMgr == nil {
		zerolog.Ctx(event.Context()).Debug().Msg("EntitlementManager not found on client, skipping customer creation")

		return nil
	}

	// setup the context to allow the creation of a customer subscription without any restrictions
	allowCtx := privacy.DecisionContext(event.Context(), privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})

	org, err := client.Organization.Get(allowCtx, lo.ValueOr(event.Properties(), "ID", "").(string))
	if err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization")

		return err
	}

	orgSubs, err := client.OrgSubscription.Query().Where(orgsubscription.OwnerID(org.ID)).First(allowCtx)
	if err != nil {
		return err
	}

	orgCustomer := &entitlements.OrganizationCustomer{OrganizationSubscriptionID: orgSubs.ID}

	orgCustomer, err = updateOrgCustomerWithSubscription(allowCtx, orgSubs, client, orgCustomer)
	if err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to fetch organization from subscription")

		return nil
	}

	// personalOrg flag is set by updateOrgCustomerWithSubscription so compute pricing afterwards
	orgCustomer = catalog.PopulatePricesForOrganizationCustomer(orgCustomer)

	zerolog.Ctx(event.Context()).Debug().Msgf("Prices attached to organization customer: %+v", orgCustomer.Prices)

	if err = entMgr.CreateCustomerAndSubscription(allowCtx, orgCustomer); err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to create customer")

		return err
	}

	if err := updateCustomerOrgSub(allowCtx, orgCustomer, client); err != nil {
		zerolog.Ctx(event.Context()).Err(err).Msg("Failed to map customer to org subscription")

		return err
	}

	return nil
}

// updateCustomerOrgSub maps the customer fields to the organization subscription and update the organization subscription in the database
func updateCustomerOrgSub(ctx context.Context, customer *entitlements.OrganizationCustomer, client any) error {
	if customer.OrganizationSubscriptionID == "" {
		zerolog.Ctx(ctx).Error().Msg("organization subscription ID is empty on customer, unable to update organization subscription")

		return ErrNoSubscriptions
	}

	// update the expiration date based on the subscription status
	// if the subscription is trialing, set the expiration date to the trial end date
	// otherwise, set the expiration date to the end date if it exists
	trialExpiresAt := time.Unix(0, 0)
	if customer.Status == string(stripe.SubscriptionStatusTrialing) {
		trialExpiresAt = time.Unix(customer.TrialEnd, 0)
	}

	expiresAt := time.Unix(0, 0)
	if customer.EndDate > 0 {
		expiresAt = time.Unix(customer.EndDate, 0)
	}

	active := customer.Status == string(stripe.SubscriptionStatusActive) || customer.Status == string(stripe.SubscriptionStatusTrialing)

	c := client.(*entgen.Client)

	err := c.Organization.UpdateOneID(customer.OrganizationID).
		SetStripeCustomerID(customer.StripeCustomerID).
		Exec(ctx)
	if err != nil {
		return err
	}

	update := c.OrgSubscription.UpdateOneID(customer.OrganizationSubscriptionID).
		SetStripeSubscriptionID(customer.StripeSubscriptionID).
		SetStripeSubscriptionStatus(customer.Subscription.Status).
		SetActive(active)

	// ensure the correct expiration date is set based on the subscription status
	// if the subscription is trialing, set the expiration date to the trial end date
	// otherwise, set the expiration date to the end date
	if customer.Status == string(stripe.SubscriptionStatusTrialing) {
		update.SetTrialExpiresAt(trialExpiresAt)
	} else {
		update.SetExpiresAt(expiresAt)
	}

	return update.Exec(ctx)
}

// updateOrgCustomerWithSubscription updates the organization customer with the subscription data
// by querying the organization and organization settings
func updateOrgCustomerWithSubscription(ctx context.Context, orgSubs *entgen.OrgSubscription, client any, o *entitlements.OrganizationCustomer) (*entitlements.OrganizationCustomer, error) {
	if orgSubs == nil {
		return nil, ErrNoSubscriptions
	}

	org, err := client.(*entgen.Client).Organization.Query().Where(organization.ID(orgSubs.OwnerID)).WithSetting().Only(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msgf("Failed to fetch organization by organization ID %s", orgSubs.OwnerID)

		return nil, err
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

	return o, nil
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
