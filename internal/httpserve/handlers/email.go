package handlers

import (
	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/sendgrid"
)

// SendVerificationEmail sends an email to a user to verify their email address
func (h *Handler) SendVerificationEmail(user *User) error {
	contact := &sendgrid.Contact{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	data := emails.VerifyEmailData{
		EmailData: emails.EmailData{
			Sender: h.EmailManager.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			},
		},
		FullName: contact.FullName(),
	}

	var err error
	if data.VerifyURL, err = h.EmailManager.VerifyURL(user.GetVerificationToken()); err != nil {
		return err
	}

	msg, err := emails.VerifyEmail(data)
	if err != nil {
		return err
	}

	// Send the email
	return h.EmailManager.Send(msg)
}

// SendSubscriberEmail sends an email to confirm a user's subscription
func (h *Handler) SendSubscriberEmail(user *User, orgName string) error {
	data := emails.SubscriberEmailData{
		OrgName: orgName,
		EmailData: emails.EmailData{
			Sender: h.EmailManager.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: user.Email,
			},
		},
	}

	var err error
	if data.VerifySubscriberURL, err = h.EmailManager.SubscriberVerifyURL(user.GetVerificationToken()); err != nil {
		return err
	}

	msg, err := emails.SubscribeEmail(data)
	if err != nil {
		return err
	}

	// Send the email
	return h.EmailManager.Send(msg)
}

// SendPasswordResetRequestEmail Send an email to a user to request them to reset their password
func (h *Handler) SendPasswordResetRequestEmail(user *User) error {
	data := emails.ResetRequestData{
		EmailData: emails.EmailData{
			Sender: h.EmailManager.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			},
		},
	}
	data.Recipient.ParseName(user.Name)

	var err error
	if data.ResetURL, err = h.EmailManager.ResetURL(user.GetPasswordResetToken()); err != nil {
		return err
	}

	msg, err := emails.PasswordResetRequestEmail(data)
	if err != nil {
		return err
	}

	// Send the email
	return h.EmailManager.Send(msg)
}

// SendPasswordResetSuccessEmail Send an email to a user to inform them that their password has been reset
func (h *Handler) SendPasswordResetSuccessEmail(user *User) error {
	data := emails.ResetSuccessData{
		EmailData: emails.EmailData{
			Sender: h.EmailManager.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: user.Email,
			},
		},
	}

	data.Recipient.ParseName(user.Name)

	msg, err := emails.PasswordResetSuccessEmail(data)
	if err != nil {
		return err
	}

	// Send the email
	return h.EmailManager.Send(msg)
}

// SendOrgInvitationEmail sends an email inviting a user to join Datum and an existing organization
func (h *Handler) SendOrgInvitationEmail(i *emails.Invite) error {
	data := emails.InviteData{
		InviterName: i.Requestor,
		OrgName:     i.OrgName,
		EmailData: emails.EmailData{
			Sender: h.EmailManager.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: i.Recipient,
			},
		},
	}

	var err error
	if data.InviteURL, err = h.EmailManager.InviteURL(i.Token); err != nil {
		return err
	}

	msg, err := emails.InviteEmail(data)
	if err != nil {
		return err
	}

	return h.EmailManager.Send(msg)
}
