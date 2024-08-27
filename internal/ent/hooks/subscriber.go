package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/sendgrid"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/iam/tokens"
)

// HookSubscriber runs on subscriber create mutations
func HookSubscriber() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubscriberFunc(func(ctx context.Context, m *generated.SubscriberMutation) (generated.Value, error) {
			email, ok := m.Email()
			if !ok || email == "" {
				return nil, ErrEmailRequired
			}

			if err := createVerificationToken(m, email); err != nil {
				return nil, err
			}

			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if err := queueSubscriberEmail(ctx, m); err != nil {
				return nil, err
			}

			return retValue, err
		})
	}, ent.OpCreate)
}

// queueSubscriberEmail queues the email to be sent to the subscriber
func queueSubscriberEmail(ctx context.Context, m *generated.SubscriberMutation) error {
	// Get the details from the mutation, these will never be empty because they are set in the hook
	orgID, _ := m.OwnerID()
	tok, _ := m.Token()
	email, _ := m.Email()

	// Get the organization name
	org, err := m.Client().Organization.Get(ctx, orgID)
	if err != nil {
		return err
	}

	// send emails via Marionette as to not create blocking operations in the server
	if err := m.Marionette.Queue(marionette.TaskFunc(func(ctx context.Context) error {
		return sendSubscriberEmail(m, org.Name, tok)
	}), marionette.WithRetries(3), //nolint:mnd
		marionette.WithErrorf("could not send subscriber verification email to user %s", email),
	); err != nil {
		return err
	}

	return nil
}

// sendSubscriberEmail sends an email to confirm a user's subscription
func sendSubscriberEmail(m *generated.SubscriberMutation, orgName, token string) error {
	e, _ := m.Email()

	data := emails.SubscriberEmailData{
		OrgName: orgName,
		EmailData: emails.EmailData{
			Sender: m.Emails.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: e,
			},
		},
	}

	var err error
	if data.VerifySubscriberURL, err = m.Emails.SubscriberVerifyURL(token); err != nil {
		return err
	}

	msg, err := emails.SubscribeEmail(data)
	if err != nil {
		return err
	}

	// Send the email
	return m.Emails.Send(msg)
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
