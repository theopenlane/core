package handlers

import (
	"context"
	"time"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailverificationtoken"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/passwordresettoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/generated/webauthn"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// updateUserLastSeen updates the last seen timestamp of the user
func (h *Handler) updateUserLastSeen(ctx context.Context, id string) error {
	if _, err := transaction.FromContext(ctx).
		User.
		UpdateOneID(id).
		SetLastSeen(time.Now()).
		Save(ctx); err != nil {
		h.Logger.Errorw("error updating user last seen", "error", err)

		return err
	}

	return nil
}

// createUser creates a user in the database based on the input and returns the user with user settings
func (h *Handler) createUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error) {
	meowuser, err := transaction.FromContext(ctx).User.Create().
		SetInput(input).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error creating new user", "error", err)

		return nil, err
	}

	return meowuser, nil
}

// updateSubscriberVerifiedEmail updates a subscriber by in the database based on the input and sets to active with verified email
func (h *Handler) updateSubscriberVerifiedEmail(ctx context.Context, id string, input ent.UpdateSubscriberInput) error {
	_, err := transaction.FromContext(ctx).Subscriber.UpdateOneID(id).
		SetInput(input).
		SetActive(true).
		SetVerifiedEmail(true).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error updating subscriber verified", "error", err)
		return err
	}

	return nil
}

// updateSubscriber updates a subscriber by in the database based on the input
func (h *Handler) updateSubscriberVerificationToken(ctx context.Context, user *User) error {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		h.Logger.Errorw("unable to parse ttl", "error", err)
		return err
	}

	_, err = transaction.FromContext(ctx).Subscriber.UpdateOneID(user.ID).
		SetToken(user.EmailVerificationToken.String).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(ttl).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error updating subscriber tokens", "error", err)

		return err
	}

	return nil
}

// createEmailVerificationToken creates a new email verification for the user
func (h *Handler) createEmailVerificationToken(ctx context.Context, user *User) (*ent.EmailVerificationToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		h.Logger.Errorw("unable to parse ttl", "error", err)
		return nil, err
	}

	meowtoken, err := transaction.FromContext(ctx).EmailVerificationToken.Create().
		SetOwnerID(user.ID).
		SetToken(user.EmailVerificationToken.String).
		SetTTL(ttl).
		SetEmail(user.Email).
		SetSecret(user.EmailVerificationSecret).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error creating email verification token", "error", err)

		return nil, err
	}

	return meowtoken, nil
}

func (h *Handler) createPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		h.Logger.Errorw("unable to parse ttl", "error", err)
		return nil, err
	}

	meowtoken, err := transaction.FromContext(ctx).PasswordResetToken.Create().
		SetOwnerID(user.ID).
		SetToken(user.PasswordResetToken.String).
		SetTTL(ttl).
		SetEmail(user.Email).
		SetSecret(user.PasswordResetSecret).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error creating password reset token", "error", err)

		return nil, err
	}

	return meowtoken, nil
}

// getUserByEVToken returns the ent user with the user settings and email verification token fields based on the
// token in the request
func (h *Handler) getUserByEVToken(ctx context.Context, token string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).EmailVerificationToken.Query().
		Where(
			emailverificationtoken.Token(token),
		).
		QueryOwner().WithSetting().WithEmailVerificationTokens().Only(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining user from email verification token", "error", err)

		return nil, err
	}

	return user, nil
}

// getUserByResetToken returns the ent user with the user settings and password reset tokens based on the
// token in the request
func (h *Handler) getUserByResetToken(ctx context.Context, token string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).PasswordResetToken.Query().
		Where(
			passwordresettoken.Token(token),
		).
		QueryOwner().WithSetting().WithPasswordResetTokens().Only(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining user from reset token", "error", err)

		return nil, err
	}

	return user, nil
}

// getUserByEmail returns the ent user with the user settings based on the email and auth provider in the request
func (h *Handler) getUserByEmail(ctx context.Context, email string, authProvider enums.AuthProvider) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.Email(email)).
		Where(user.AuthProviderEQ(authProvider)).
		Only(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining user from email", "error", err)

		return nil, err
	}

	return user, nil
}

// getUserByID returns the ent user with the user settings based on the email and auth provider in the request
func (h *Handler) getUserByID(ctx context.Context, id string, authProvider enums.AuthProvider) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.ID(id)).
		Where(user.AuthProviderEQ(authProvider)).
		Only(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining user from id", "error", err)

		return nil, err
	}

	// Add webauthn to the response
	user.Edges.Webauthn = user.QueryWebauthn().AllX(ctx)

	return user, nil
}

// addCredentialToUser adds a new webauthn credential to the user
func (h *Handler) addCredentialToUser(ctx context.Context, user *ent.User, credential gowebauthn.Credential) error {
	transports := []string{}
	for _, t := range credential.Transport {
		transports = append(transports, string(t))
	}

	count, err := transaction.FromContext(ctx).Webauthn.Query().Where(
		webauthn.OwnerID(user.ID),
	).Count(ctx)
	if err != nil {
		h.Logger.Errorw("error checking existing webauthn credentials", "error", err)

		return err
	}

	if count >= h.OauthProvider.Webauthn.MaxDevices {
		h.Logger.Errorw("max devices reached", "error", err)

		return ErrMaxDeviceLimit
	}

	_, err = transaction.FromContext(ctx).Webauthn.Create().
		SetOwnerID(user.ID).
		SetTransports(transports).
		SetAttestationType(credential.AttestationType).
		SetAaguid(credential.Authenticator.AAGUID).
		SetCredentialID(credential.ID).
		SetPublicKey(credential.PublicKey).
		SetBackupState(credential.Flags.BackupEligible).
		SetBackupEligible(credential.Flags.BackupEligible).
		SetUserPresent(credential.Flags.UserPresent).
		SetUserVerified(credential.Flags.UserVerified).
		SetSignCount(int32(credential.Authenticator.SignCount)).
		Save(ctx)
	if err != nil {
		h.Logger.Errorw("error creating email verification token", "error", err)

		return err
	}

	return nil
}

// getUserDetailsByID returns the ent user with the user settings based on the user ID
func (h *Handler) getUserDetailsByID(ctx context.Context, userID string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().Where(
		user.ID(userID),
	).Only(ctx)
	if err != nil {
		h.Logger.Errorf("error retrieving user", "error", err)

		return nil, err
	}

	return user, nil
}

// getUserByInviteToken returns the ent user based on the invite token in the request
func (h *Handler) getUserByInviteToken(ctx context.Context, token string) (*ent.Invite, error) {
	recipient, err := transaction.FromContext(ctx).Invite.Query().
		Where(
			invite.Token(token),
		).WithOwner().Only(ctx)

	if err != nil {
		h.Logger.Errorw("error obtaining user from token", "error", err)

		return nil, err
	}

	return recipient, err
}

// countVerificationTokensUserByEmail counts number of existing email verification attempts before issuing a new one
func (h *Handler) countVerificationTokensUserByEmail(ctx context.Context, email string) (int, error) {
	attempts, err := transaction.FromContext(ctx).EmailVerificationToken.Query().WithOwner().Where(
		emailverificationtoken.And(
			emailverificationtoken.Email(email),
		)).Count(ctx)
	if err != nil {
		h.Logger.Errorw("error counting verification reset tokens", "error", err)

		return 0, err
	}

	return attempts, nil
}

// expireAllVerificationTokensUserByEmail expires all existing email verification tokens before issuing a new one
func (h *Handler) expireAllVerificationTokensUserByEmail(ctx context.Context, email string) error {
	prs, err := transaction.FromContext(ctx).EmailVerificationToken.Query().WithOwner().Where(
		emailverificationtoken.And(
			emailverificationtoken.Email(email),
			emailverificationtoken.TTLGT(time.Now()),
		)).All(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining verification reset tokens", "error", err)

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			h.Logger.Errorw("error expiring verification token", "error", err)

			return err
		}
	}

	return nil
}

// expireAllResetTokensUserByEmail expires all existing password reset tokens before issuing a new one
func (h *Handler) expireAllResetTokensUserByEmail(ctx context.Context, email string) error {
	prs, err := transaction.FromContext(ctx).PasswordResetToken.Query().WithOwner().Where(
		passwordresettoken.And(
			passwordresettoken.Email(email),
			passwordresettoken.TTLGT(time.Now()),
		)).All(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining password reset tokens", "error", err)

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			h.Logger.Errorw("error expiring password reset token", "error", err)

			return err
		}
	}

	return nil
}

// setEmailConfirmed sets the user setting field email_confirmed to true within a transaction
func (h *Handler) setEmailConfirmed(ctx context.Context, user *ent.User) error {
	if _, err := transaction.FromContext(ctx).UserSetting.Update().SetEmailConfirmed(true).
		Where(
			usersetting.ID(user.Edges.Setting.ID),
		).Save(ctx); err != nil {
		h.Logger.Errorw("error setting email confirmed", "error", err)

		return err
	}

	return nil
}

// updateUserPassword changes a updates a user's password in the database
func (h *Handler) updateUserPassword(ctx context.Context, id string, password string) error {
	if _, err := transaction.FromContext(ctx).User.UpdateOneID(id).SetPassword(password).Save(ctx); err != nil {
		h.Logger.Errorw("error updating user password", "error", err)

		return err
	}

	return nil
}

// addDefaultOrgToUserQuery adds the default org to the user object, user must be authenticated before calling this
func (h *Handler) addDefaultOrgToUserQuery(ctx context.Context, user *ent.User) error {
	// get the default org for the user, allow access, accessible orgs will be filtered by the interceptor
	orgCtx := privacy.DecisionContext(ctx, privacy.Allow)

	org, err := user.Edges.Setting.DefaultOrg(orgCtx)
	if err != nil {
		h.Logger.Errorw("error obtaining default org", "error", err)

		return err
	}

	// add default org to user object
	user.Edges.Setting.Edges.DefaultOrg = org

	return nil
}

// CheckAndCreateUser takes a user with an OauthTooToken set in the context and checks if the user is already created
// if the user already exists, update last seen
func (h *Handler) CheckAndCreateUser(ctx context.Context, name, email string, provider enums.AuthProvider, image string) (*ent.User, error) {
	// check if users exists
	entUser, err := h.getUserByEmail(ctx, email, provider)
	if err != nil {
		// if the user is not found, create now
		if ent.IsNotFound(err) {
			// create the input based on the provider
			input := createUserInput(name, email, provider, image)

			// create user in the database
			entUser, err = h.createUser(ctx, input)
			if err != nil {
				h.Logger.Errorw("error creating new user", "error", err)

				return nil, err
			}

			// return newly created user
			return entUser, nil
		}

		return nil, err
	}

	// update last seen of user
	if err := h.updateUserLastSeen(ctx, entUser.ID); err != nil {
		h.Logger.Errorw("unable to update last seen", "error", err)

		return nil, err
	}

	// update user avatar
	if err := h.updateUserAvatar(ctx, entUser, image); err != nil {
		h.Logger.Errorw("error updating user avatar", "error", err)

		return nil, err
	}

	return entUser, nil
}

// createUserInput creates a new user input based on the name, email, image and provider
func createUserInput(name, email string, provider enums.AuthProvider, image string) ent.CreateUserInput {
	lastSeen := time.Now().UTC()

	// create new user input
	input := parseName(name)
	input.Email = email
	input.AuthProvider = &provider
	input.LastSeen = &lastSeen

	if image != "" {
		input.AvatarRemoteURL = &image
	}

	return input
}

func (h *Handler) updateUserAvatar(ctx context.Context, user *ent.User, image string) error {
	if image == "" {
		return nil
	}

	if user.AvatarRemoteURL != nil && *user.AvatarRemoteURL == image {
		return nil
	}

	if _, err := transaction.FromContext(ctx).
		User.UpdateOneID(user.ID).
		SetAvatarRemoteURL(image).
		Save(ctx); err != nil {
		h.Logger.Errorw("error updating user avatar", "error", err)
		return err
	}

	return nil
}

// setWebauthnAllowed sets the user setting field is_webauthn_allowed to true within a transaction
func (h *Handler) setWebauthnAllowed(ctx context.Context, user *ent.User) error {
	if _, err := transaction.FromContext(ctx).UserSetting.Update().SetIsWebauthnAllowed(true).
		Where(
			usersetting.UserID(user.ID),
		).Save(ctx); err != nil {
		h.Logger.Errorw("error setting webauthn allowed", "error", err)

		return err
	}

	return nil
}

// getSubscriberByToken returns the subscriber based on the token in the request
func (h *Handler) getSubscriberByToken(ctx context.Context, token string) (*ent.Subscriber, error) {
	subscriber, err := transaction.FromContext(ctx).Subscriber.Query().
		Where(
			subscriber.Token(token),
		).
		Only(ctx)
	if err != nil {
		h.Logger.Errorw("error obtaining subscriber from verification token", "error", err)

		return nil, err
	}

	return subscriber, nil
}

// getOrgByID returns the organization based on the id in the request
func (h *Handler) getOrgByID(ctx context.Context, id string) (*ent.Organization, error) {
	org, err := transaction.FromContext(ctx).Organization.Get(ctx, id)
	if err != nil {
		h.Logger.Errorw("error obtaining organization from id", "error", err)

		return nil, err
	}

	return org, nil
}
