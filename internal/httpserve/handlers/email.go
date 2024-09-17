package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/pkg/jobs"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/utils/emails"
)

// SendVerificationEmail sends an email to a user to verify their email address
func (h *Handler) SendVerificationEmail(ctx context.Context, user *User) error {
	email, err := h.Email.NewVerifyEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, user.GetVerificationToken())
	if err != nil {
		log.Error().Err(err).Msg("error creating email verification")

		return err
	}

	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
}

// SendSubscriberEmail sends an email to confirm a user's subscription
func (h *Handler) SendSubscriberEmail(ctx context.Context, user *User, orgName string) error {
	email, err := h.Email.NewSubscriberEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, orgName, user.GetVerificationToken())
	if err != nil {
		log.Error().Err(err).Msg("error creating email verification")

		return err
	}

	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
}

// SendPasswordResetRequestEmail Send an email to a user to request them to reset their password
func (h *Handler) SendPasswordResetRequestEmail(ctx context.Context, user *User) error {
	email, err := h.Email.NewPasswordResetRequestEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, user.GetVerificationToken())
	if err != nil {
		log.Error().Err(err).Msg("error creating password reset email")

		return err
	}

	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing  password reset email")

		return err
	}

	return nil
}

// SendPasswordResetSuccessEmail Send an email to a user to inform them that their password has been reset
func (h *Handler) SendPasswordResetSuccessEmail(ctx context.Context, user *User) error {
	email, err := h.Email.NewPasswordResetSuccessEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		log.Error().Err(err).Msg("error creating password reset success email")

		return err
	}

	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing  password reset success email")

		return err
	}

	return nil
}

// SendOrgInvitationEmail sends an email inviting a user to join an existing organization
func (h *Handler) SendOrgInvitationEmail(ctx context.Context, i *emails.Invite) error {
	email, err := h.Email.NewInviteEmail(emailtemplates.Recipient{
		Email: i.Recipient,
	}, i.Requestor, i.OrgName, i.Role, i.Token)
	if err != nil {
		log.Error().Err(err).Msg("error creating password reset success email")

		return err
	}

	_, err = h.JobQueue.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		log.Error().Err(err).Msg("error queueing  password reset success email")

		return err
	}

	return nil
}
