package handlers

import (
	"context"
	"time"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailverificationtoken"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
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
		log.Error().Err(err).Msg("error updating user last seen")

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
		log.Error().Err(err).Msg("error creating new user")

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
		log.Error().Err(err).Msg("error updating subscriber verified")
		return err
	}

	return nil
}

// updateSubscriber updates a subscriber by in the database based on the input
func (h *Handler) updateSubscriberVerificationToken(ctx context.Context, user *User) error {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse ttl")
		return err
	}

	_, err = transaction.FromContext(ctx).Subscriber.UpdateOneID(user.ID).
		SetToken(user.EmailVerificationToken.String).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(ttl).
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error updating subscriber tokens")

		return err
	}

	return nil
}

// createEmailVerificationToken creates a new email verification for the user
func (h *Handler) createEmailVerificationToken(ctx context.Context, user *User) (*ent.EmailVerificationToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse ttl")
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
		log.Error().Err(err).Msg("error creating email verification token")

		return nil, err
	}

	return meowtoken, nil
}

func (h *Handler) createPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse ttl")
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
		log.Error().Err(err).Msg("error creating password reset token")

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
		log.Error().Err(err).Msg("error obtaining user from email verification token")

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
		log.Error().Err(err).Msg("error obtaining user from reset token")

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
		log.Error().Err(err).Msg("error obtaining user from email")

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
		log.Error().Err(err).Msg("error obtaining user from id")

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
		log.Error().Err(err).Msg("error checking existing webauthn credentials")

		return err
	}

	if count >= h.OauthProvider.Webauthn.MaxDevices {
		log.Error().Err(err).Msg("max devices reached")

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
		SetSignCount(int32(credential.Authenticator.SignCount)). // nolint:gosec
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error creating email verification token")

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
		log.Error().Err(err).Msg("error retrieving user")

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
		log.Error().Err(err).Msg("error obtaining user from token")

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
		log.Error().Err(err).Msg("error counting verification reset tokens")

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
		log.Error().Err(err).Msg("error obtaining verification reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			log.Error().Err(err).Msg("error expiring verification token")

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
		log.Error().Err(err).Msg("error obtaining password reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			log.Error().Err(err).Msg("error expiring password reset token")

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
		log.Error().Err(err).Msg("error setting email confirmed")

		return err
	}

	return nil
}

// updateUserPassword changes a updates a user's password in the database
func (h *Handler) updateUserPassword(ctx context.Context, id string, password string) error {
	if _, err := transaction.FromContext(ctx).User.UpdateOneID(id).SetPassword(password).Save(ctx); err != nil {
		log.Error().Err(err).Msg("error updating user password")

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
		log.Error().Err(err).Msg("error obtaining default org")

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
				log.Error().Err(err).Msg("error creating new user")

				return nil, err
			}

			// return newly created user
			return entUser, nil
		}

		return nil, err
	}

	// update last seen of user
	if err := h.updateUserLastSeen(ctx, entUser.ID); err != nil {
		log.Error().Err(err).Msg("error updating user last seen")

		return nil, err
	}

	// update user avatar
	if err := h.updateUserAvatar(ctx, entUser, image); err != nil {
		log.Error().Err(err).Msg("error updating user avatar")

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
		log.Error().Err(err).Msg("error updating user avatar")
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
		log.Error().Err(err).Msg("error setting webauthn allowed")

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
		log.Error().Err(err).Msg("error obtaining subscriber from token")

		return nil, err
	}

	return subscriber, nil
}

// getOrgByID returns the organization based on the id in the request
func (h *Handler) getOrgByID(ctx context.Context, id string) (*ent.Organization, error) {
	org, err := transaction.FromContext(ctx).Organization.Get(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("error obtaining organization from id")

		return nil, err
	}

	return org, nil
}

// getOrgSettingByOrgID returns the organization settings from an organization ID and context
func (h *Handler) getOrgSettingByOrgID(ctx context.Context, id string) (*ent.OrganizationSetting, error) {
	settings, err := transaction.FromContext(ctx).OrganizationSetting.Query().Where(
		organizationsetting.OrganizationID(id),
	).Only(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error obtaining organization settings from id")

		return nil, err
	}

	return settings, nil
}

func (h *Handler) fetchOrCreateStripe(context context.Context, orgsetting *ent.OrganizationSetting) (*stripe.Customer, error) {
	if orgsetting.BillingEmail == "" {
		log.Error().Msgf("billing email is required to be set to create a checkout session")
		return nil, ErrNoBillingEmail
	}

	var stripeCustomer *stripe.Customer

	if orgsetting.StripeID != "" {
		cust, err := h.Entitlements.Client.Customers.Get(orgsetting.BillingEmail, nil)
		if err != nil {
			log.Error().Err(err).Msg("error fetching stripe customer")
			return nil, err
		}

		if cust.Email != orgsetting.BillingEmail {
			log.Error().Msgf("customer email does not match, updating stripe customer")

			_, err := h.Entitlements.Client.Customers.Update(orgsetting.StripeID, &stripe.CustomerParams{
				Email: &orgsetting.BillingEmail,
			})
			if err != nil {
				log.Error().Err(err).Msg("error updating stripe customer")
				return nil, err
			}
		}

		return cust, nil
	}

	stripeCustomer, err := h.Entitlements.Client.Customers.New(&stripe.CustomerParams{
		Email: &orgsetting.BillingEmail,
	})
	if err != nil {
		log.Error().Err(err).Msg("error creating stripe customer")
		return nil, err
	}

	if err := h.updateOrganizationSettingWithCustomerID(context, orgsetting.ID, stripeCustomer.ID); err != nil {
		log.Error().Err(err).Msg("error updating organization setting with stripe customer id")
		return nil, err
	}

	return stripeCustomer, nil
}

func (h *Handler) updateOrganizationSettingWithCustomerID(ctx context.Context, orgsettingID, customerID string) error {
	if _, err := transaction.FromContext(ctx).OrganizationSetting.UpdateOneID(orgsettingID).
		SetStripeID(customerID).
		Save(ctx); err != nil {
		log.Error().Err(err).Msg("error updating organization setting with stripe customer id")

		return err
	}

	return nil
}
