package hooks

import (
	"context"

	"entgo.io/ent"

	ph "github.com/posthog/posthog-go"

	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/sendgrid"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/enums"
)

// HookInvite runs on invite create mutations
func HookInvite() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InviteFunc(func(ctx context.Context, m *generated.InviteMutation) (generated.Value, error) {
			m, err := setRequestor(ctx, m)
			if err != nil {
				m.Logger.Errorw("unable to determine requestor")

				return nil, err
			}

			// check that the invite isn't to a personal organization
			if err := personalOrgNoInvite(ctx, m); err != nil {
				m.Logger.Infow("unable to add user to specified organization", "error", err)

				return nil, err
			}

			// generate token based on recipient + target org ID
			m, err = setRecipientAndToken(m)
			if err != nil {
				m.Logger.Errorw("error creating verification token", "error", err)

				return nil, err
			}

			// check if the invite already exists
			existingInvite, err := getInvite(ctx, m)

			// attempt to do the mutation for a new user invite
			var retValue ent.Value

			// if the invite exists, update the token and resend
			if existingInvite != nil && err == nil {
				m.Logger.Infow("invitation for user already exists")

				// update invite instead
				retValue, err = updateInvite(ctx, m)
				if err != nil {
					m.Logger.Errorw("unable to update invitation", "error", err)

					return retValue, err
				}
			} else {
				// create new invite
				retValue, err = next.Mutate(ctx, m)
				if err != nil {
					return retValue, err
				}
			}

			// non-blocking queued email
			if err := createInviteToSend(ctx, m); err != nil {
				m.Logger.Errorw("error sending email to user", "error", err)
			}

			orgID, _ := m.OwnerID()
			org, _ := m.Client().Organization.Get(ctx, orgID)
			reqID, _ := m.RequestorID()
			requestor, _ := m.Client().User.Get(ctx, reqID)
			email, _ := m.Recipient()
			role, _ := m.Role()

			props := ph.NewProperties().
				Set("organization_id", orgID).
				Set("organization_name", org.Name).
				Set("requestor_id", reqID).
				Set("recipient_email", email).
				Set("recipient_role", role)

			// if we have the requestor, add their name and email to the properties
			if requestor != nil {
				props.Set("requestor_name", requestor.FirstName).
					Set("requestor_email", requestor.Email)
			}

			m.Analytics.Event("organization_invite_created", props)

			return retValue, err
		})
	}, ent.OpCreate)
}

// HookInviteAccepted adds the user to the organization when the status is accepted
func HookInviteAccepted() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InviteFunc(func(ctx context.Context, m *generated.InviteMutation) (generated.Value, error) {
			status, ok := m.Status()
			if !ok || status != enums.InvitationAccepted {
				// nothing to do here
				return next.Mutate(ctx, m)
			}

			ownerID, ownerOK := m.OwnerID()
			role, roleOK := m.Role()
			recipient, recipientOK := m.Recipient()

			// if we are missing any, get them from the db
			// this should happen on an update mutation
			if !ownerOK || !roleOK || !recipientOK {
				id, _ := m.ID()

				invite, err := m.Client().Invite.Get(ctx, id)
				if err != nil {
					m.Logger.Errorw("unable to get existing invite", "error", err)

					return nil, err
				}

				ownerID = invite.OwnerID
				role = invite.Role
				recipient = invite.Recipient
			}

			// user must be authenticated to accept an invite, get their id from the context
			userID, err := auth.GetUserIDFromContext(ctx)
			if err != nil {
				m.Logger.Errorw("unable to get user to add to organization", "error", err)

				return nil, err
			}

			input := generated.CreateOrgMembershipInput{
				UserID:         userID,
				OrganizationID: ownerID,
				Role:           &role,
			}

			// add user to the inviting org
			if _, err := m.Client().OrgMembership.Create().SetInput(input).Save(ctx); err != nil {
				m.Logger.Errorw("unable to add user to organization", "error", err)

				return nil, err
			}

			// finish the mutation
			retValue, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// fetch org details to pass the name in the email
			org, err := m.Client().Organization.Query().Clone().Where(organization.ID(ownerID)).Only(ctx)
			if err != nil {
				m.Logger.Errorw("unable to get organization", "error", err)

				return retValue, err
			}

			invite := &emails.Invite{
				OrgName:   org.Name,
				Recipient: recipient,
				Role:      string(role),
			}

			props := ph.NewProperties().
				Set("organization_id", org.ID).
				Set("organization_name", org.Name).
				Set("acceptor_email", recipient).
				Set("acceptor_id", userID)

			m.Analytics.Event("organization_invite_accepted", props)

			// send an email to recipient notifying them they've been added to an organization
			if err := m.Marionette.Queue(marionette.TaskFunc(func(ctx context.Context) error {
				return sendOrgAccepted(ctx, m, invite)
			}), marionette.WithErrorf("could not send invitation email to user %s", recipient),
			); err != nil {
				m.Logger.Errorw("unable to queue email for sending")

				return retValue, err
			}

			// delete the invite that has been accepted
			if err := deleteInvite(ctx, m); err != nil {
				return retValue, err
			}

			return retValue, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// personalOrgNoInvite checks if the mutation is for a personal org and denies if true or
// if the user does not have access to that organization
func personalOrgNoInvite(ctx context.Context, m *generated.InviteMutation) error {
	orgID, ok := m.OwnerID()
	if ok {
		org, err := m.Client().Organization.Get(ctx, orgID)
		if err != nil {
			return err
		}

		if org.PersonalOrg {
			return ErrPersonalOrgsNoChildren
		}
	}

	return nil
}

// setRecipientAndToken function is responsible for generating a invite token based on the
// recipient's email and the target organization ID
func setRecipientAndToken(m *generated.InviteMutation) (*generated.InviteMutation, error) {
	email, ok := m.Recipient()
	if !ok || email == "" {
		return nil, ErrMissingRecipientEmail
	}

	owner, _ := m.OwnerID()

	oid, err := ulids.Parse(owner)
	if err != nil {
		return nil, err
	}

	verify, err := tokens.NewOrgInvitationToken(email, oid)
	if err != nil {
		return nil, err
	}

	token, secret, err := verify.Sign()
	if err != nil {
		return nil, err
	}

	// set values on mutation
	m.SetToken(token)
	m.SetExpires(verify.ExpiresAt)
	m.SetSecret(secret)

	return m, nil
}

// setRequestor sets the requestor on the mutation
func setRequestor(ctx context.Context, m *generated.InviteMutation) (*generated.InviteMutation, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		m.Logger.Errorw("unable to get requestor", "error", err)

		return m, err
	}

	m.SetRequestorID(userID)

	return m, nil
}

// createInviteToSend sets the necessary data to send invite email + token
func createInviteToSend(ctx context.Context, m *generated.InviteMutation) error {
	// these are all required fields on create so should be found
	orgID, _ := m.OwnerID()
	reqID, _ := m.RequestorID()
	token, _ := m.Token()
	email, _ := m.Recipient()
	role, _ := m.Role()

	org, err := m.Client().Organization.Get(ctx, orgID)
	if err != nil {
		return err
	}

	requestor, err := m.Client().User.Get(ctx, reqID)
	if err != nil {
		return err
	}

	invite := &emails.Invite{
		OrgName:   org.Name,
		Token:     token,
		Requestor: requestor.FirstName,
		Recipient: email,
		Role:      string(role),
	}

	if err := m.Marionette.Queue(marionette.TaskFunc(func(ctx context.Context) error {
		return sendOrgInvitationEmail(ctx, m, invite)
	}), marionette.WithErrorf("could not send invitation email to user %s", email),
	); err != nil {
		m.Logger.Errorw("unable to queue email for sending")

		return err
	}

	return nil
}

// sendOrgInvitationEmail composes the email metadata and sends via email manager
func sendOrgInvitationEmail(ctx context.Context, m *generated.InviteMutation, i *emails.Invite) (err error) {
	data := emails.InviteData{
		InviterName: i.Requestor,
		OrgName:     i.OrgName,
		EmailData: emails.EmailData{
			Sender: m.Emails.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: i.Recipient,
			},
		},
	}

	if data.InviteURL, err = m.Emails.InviteURL(i.Token); err != nil {
		return err
	}

	msg, err := emails.InviteEmail(data)
	if err != nil {
		return err
	}

	return m.Emails.Send(msg)
}

// sendOrgAccepted composes the email metadata to notify the user they've been joined to the org
func sendOrgAccepted(ctx context.Context, m *generated.InviteMutation, i *emails.Invite) (err error) {
	data := emails.InviteData{
		InviterName: i.Requestor,
		OrgName:     i.OrgName,
		EmailData: emails.EmailData{
			Sender: m.Emails.MustFromContact(),
			Recipient: sendgrid.Contact{
				Email: i.Recipient,
			},
		},
	}

	msg, err := emails.InviteAccepted(data)
	if err != nil {
		return err
	}

	return m.Emails.Send(msg)
}

var maxAttempts = 5

// updateInvite if the invite already exists, set a new token, secret, expiration, and increment the attempts
// error at max attempts to resend
func updateInvite(ctx context.Context, m *generated.InviteMutation) (*generated.Invite, error) {
	// get the existing invite by recipient and owner
	rec, _ := m.Recipient()
	ownerID, _ := m.OwnerID()

	invite, err := m.Client().Invite.Query().Where(invite.Recipient(rec)).Where(invite.OwnerID(ownerID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	// create update mutation
	if invite.SendAttempts >= maxAttempts {
		return nil, ErrMaxAttempts
	}

	// increment attempts
	invite.SendAttempts++

	m.SetSendAttempts(invite.SendAttempts)

	// these were already set when the invite was attempted to be added
	// we do not need to create these again
	secret, _ := m.Secret()
	token, _ := m.Token()
	expiresAt, _ := m.Expires()

	// update the invite
	return m.Client().Invite.
		UpdateOneID(invite.ID).
		SetSendAttempts(invite.SendAttempts).
		SetToken(token).
		SetExpires(expiresAt).
		SetSecret(secret).
		Save(ctx)
}

// deleteInvite deletes an invite from the database
func deleteInvite(ctx context.Context, m *generated.InviteMutation) error {
	id, _ := m.ID()

	if err := m.Client().Invite.DeleteOneID(id).Exec(ctx); err != nil {
		m.Logger.Errorw("unable to delete invite", "error", err)

		return err
	}

	return nil
}

func getInvite(ctx context.Context, m *generated.InviteMutation) (*generated.Invite, error) {
	rec, _ := m.Recipient()
	ownerID, _ := m.OwnerID()

	return m.Client().Invite.Query().Where(invite.Recipient(rec)).Where(invite.OwnerID(ownerID)).Only(ctx)
}
