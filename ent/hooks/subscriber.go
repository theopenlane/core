package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/subscriber"
	"github.com/theopenlane/shared/gqlerrors"
	"github.com/theopenlane/shared/logx"
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

			if err := queueSubscriberEmail(ctx, m); err != nil {
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

// queueSubscriberEmail queues the email to be sent to the subscriber
func queueSubscriberEmail(ctx context.Context, m *generated.SubscriberMutation) error {
	// Get the details from the mutation, these will never be empty because they are set in the hook
	// or are required fields
	orgID, _ := m.OwnerID()
	tok, _ := m.Token()
	e, _ := m.Email()

	// Get the organization name
	org, err := m.Client().Organization.Get(ctx, orgID)
	if err != nil {
		return err
	}

	email, err := m.Emailer.NewSubscriberEmail(emailtemplates.Recipient{
		Email: e,
	}, org.DisplayName, tok)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error rendering email")

		return err
	}

	// send the email via the job queue
	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing email verification")

		return err
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
