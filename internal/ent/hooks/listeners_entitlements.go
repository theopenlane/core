package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	sync "github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// handleOrganizationMutation routes organization mutations to the correct entitlement handler
func handleOrganizationMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload == nil {
		return nil
	}

	switch payload.Operation {
	case ent.OpCreate.String():
		return handleOrganizationCreated(ctx, payload)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return handleOrganizationDelete(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationSettingMutation handles billing-related updates on organization settings
func handleOrganizationSettingMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload == nil {
		return nil
	}

	switch payload.Operation {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return handleOrganizationSettingsUpdateOne(ctx, payload)
	default:
		return nil
	}
}

// handleOrganizationDelete deactivates an organization's customer subscription when it is deleted
func handleOrganizationDelete(ctx *soiree.EventContext, payload *MutationPayload) error {
	inv, ok := newEntitlementInvocation(ctx, payload, softDeleteAllowContext)
	if !ok {
		return nil
	}

	org, err := inv.client.Organization.Query().Where(
		organization.And(
			organization.ID(inv.orgID),
			organization.DeletedAtNotNil(),
		),
	).Only(inv.Allow())
	if err != nil {
		inv.Logger().Err(err).Str("organization_id", inv.orgID).Msg("organization delete event unable to load organization")
		return nil
	}

	if org.StripeCustomerID == nil {
		return nil
	}

	if err := inv.client.EntitlementManager.FindAndDeactivateCustomerSubscription(inv.Context(), *org.StripeCustomerID); err != nil {
		inv.Logger().Error().Err(err).Msg("failed to deactivate customer subscription")
		return err
	}

	return nil
}

// handleOrganizationCreated reconciles entitlements after an organization is created
func handleOrganizationCreated(ctx *soiree.EventContext, payload *MutationPayload) error {
	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	return inv.reconcile()
}

// handleOrganizationSettingsUpdateOne updates Stripe customer details when billing fields change
func handleOrganizationSettingsUpdateOne(ctx *soiree.EventContext, payload *MutationPayload) error {
	if !mutationTouches(payload.Mutation, "billing_email", "billing_phone", "billing_address") {
		return nil
	}

	inv, ok := newEntitlementInvocation(ctx, payload, orgAllowContext)
	if !ok {
		return nil
	}

	orgSettingID := inv.entityID
	if orgSettingID == "" {
		if id, ok := mutationEntityID(ctx, payload); ok {
			orgSettingID = id
		}
	}

	if orgSettingID == "" {
		inv.Logger().Warn().Msg("organization settings update missing entity id; skipping stripe update")
		return nil
	}

	orgCustomer, err := fetchOrganizationCustomerByOrgSettingID(inv, orgSettingID)
	if err != nil {
		// We deliberately bubble the failure so the GraphQL mutation surfaces the issue
		inv.Logger().Err(err).Str("organization_setting_id", orgSettingID).Msg("failed to fetch organization customer")

		return err
	}

	if orgCustomer == nil || orgCustomer.StripeCustomerID == "" {
		return inv.reconcile()
	}

	params := entitlements.GetUpdatedFields(ctx.Properties(), orgCustomer)

	if params != nil {
		if _, err := inv.client.EntitlementManager.UpdateCustomer(inv.Context(), orgCustomer.StripeCustomerID, params); err != nil {
			// Stripe update failures should not fall back to reconciliation because the reconciler
			// assumes the customer metadata is already aligned
			inv.Logger().Err(err).Str("stripe_customer_id", orgCustomer.StripeCustomerID).Msg("failed to update stripe customer metadata")

			return err
		}
	}

	return inv.reconcile()
}

var errMissingOrgCustomerPrereqs = errors.New("entitlement invocation missing prerequisites")

// entitlementInvocation bundles the data required for entitlement listeners to perform their work
type entitlementInvocation struct {
	event    *soiree.EventContext
	payload  *MutationPayload
	client   *entgen.Client
	orgID    string
	entityID string
	allow    context.Context
}

// Context returns the listener context associated with the invocation
func (inv *entitlementInvocation) Context() context.Context {
	return inv.event.Context()
}

// Logger returns a contextual logger for the invocation
func (inv *entitlementInvocation) Logger() *zerolog.Logger {
	return logx.FromContext(inv.Context())
}

// Allow returns the elevated context used for entitlement operations
func (inv *entitlementInvocation) Allow() context.Context {
	return inv.allow
}

// orgAllowContext returns a context that bypasses privacy rules when running entitlement logic against an organization
func orgAllowContext(ctx context.Context) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})
}

// softDeleteAllowContext extends orgAllowContext to skip soft delete filters so listeners can access archived records
func softDeleteAllowContext(ctx context.Context) context.Context {
	ctx = orgAllowContext(ctx)

	return context.WithValue(ctx, entx.SoftDeleteSkipKey{}, true)
}

// newEntitlementInvocation gathers the elements required to run entitlement logic for a mutation
func newEntitlementInvocation(event *soiree.EventContext, payload *MutationPayload, allow func(context.Context) context.Context) (*entitlementInvocation, bool) {
	client := mutationClient(event, payload)
	if client == nil || client.EntitlementManager == nil {
		return nil, false
	}

	if allow == nil {
		// hard to spot but this is a func signature check, not a nil comparison
		allow = orgAllowContext
	}

	allowCtx := allow(event.Context())

	entityID, ok := mutationEntityID(event, payload)
	if !ok {
		return nil, false
	}

	orgID := entityID

	if payload != nil && payload.Mutation != nil && payload.Mutation.Type() == entgen.TypeOrganizationSetting {
		// OrganizationSetting mutations carry the setting ID, but the reconciler needs the owning organization
		setting, err := client.OrganizationSetting.Get(allowCtx, entityID)
		if err != nil {
			logx.FromContext(event.Context()).Err(err).Str("organization_setting_id", entityID).Msg("failed to resolve organization from organization setting")
			return nil, false
		}

		orgID = setting.OrganizationID
	}

	return &entitlementInvocation{
		event:    event,
		payload:  payload,
		client:   client,
		orgID:    orgID,
		entityID: entityID,
		allow:    allowCtx,
	}, true
}

// mutationEntityID derives the entity identifier from the payload or event properties
func mutationEntityID(ctx *soiree.EventContext, payload *MutationPayload) (string, bool) {
	if payload != nil && payload.EntityID != "" {
		return payload.EntityID, true
	}

	if ctx == nil {
		return "", false
	}

	if id, ok := ctx.PropertyString("ID"); ok && id != "" {
		return id, true
	}

	if raw, ok := ctx.Property("ID"); ok && raw != nil {
		if str, ok := raw.(fmt.Stringer); ok {
			value := str.String()
			if value == "" {
				return "", false
			}

			return value, true
		}

		value := fmt.Sprint(raw)
		if value == "" || value == "<nil>" {
			return "", false
		}

		return value, true
	}

	return "", false
}

// mutationClient returns the ent client associated with the mutation
func mutationClient(ctx *soiree.EventContext, payload *MutationPayload) *entgen.Client {
	if payload != nil && payload.Client != nil {
		return payload.Client
	}

	client, ok := soiree.ClientAs[*entgen.Client](ctx)
	if !ok {
		return nil
	}

	return client
}

// mutationTouches reports whether the mutation updates ("touches" don't get weird) any of the requested fields
func mutationTouches(m ent.Mutation, fields ...string) bool {
	if m == nil {
		return false
	}

	for _, field := range fields {
		if _, ok := m.Field(field); ok {
			return true
		}
	}

	return false
}

// fetchOrganizationCustomerByOrgSettingID loads the organization and customer data linked to an organization setting
func fetchOrganizationCustomerByOrgSettingID(inv *entitlementInvocation, orgSettingID string) (*entitlements.OrganizationCustomer, error) {
	if inv == nil || inv.client == nil || orgSettingID == "" {
		return nil, fmt.Errorf("%w: organization_setting_id=%s", errMissingOrgCustomerPrereqs, orgSettingID)
	}

	orgSetting, err := inv.client.OrganizationSetting.Get(inv.Allow(), orgSettingID)
	if err != nil {
		return nil, err
	}

	org, err := inv.client.Organization.Query().
		Where(organization.ID(orgSetting.OrganizationID)).
		Only(inv.Allow())
	if err != nil {
		return nil, err
	}

	stripeCustomerID := ""
	if org.StripeCustomerID != nil {
		stripeCustomerID = *org.StripeCustomerID
	}

	contact := entitlements.ContactInfo{
		Email: orgSetting.BillingEmail,
		Phone: orgSetting.BillingPhone,
	}

	if address := orgSetting.BillingAddress; address != (models.Address{}) {
		contact.Line1 = lo.ToPtr(address.Line1)
		contact.Line2 = lo.ToPtr(address.Line2)
		contact.City = lo.ToPtr(address.City)
		contact.State = lo.ToPtr(address.State)
		contact.Country = lo.ToPtr(address.Country)
		contact.PostalCode = lo.ToPtr(address.PostalCode)
	}

	return &entitlements.OrganizationCustomer{
		OrganizationID:         org.ID,
		OrganizationName:       org.Name,
		OrganizationSettingsID: orgSetting.ID,
		StripeCustomerID:       stripeCustomerID,
		ContactInfo:            contact,
	}, nil
}

// reconcile runs the entitlement reconciler for the invocation's organization
func (inv *entitlementInvocation) reconcile() error {
	if inv == nil || inv.client == nil || inv.client.EntitlementManager == nil {
		return nil
	}

	reconciler, err := sync.New(
		sync.WithDB(inv.client),
		sync.WithStripeClient(inv.client.EntitlementManager),
	)
	if err != nil {
		inv.Logger().Err(err).Msg("unable to construct entitlement reconciler")

		return err
	}

	if _, err := reconciler.Reconcile(inv.Context(), []string{inv.orgID}); err != nil {
		inv.Logger().Err(err).Msg("entitlement reconciliation failed")

		return err
	}

	inv.Logger().Debug().Msg("entitlement reconciliation completed")

	return nil
}
