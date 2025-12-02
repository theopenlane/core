package handlers

import (
	"context"
	"fmt"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/shared/logx"
)

// sendVerificationEmail sends an email to a user to verify their email address
func (h *Handler) sendVerificationEmail(ctx context.Context, user *User, token string) error {
	email, err := h.Emailer.NewVerifyEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating email verification")

		return err
	}

	_, err = h.DBClient.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing email verification")

		return err
	}

	logx.FromContext(ctx).Debug().Msg("queued email")

	return nil
}

// SendSubscriberEmail sends an email to confirm a user's subscription
func (h *Handler) sendSubscriberEmail(ctx context.Context, user *User, orgID string) error {
	if orgID == "" {
		return fmt.Errorf("%w, subscriber organization not found", ErrMissingField)
	}

	org, err := h.getOrgByID(ctx, orgID)
	if err != nil {
		return err
	}

	email, err := h.Emailer.NewSubscriberEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, org.DisplayName, user.GetVerificationToken())
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating email verification")

		return err
	}

	_, err = h.DBClient.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
}

// sendPasswordResetRequestEmail to a user to request them to reset their password
func (h *Handler) sendPasswordResetRequestEmail(ctx context.Context, user *User) error {
	email, err := h.Emailer.NewPasswordResetRequestEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, user.GetPasswordResetToken())
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating password reset email")

		return err
	}

	_, err = h.DBClient.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing  password reset email")

		return err
	}

	return nil
}

// SendPasswordResetSuccessEmail Send an email to a user to inform them that their password has been reset
func (h *Handler) sendPasswordResetSuccessEmail(ctx context.Context, user *User) error {
	email, err := h.Emailer.NewPasswordResetSuccessEmail(emailtemplates.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating password reset success email")

		return err
	}

	_, err = h.DBClient.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error queueing  password reset success email")

		return err
	}

	return nil
}
