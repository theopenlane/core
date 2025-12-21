package handlers

import (
	"context"
	"fmt"
	"time"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailverificationtoken"
	"github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/filedownloadtoken"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnerregistrationtoken"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/passwordresettoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/generated/webauthn"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	apimodels "github.com/theopenlane/core/pkg/openapi"
)

// updateUserLastSeen updates the last seen timestamp of the user and login method used
func (h *Handler) updateUserLastSeen(ctx context.Context, id string, authProvider enums.AuthProvider) error {
	if err := transaction.FromContext(ctx).
		User.
		UpdateOneID(id).
		SetLastSeen(time.Now()).
		SetLastLoginProvider(authProvider).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user last seen")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error creating new user")

		return nil, err
	}

	metrics.RecordRegistration()

	return meowuser, nil
}

// updateSubscriberVerifiedEmail updates a subscriber by in the database based on the input and sets to active with verified email
func (h *Handler) updateSubscriberVerifiedEmail(ctx context.Context, id string, input ent.UpdateSubscriberInput) error {
	err := transaction.FromContext(ctx).Subscriber.UpdateOneID(id).
		SetInput(input).
		SetActive(true).
		SetVerifiedEmail(true).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber verified")
		return err
	}

	return nil
}

// updateSubscriber updates a subscriber by in the database based on the input
func (h *Handler) updateSubscriberVerificationToken(ctx context.Context, user *User) error {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse ttl")
		return err
	}

	err = transaction.FromContext(ctx).Subscriber.UpdateOneID(user.ID).
		SetToken(user.EmailVerificationToken.String).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(ttl).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber tokens")

		return err
	}

	return nil
}

// createEmailVerificationToken creates a new email verification for the user
func (h *Handler) createEmailVerificationToken(ctx context.Context, user *User) (*ent.EmailVerificationToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse ttl")
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
		logx.FromContext(ctx).Error().Err(err).Msg("error creating email verification token")

		return nil, err
	}

	return meowtoken, nil
}

// createPasswordResetToken creates a new password reset token for the user
func (h *Handler) createPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse ttl")
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
		logx.FromContext(ctx).Error().Err(err).Msg("error creating password reset token")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from email verification token")

		return nil, err
	}

	return user, nil
}

// getFilebyDownloadToken returns the ent file and download token based on the token in the request
func (h *Handler) getFilebyDownloadToken(ctx context.Context, token string) (*ent.File, *ent.FileDownloadToken, error) {
	tokenRecord, err := transaction.FromContext(ctx).FileDownloadToken.Query().
		Where(filedownloadtoken.Token(token)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining file download token")

		return nil, nil, err
	}

	if tokenRecord.FileID == nil || *tokenRecord.FileID == "" {
		logx.FromContext(ctx).Error().Msg("file download token missing file id")

		return nil, nil, ErrDownloadTokenMissingFile
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	fileRecord, err := transaction.FromContext(ctx).File.Get(allowCtx, *tokenRecord.FileID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining file from download token")

		return nil, nil, err
	}

	return fileRecord, tokenRecord, nil
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
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from reset token")

		return nil, err
	}

	return user, nil
}

// getUserByEmail returns the ent user with the user settings based on the email in the request
func (h *Handler) getUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.EmailEqualFold(email)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from email")

		return nil, err
	}

	return user, nil
}

// getUserByID returns the ent user with the user settings based on the email in the request
func (h *Handler) getUserByID(ctx context.Context, id string) (*ent.User, context.Context, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.ID(id)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from id")

		return nil, ctx, err
	}

	// set user in the viewer context for the rest of the request
	ctx = setAuthenticatedContext(ctx, user)

	// Add webauthn to the response
	webAuthns, err := user.QueryWebauthns().All(ctx)
	if err != nil {
		return user, ctx, err
	}

	user.Edges.Webauthns = webAuthns

	return user, ctx, nil
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
		logx.FromContext(ctx).Error().Err(err).Msg("error checking existing webauthn credentials")

		return err
	}

	if count >= h.OauthProvider.Webauthn.MaxDevices {
		logx.FromContext(ctx).Error().Err(err).Msg("max devices reached")

		return ErrMaxDeviceLimit
	}

	err = transaction.FromContext(ctx).Webauthn.Create().
		SetOwnerID(user.ID).
		SetTransports(transports).
		SetAttestationType(credential.AttestationType).
		SetAaguid(models.ToAAGUID(credential.Authenticator.AAGUID)).
		SetCredentialID(credential.ID).
		SetPublicKey(credential.PublicKey).
		SetBackupState(credential.Flags.BackupEligible).
		SetBackupEligible(credential.Flags.BackupEligible).
		SetUserPresent(credential.Flags.UserPresent).
		SetUserVerified(credential.Flags.UserVerified).
		SetSignCount(int32(credential.Authenticator.SignCount)). //nolint:gosec
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating passkey")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error retrieving user")

		return nil, err
	}

	return user, nil
}

// getUserTFASettings returns the the user with their tfa settings based on the user ID
func (h *Handler) getUserTFASettings(ctx context.Context, userID string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().Where(
		user.ID(userID),
	).WithTfaSettings().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error retrieving tfa settings for user")

		return nil, err
	}

	return user, nil
}

// updateRecoveryCodes updates the recovery codes for the user in their tfa settings
func (h *Handler) updateRecoveryCodes(ctx context.Context, tfaID string, codes []string) error {
	if err := transaction.FromContext(ctx).TFASetting.UpdateOneID(tfaID).
		SetRecoveryCodes(codes).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating recovery codes")

		return err
	}

	return nil
}

// getUserByInviteToken returns the ent user based on the invite token in the request
func (h *Handler) getUserByInviteToken(ctx context.Context, token string) (*ent.Invite, error) {
	recipient, err := transaction.FromContext(ctx).Invite.Query().
		Where(
			invite.Token(token),
		).WithOwner().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from token")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error counting verification reset tokens")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining verification reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error expiring verification token")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining password reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error expiring password reset token")

			return err
		}
	}

	return nil
}

// setEmailConfirmed sets the user setting field email_confirmed to true within a transaction
func (h *Handler) setEmailConfirmed(ctx context.Context, user *ent.User) error {
	if err := transaction.FromContext(ctx).UserSetting.
		UpdateOne(user.Edges.Setting).SetEmailConfirmed(true).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error setting email confirmed")

		return err
	}

	return nil
}

// updateUserPassword changes a updates a user's password in the database
func (h *Handler) updateUserPassword(ctx context.Context, id string, password string) error {
	if err := transaction.FromContext(ctx).User.UpdateOneID(id).SetPassword(password).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user password")

		return err
	}

	return nil
}

// CheckAndCreateUser takes a user with an OauthTooToken set in the context and checks if the user is already created
// if the user already exists, update last seen
func (h *Handler) CheckAndCreateUser(ctx context.Context, name, email string, provider enums.AuthProvider, image string) (*ent.User, error) {
	// check if users exists
	entUser, err := h.getUserByEmail(ctx, email)
	if err != nil {
		// if the user is not found, create now
		if ent.IsNotFound(err) {
			// create the input based on the provider
			input := createUserInput(name, email, provider, image)

			// create user in the database
			entUser, err = h.createUser(ctx, input)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error creating new user")

				return nil, err
			}

			// pull latest user settings to ensure any hook updates are included
			entUser.Edges.Setting, err = entUser.QuerySetting().Only(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error fetching user settings")

				return nil, err
			}

			// return newly created user
			return entUser, nil
		}

		return nil, err
	}

	// update last seen of user
	if err := h.updateUserLastSeen(ctx, entUser.ID, provider); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user last seen")

		return nil, err
	}

	// update the return
	entUser.LastLoginProvider = provider

	// update user avatar
	if err := h.updateUserAvatar(ctx, entUser, image); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user avatar")

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
	input.LastLoginProvider = &provider
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

	if err := transaction.FromContext(ctx).
		User.UpdateOneID(user.ID).
		SetAvatarRemoteURL(image).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user avatar")
		return err
	}

	return nil
}

// setWebauthnAllowed sets the user setting field is_webauthn_allowed to true within a transaction
func (h *Handler) setWebauthnAllowed(ctx context.Context, user *ent.User) error {
	if err := transaction.FromContext(ctx).UserSetting.Update().SetIsWebauthnAllowed(true).
		Where(
			usersetting.UserID(user.ID),
		).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error setting webauthn allowed")

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
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining subscriber from token")

		return nil, err
	}

	return subscriber, nil
}

// getOrgByID returns the organization based on the id in the request
func (h *Handler) getOrgByID(ctx context.Context, id string) (*ent.Organization, error) {
	org, err := transaction.FromContext(ctx).Organization.Get(ctx, id)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining organization from id")

		return nil, err
	}

	return org, nil
}

// createEvent creates a new event in the database but requires mapped input
func (h *Handler) createEvent(ctx context.Context, input ent.CreateEventInput) (*ent.Event, error) {
	event, err := transaction.FromContext(ctx).Event.Create().SetInput(input).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating event")

		return nil, err
	}

	return event, nil
}

// checkForEventID checks if the event ID exists in the database
func (h *Handler) checkForEventID(ctx context.Context, id string) (bool, error) {
	exists, err := transaction.FromContext(ctx).Event.Query().Where(event.EventID(id)).Exist(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking for event ID")
		return false, err
	}

	return exists, nil
}

func (h *Handler) getOrgByJobRunnerVerificationToken(ctx context.Context, token string) (*ent.JobRunnerRegistrationToken, error) {
	registrationToken, err := transaction.FromContext(ctx).
		JobRunnerRegistrationToken.Query().
		Where(
			jobrunnerregistrationtoken.Token(token),
		).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching runner registration token from database")

		return nil, err
	}

	return registrationToken, nil
}

func (h *Handler) createJobRunner(ctx context.Context, token *ent.JobRunnerRegistrationToken, req apimodels.JobRunnerRegistrationRequest) error {
	input := ent.CreateJobRunnerInput{
		Name:    req.Name,
		Tags:    req.Tags,
		OwnerID: &token.OwnerID,
	}

	err := transaction.FromContext(ctx).JobRunner.Create().
		SetInput(input).
		SetCreatedBy(token.ID).
		SetUpdatedBy(token.ID).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("could not create job runner")
		return err
	}

	return nil
}

// getOrganizationSettingByOrgID returns the organization setting for a given organization
func (h *Handler) getOrganizationSettingByOrgID(ctx context.Context, orgID string) (*ent.OrganizationSetting, error) {
	setting, err := transaction.FromContext(ctx).OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching organization settings")

		return nil, err
	}

	return setting, nil
}

// getUserDefaultOrgID returns the default organization ID for a user
func (h *Handler) getUserDefaultOrgID(ctx context.Context, userID string) (string, error) {
	us, err := transaction.FromContext(ctx).UserSetting.Query().Where(usersetting.UserID(userID)).WithDefaultOrg().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching user settings")

		return "", err
	}

	if us.Edges.DefaultOrg == nil {
		return "", fmt.Errorf("%w: default org", ErrMissingField)
	}

	return us.Edges.DefaultOrg.ID, nil
}
