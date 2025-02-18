package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
)

// HookInvite runs on invite create mutations
func HookInvite() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.InviteFunc(func(ctx context.Context, m *generated.InviteMutation) (generated.Value, error) {
			m, err := setRequestor(ctx, m)
			if err != nil {
				log.Error().Err(err).Msg("unable to determine requestor")

				return nil, err
			}

			// validate the invite
			if err := validateCanCreateInvite(ctx, m); err != nil {
				log.Info().Err(err).Msg("unable to add user to specified organization")

				return nil, err
			}

			// generate token based on recipient + target org ID
			m, err = setRecipientAndToken(m)
			if err != nil {
				log.Error().Err(err).Msg("unable to create verification token")

				return nil, err
			}

			// attempt to do the mutation for a new user invite
			var retValue ent.Value

			// check if the invite already exists
			existingInvite, err := getInvite(ctx, m)

			// if the invite exists, update the token and resend
			if existingInvite != nil && err == nil {
				log.Info().Msg("invitation for user already exists")

				// update invite instead
				retValue, err = updateInvite(ctx, m)
				if err != nil {
					log.Error().Err(err).Msg("unable to update invitation")

					return retValue, err
				}
			} else {
				// create new invite
				retValue, err = next.Mutate(ctx, m)
				if err != nil {
					return retValue, err
				}
			}

			// queue the email to be sent
			if err := createInviteToSend(ctx, m); err != nil {
				log.Error().Err(err).Msg("error sending email to user")
			}

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
					log.Error().Err(err).Msg("unable to get existing invite")

					return nil, err
				}

				ownerID = invite.OwnerID
				role = invite.Role
				recipient = invite.Recipient
			}

			// user must be authenticated to accept an invite, get their id from the context
			userID, err := auth.GetUserIDFromContext(ctx)
			if err != nil {
				log.Error().Err(err).Msg("unable to get user to add to organization")

				return nil, err
			}

			input := generated.CreateOrgMembershipInput{
				UserID:         userID,
				OrganizationID: ownerID,
				Role:           &role,
			}

			// add user to the inviting org, allow the context to bypass privacy checks
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
			if _, err := m.Client().OrgMembership.Create().SetInput(input).Save(allowCtx); err != nil {
				log.Error().Err(err).Msg("unable to add user to organization")

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
				log.Error().Err(err).Msg("unable to get organization")

				return retValue, err
			}

			invite := emailtemplates.InviteTemplateData{
				OrganizationName: org.Name,
				Role:             string(role),
			}

			if err := createOrgInviteAcceptedToSend(ctx, m, recipient, invite); err != nil {
				return retValue, err
			}

			// delete the invite that has been accepted
			if err := deleteInvite(ctx, m); err != nil {
				log.Error().Err(err).Msg("unable to delete invite")

				return retValue, err
			}

			return retValue, err
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// validateCanCreateInvite checks if the mutation is for a personal org and denies if true or
// if the user does not have access to that organization
func validateCanCreateInvite(ctx context.Context, m *generated.InviteMutation) error {
	orgID, ok := m.OwnerID()
	if !ok {
		return nil
	}

	org, err := m.Client().Organization.Query().
		WithSetting().
		Where(organization.ID(orgID)).
		Only(ctx)
	if err != nil {
		return err
	}

	if org.PersonalOrg {
		return ErrPersonalOrgsNoChildren
	}

	// check if the the email can be invited to the organization
	email, _ := m.Recipient()

	return checkAllowedEmailDomain(email, org.Edges.Setting)
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
	emailAddress, _ := m.Recipient()
	role, _ := m.Role()

	org, err := m.Client().Organization.Get(ctx, orgID)
	if err != nil {
		return err
	}

	requestor, err := m.Client().User.Get(ctx, reqID)
	if err != nil {
		return err
	}

	invite := emailtemplates.InviteTemplateData{
		InviterName:      requestor.FirstName,
		OrganizationName: org.Name,
		Role:             string(role),
	}

	email, err := m.Emailer.NewInviteEmail(emailtemplates.Recipient{
		Email: emailAddress,
	}, invite, token)
	if err != nil {
		log.Error().Err(err).Msg("error rendering email")

		return err
	}

	// send the email
	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
}

// createOrgInviteAcceptedToSend composes the email metadata and queues the email to be sent
func createOrgInviteAcceptedToSend(ctx context.Context, m *generated.InviteMutation, recipient string, i emailtemplates.InviteTemplateData) error {
	email, err := m.Emailer.NewInviteAcceptedEmail(emailtemplates.Recipient{
		Email: recipient,
	}, i)
	if err != nil {
		log.Error().Err(err).Msg("error rendering email")

		return err
	}

	// send the email
	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		log.Error().Err(err).Msg("error queueing email verification")

		return err
	}

	return nil
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

	return m.Client().Invite.DeleteOneID(id).Exec(ctx)
}

func getInvite(ctx context.Context, m *generated.InviteMutation) (*generated.Invite, error) {
	rec, _ := m.Recipient()
	ownerID, _ := m.OwnerID()

	return m.Client().Invite.Query().Where(invite.Recipient(rec)).Where(invite.OwnerID(ownerID)).Only(ctx)
}

// checkAllowedEmailDomain checks if the email domain is allowed for the organization
func checkAllowedEmailDomain(email string, orgSetting *generated.OrganizationSetting) error {
	if orgSetting == nil || email == "" {
		log.Info().Msg("no organization setting or email provided, cannot check settings")

		return nil
	}

	// allow all domains if none are set
	if orgSetting.AllowedEmailDomains == nil {
		return nil
	}

	emailDomain := strings.SplitAfter(email, "@")[1]

	allowed := false

	for _, domain := range orgSetting.AllowedEmailDomains {
		if domain == emailDomain {
			allowed = true
			break
		}
	}

	if !allowed {
		return ErrEmailDomainNotAllowed
	}

	return nil
}
