package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
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

			tokenRaw, tokenOK := m.Token()
			tokenValue, err := requiredMutationString("token", tokenRaw, tokenOK)
			if err != nil {
				return nil, err
			}

			emailRaw, emailOK := m.Email()
			emailAddress, err := requiredMutationString("email", emailRaw, emailOK)
			if err != nil {
				return nil, err
			}

			orgIDValue, ownerOK := m.OwnerID()
			orgID, err := requiredMutationString("owner_id", orgIDValue, ownerOK)
			if err != nil {
				return nil, err
			}

			orgName, err := organizationDisplayNameByID(ctx, m.Client(), orgID)
			if err != nil {
				return nil, err
			}

			input := emaildef.SubscribeRequest{
				RecipientInfo: emaildef.RecipientInfo{Email: emailAddress},
				OrgName:       orgName,
				Token:         tokenValue,
			}

			if receipt := emailGala.EmitWithHeaders(ctx, emaildef.SubscribeOp().Topic(), input,
				gala.NewHeaders([]string{"email", "subscriber"}, input)); receipt.Err != nil {
				return nil, receipt.Err
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

func getSubscriber(ctx context.Context, m *generated.SubscriberMutation) (*generated.Subscriber, error) {
	email, _ := m.Email()
	ownerID, _ := m.OwnerID()

	return m.Client().Subscriber.Query().
		Where(subscriber.Email(email)).
		Where(subscriber.OwnerID(ownerID)).Only(ctx)
}

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

	// if a user is unsubscribed but getting here again
	// we should toggle that
	if subscriber.Unsubscribed {
		subscriber.Unsubscribed = false
	}

	secret, _ := m.Secret()
	token, _ := m.Token()

	return m.Client().Subscriber.
		UpdateOneID(subscriber.ID).
		SetSendAttempts(subscriber.SendAttempts).
		SetUnsubscribed(subscriber.Unsubscribed).
		SetToken(token).
		SetSecret(secret).
		Save(ctx)
}
