package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

const (
	propertyOrganizationDataKey      = "hooks.organization.data"
	propertyOrgSettingCustomerKey    = "hooks.organization.setting.customer"
	propertyOrganizationSettingIDKey = "hooks.organization.setting.id"
)

type organizationEventData struct {
	AllowCtx     context.Context
	Organization *entgen.Organization
	Subscription *entgen.OrgSubscription
	Customer     *entitlements.OrganizationCustomer
}

func organizationCreatePropertyExtractor(ctx context.Context, mutation ent.Mutation) soiree.Properties {
	m, ok := mutation.(*entgen.OrganizationMutation)
	if !ok {
		return nil
	}

	id, exists := m.ID()
	if !exists {
		return nil
	}

	client := m.Client()

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})

	org, err := client.Organization.Query().
		Where(organization.ID(id)).
		WithSetting().
		Only(allowCtx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("organization_id", id).Msg("property extractor: failed to fetch organization after create")

		return nil
	}

	orgSubs, err := client.OrgSubscription.Query().Where(orgsubscription.OwnerID(org.ID)).First(allowCtx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("organization_id", id).Msg("property extractor: failed to fetch org subscription after create")

		return nil
	}

	orgCustomer := &entitlements.OrganizationCustomer{OrganizationSubscriptionID: orgSubs.ID}

	orgCustomer, err = updateOrgCustomerWithSubscription(allowCtx, orgSubs, orgCustomer, org)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("organization_id", id).Msg("property extractor: failed to build organization customer after create")

		return nil
	}

	props := soiree.NewProperties()
	props.Set(propertyOrganizationDataKey, &organizationEventData{
		AllowCtx:     allowCtx,
		Organization: org,
		Subscription: orgSubs,
		Customer:     orgCustomer,
	})

	return props
}

func organizationDeletePropertyExtractor(ctx context.Context, mutation ent.Mutation) soiree.Properties {
	m, ok := mutation.(*entgen.OrganizationMutation)
	if !ok {
		return nil
	}

	id, exists := m.ID()
	if !exists {
		return nil
	}

	client := m.Client()

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{})
	allowCtx = context.WithValue(allowCtx, entx.SoftDeleteSkipKey{}, true)

	org, err := client.Organization.Query().Where(
		organization.And(
			organization.ID(id),
			organization.DeletedAtNotNil(),
		)).
		Only(allowCtx)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("organization_id", id).Msg("property extractor: failed to fetch organization after delete")

		return nil
	}

	props := soiree.NewProperties()
	props.Set(propertyOrganizationDataKey, &organizationEventData{
		AllowCtx:     allowCtx,
		Organization: org,
	})

	return props
}

func organizationSettingPropertyExtractor(ctx context.Context, mutation ent.Mutation) soiree.Properties {
	m, ok := mutation.(*entgen.OrganizationSettingMutation)
	if !ok {
		return nil
	}

	id, exists := m.ID()
	if !exists {
		return nil
	}

	client := m.Client()

	orgCustomer, err := fetchOrganizationCustomerByOrgSettingID(ctx, id, client)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("organization_setting_id", id).Msg("property extractor: failed to load organization customer for setting update")

		return nil
	}

	props := soiree.NewProperties()
	props.Set(propertyOrgSettingCustomerKey, orgCustomer)
	props.Set(propertyOrganizationSettingIDKey, id)

	return props
}
