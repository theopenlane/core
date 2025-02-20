package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookSubscriber runs on subscriber create mutations
func HookSubscriber() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.SubscriberFunc(func(ctx context.Context, m *generated.SubscriberMutation) (generated.Value, error) {
			email, ok := m.Email()
			if !ok || email == "" {
				return nil, ErrEmailRequired
			}

			// lowercase the email for uniqueness
			m.SetEmail(strings.ToLower(email))

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
	}, org.Name, tok)
	if err != nil {
		log.Error().Err(err).Msg("error rendering email")

		return err
	}

	// send the email via the job queue
	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

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
