package handlers

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/oklog/ulid/v2"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// Invite holds the Token, InviteToken references, and the additional user input to complete acceptance of the invitation
type Invite struct {
	Token     string
	UserID    ulid.ULID
	Email     string
	DestOrgID ulid.ULID
	Role      enums.Role
	InviteToken
}

// InviteToken holds data specific to a future user of the system for invite logic
type InviteToken struct {
	Expires sql.NullString
	Token   sql.NullString
	Secret  []byte
}

// OrganizationInviteAccept is responsible for handling the invitation of a user to an organization.
// It receives a request with the user's invitation details, validates the request,
// and creates organization membership for the user
// On success, it returns a response with the organization information
func (h *Handler) OrganizationInviteAccept(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleInviteRequest, models.ExampleInviteResponse, openapi.Registry)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// get the authenticated user from the context
	userID, err := auth.GetSubjectIDFromContext(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err, openapi)
	}

	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving user details")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	ctxWithToken, user, invitedUser, err := h.processInvitation(ctx, in.Token, user.Email)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSessionWithOrg(ctxWithToken, ctx.Response().Writer, user, invitedUser.OwnerID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// reply with the relevant details
	out := &models.InviteReply{
		Reply:       rout.Reply{Success: true},
		ID:          userID,
		Email:       invitedUser.Recipient,
		JoinedOrgID: invitedUser.OwnerID,
		Role:        string(invitedUser.Role),
		Message:     "Welcome to your new organization!",
		AuthData:    *auth,
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	status, err := h.fetchSSOStatus(allowCtx, invitedUser.OwnerID, "")

	if err == nil && status.Enforced {
		out.NeedsSSO = true
	}

	return h.Created(ctx, out)
}

func (h *Handler) processInvitation(ctx echo.Context, invitationToken, userEmail string) (context.Context, *generated.User, *generated.Invite, error) {
	inv := &Invite{
		Token: invitationToken,
	}

	// ensure the user that is logged in, matches the invited user
	if err := inv.validateInviteRequest(); err != nil {
		return nil, nil, nil, err
	}

	reqCtx := ctx.Request().Context()

	// set the initial context based on the token
	ctxWithToken := token.NewContextWithOrgInviteToken(reqCtx, inv.Token)

	// fetch the recipient and org owner based on token
	invitedUser, err := h.getUserByInviteToken(ctxWithToken, inv.Token)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, nil, nil, err
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving invite token")

		return nil, nil, nil, err
	}

	// check if we already have an authenticated user in the context (from registration flow)
	au, err := auth.GetAuthenticatedUserFromContext(ctxWithToken)
	if err != nil {
		return nil, nil, nil, err
	}

	// update the authenticated user context with the organization ID
	// since the hook requires the org id
	au.OrganizationID = invitedUser.OwnerID
	au.OrganizationIDs = []string{invitedUser.OwnerID}

	ctxWithToken = auth.WithAuthenticatedUser(ctxWithToken, au)

	// add email to the invite
	inv.Email = invitedUser.Recipient

	// get user details for logged in user
	user, err := h.getUserByEmail(reqCtx, userEmail)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user for request")

		return nil, nil, nil, err
	}

	// ensure the user that is logged in, matches the invited user
	if err := inv.validateUser(user.Email); err != nil {
		return nil, nil, nil, err
	}

	// string to ulid so we can match the token input
	oid, err := ulid.Parse(invitedUser.OwnerID)
	if err != nil {
		return nil, nil, nil, err
	}

	uid, err := ulid.Parse(user.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	// construct the invite details but set email to the original recipient, and the joining organization ID as the current owner of the invitation
	invite := &Invite{
		Email:     invitedUser.Recipient,
		UserID:    uid,
		DestOrgID: oid,
		Role:      invitedUser.Role,
	}

	// set tokens for request
	if err := invite.setOrgInviteTokens(invitedUser, inv.Token); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to set invite token for request")

		return nil, nil, nil, err
	}

	// reconstruct the token based on recipient & owning organization so we can compare it to the one were receiving
	t := &tokens.OrgInviteToken{
		Email: invitedUser.Recipient,
		OrgID: oid,
	}

	// check and ensure the token has not expired
	if t.ExpiresAt, err = invite.GetInviteExpires(); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to parse expiration")

		return nil, nil, nil, err
	}

	// Verify the token is valid with the stored secret
	if err = t.Verify(invite.GetInviteToken(), invite.Secret); err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			if err := updateInviteStatusExpired(ctxWithToken, invitedUser); err != nil {
				return nil, nil, nil, err
			}

			return nil, nil, nil, tokens.ErrTokenExpired
		}

		return nil, nil, nil, err
	}

	if err := updateInviteStatusAccepted(ctxWithToken, invitedUser); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update invite status")

		return nil, nil, nil, err
	}

	return ctxWithToken, user, invitedUser, nil
}

// validateInviteRequest is a helper function that validates the required fields are set in the user request
func (i *Invite) validateInviteRequest() error {
	// ensure the token is set
	if i.Token == "" {
		return rout.NewMissingRequiredFieldError("token")
	}

	return nil
}

// validateUser is a helper function that ensures the logged-in user is the same as the invite
func (i *Invite) validateUser(email string) error {
	// ensure the logged in user is the same as the invite
	if i.Email != email {
		return ErrUnableToVerifyEmail
	}

	return nil
}

// GetInviteToken returns the invitation token if it's valid
func (i *Invite) GetInviteToken() string {
	if i.InviteToken.Token.Valid {
		return i.InviteToken.Token.String
	}

	return ""
}

// GetInviteExpires returns the expiration time of the invite token
func (i *Invite) GetInviteExpires() (time.Time, error) {
	if i.Expires.Valid {
		return time.Parse(time.RFC3339Nano, i.Expires.String)
	}

	return time.Time{}, nil
}

// setOrgInviteTokens sets the fields of the `Invite` struct to verify the email
// invitation. It takes in an `Invite` object and an invitation token as parameters. If
// the invitation token matches the token stored in the `Invite` object, it sets the
// `Token`, `Secret`, and `Expires` fields of the `InviteToken` struct. This allows the
// token to be verified later when the user accepts the invitation
func (i *Invite) setOrgInviteTokens(inv *generated.Invite, invToken string) error {
	if inv.Token == invToken {
		i.InviteToken.Token = sql.NullString{String: inv.Token, Valid: true}
		i.Secret = *inv.Secret
		i.Expires = sql.NullString{String: inv.Expires.Format(time.RFC3339Nano), Valid: true}

		return nil
	}

	return ErrNotFound
}

// updateInviteStatusAccepted updates the status of an invite to "Accepted"
func updateInviteStatusAccepted(ctx context.Context, i *generated.Invite) error {
	return transaction.FromContext(ctx).Invite.UpdateOneID(i.ID).SetStatus(enums.InvitationAccepted).Exec(ctx)
}

// updateInviteStatusExpired updates the status of an invite to "Expired"
func updateInviteStatusExpired(ctx context.Context, i *generated.Invite) error {
	return transaction.FromContext(ctx).Invite.UpdateOneID(i.ID).SetStatus(enums.InvitationExpired).Exec(ctx)
}
