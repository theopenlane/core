package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/trustcenterurl"
	"github.com/theopenlane/core/pkg/logx"
)

// HookSubscriberCreate runs on subscriber create mutations
func HookSubscriberCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubscriberFunc(func(ctx context.Context, m *generated.SubscriberMutation) (generated.Value, error) {
			email, ok := m.Email()
			if !ok || email == "" {
				return nil, gqlerrors.NewCustomError(
					gqlerrors.BadRequestErrorCode,
					"subscriber email is required, please provide a valid email",

					ErrEmailRequired)
			}

			// lowercase the email for uniqueness
			m.SetEmail(strings.ToLower(email))

			// block subscriber creation for a trust center that has not enabled accepting subscribers
			if err := checkTrustCenterAllowsSubscribers(ctx, m); err != nil {
				return nil, err
			}

			if err := createVerificationToken(m, email); err != nil {
				return nil, err
			}

			var retValue ent.Value

			existingSubscriber, err := getSubscriber(ctx, m)

			if existingSubscriber != nil && err == nil {
				if existingSubscriber.Active {
					return nil, gqlerrors.NewCustomError(
						gqlerrors.AlreadyExistsErrorCode,
						"email is already subscribed to this organization",
						ErrUserAlreadySubscriber)
				}

				retValue, err = updateSubscriber(ctx, m, existingSubscriber)
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("unable to update email subscription")

					return retValue, err
				}
			} else {
				// create new subscription
				retValue, err = next.Mutate(ctx, m)
				if err != nil {
					return retValue, err
				}
			}

			tokenValue, _ := m.Token()
			emailAddress, _ := m.Email()
			orgID, _ := m.OwnerID()
			trustCenterID, _ := m.TrustCenterID()

			orgName, err := organizationDisplayNameByID(ctx, m.Client(), orgID)
			if err != nil {
				return nil, err
			}

			customDomain, slug, branding := subscriberTrustCenterDomain(ctx, m.Client(), trustCenterID)

			if err := sendSystemEmail(ctx, m.Client(), emaildef.SubscribeOp.Name(), emaildef.SubscribeRequest{
				RecipientInfo:       emaildef.RecipientInfo{Email: emailAddress},
				TrustCenterBranding: branding,
				OrgName:             orgName,
				Token:               tokenValue,
				VerifyURL:           trustcenterurl.SubscribeVerifyURLWithToken(customDomain, slug, tokenValue),
				UnsubscribeURL:      trustcenterurl.UnsubscribeURLWithToken(customDomain, slug, tokenValue),
			}); err != nil {
				return nil, err
			}

			return retValue, err
		})
	}, ent.OpCreate)
}

// HookSubscriberUpdated runs on subscriber update mutations to set the active status to false if the user is unsubscribed
func HookSubscriberUpdated() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.SubscriberFunc(func(ctx context.Context, m *generated.SubscriberMutation) (generated.Value, error) {
			unsubscribed, ok := m.Unsubscribed()
			if ok && unsubscribed {
				m.SetActive(false)
				m.SetSendAttempts(0)
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasOp(ent.OpUpdateOne),
			hook.HasFields("unsubscribed"),
		),
	)
}

// checkTrustCenterAllowsSubscribers rejects subscriber creation when the subscriber's trust center has
// not enabled accepting subscribers. Subscribers with no trust center have no setting to gate
func checkTrustCenterAllowsSubscribers(ctx context.Context, m *generated.SubscriberMutation) error {
	tcID, ok := m.TrustCenterID()
	if !ok || tcID == "" {
		return nil
	}

	setting, err := m.Client().TrustCenterSetting.Query().
		Where(
			trustcentersetting.TrustCenterID(tcID),
			trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentLive),
		).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", tcID).Msg("unable to load trust center setting for subscriber gate")

		return err
	}

	if !setting.AllowSubscribers {
		return gqlerrors.NewCustomError(
			gqlerrors.BadRequestErrorCode,
			"this trust center is not accepting new subscribers",
			ErrSubscribersNotAllowed)
	}

	return nil
}

// createVerificationToken creates a new email verification token for the user
func createVerificationToken(m *generated.SubscriberMutation, email string) error {
	// Create a unique token from the user's email address
	verify, err := tokens.NewVerificationToken(email)
	if err != nil {
		return err
	}

	// Sign the token to ensure that we can verify it later
	token, secret, err := verify.Sign()
	if err != nil {
		return err
	}

	m.SetToken(token)
	m.SetTTL(verify.ExpiresAt)
	m.SetSecret(secret)

	return nil
}

// getSubscriber looks up an existing subscriber by email and owner ID, optionally scoped to a trust center
func getSubscriber(ctx context.Context, m *generated.SubscriberMutation) (*generated.Subscriber, error) {
	email, _ := m.Email()
	ownerID, _ := m.OwnerID()

	query := m.Client().Subscriber.Query().
		Where(subscriber.Email(email)).
		Where(subscriber.OwnerID(ownerID))

	// scope the lookup to the trust center so the same email can subscribe to multiple trust centers
	// within the same organization; subscribers with no trust center match the null scope
	if tcID, ok := m.TrustCenterID(); ok && tcID != "" {
		query = query.Where(subscriber.TrustCenterID(tcID))
	} else {
		query = query.Where(subscriber.TrustCenterIDIsNil())
	}

	return query.Only(ctx)
}

// subscriberTrustCenterDomain resolves a subscriber's trust center custom domain and slug for link
// building along with the trust center branding for the confirmation email; zero values when there
// is no trust center or it cannot be resolved
func subscriberTrustCenterDomain(ctx context.Context, client *generated.Client, trustCenterID string) (customDomain, slug string, branding emaildef.TrustCenterBranding) {
	if trustCenterID == "" {
		return "", "", emaildef.TrustCenterBranding{}
	}

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to resolve trust center for subscriber link")

		return "", "", emaildef.TrustCenterBranding{}
	}

	if tc.Edges.CustomDomain != nil {
		customDomain = tc.Edges.CustomDomain.CnameRecord
	}

	return customDomain, tc.Slug, emaildef.TrustCenterBrandingFromSetting(tc.Edges.Setting)
}

// updateSubscriber updates an existing subscriber's send attempts and resets the verified email status
func updateSubscriber(ctx context.Context,
	m *generated.SubscriberMutation, subscriber *generated.Subscriber) (*generated.Subscriber, error) {
	if subscriber.SendAttempts >= maxAttempts {
		return nil, gqlerrors.NewCustomError(
			gqlerrors.MaxAttemptsErrorCode,
			"max attempts reached for this email, please reach out to support",
			ErrMaxSubscriptionAttempts)
	}

	subscriber.SendAttempts++

	m.SetSendAttempts(subscriber.SendAttempts)

	// a contact re-subscribing here (including one that previously unsubscribed) must re-confirm: clear
	// unsubscribed and reset verified_email so the fresh verify link drives reactivation through the
	// handler. This is the only path that resurrects an unsubscribed contact
	if subscriber.Unsubscribed {
		subscriber.Unsubscribed = false
	}

	secret, _ := m.Secret()
	token, _ := m.Token()
	ttl, _ := m.TTL()

	return m.Client().Subscriber.
		UpdateOneID(subscriber.ID).
		SetSendAttempts(subscriber.SendAttempts).
		SetUnsubscribed(subscriber.Unsubscribed).
		SetVerifiedEmail(false).
		SetToken(token).
		SetSecret(secret).
		SetTTL(ttl).
		Save(ctx)
}
