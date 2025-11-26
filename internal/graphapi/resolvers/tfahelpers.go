package resolvers


import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/core/internal/ent/generated"
)

// generateTFAQRCode generates a QR code for the user's TFA secret
func (r *mutationResolver) generateTFAQRCode(ctx context.Context, settings *generated.TFASetting, email, userID string) (string, error) {
	if !graphutils.CheckForRequestedField(ctx, "qrCode") {
		return "", nil
	}

	if !settings.TotpAllowed || settings.TfaSecret == nil {
		return "", nil
	}

	// generate a new QR code if it was requested
	qrCode, err := r.db.TOTP.Manager.TOTPQRString(&totp.User{
		ID:            userID,
		TFASecret:     *settings.TfaSecret,
		Email:         sql.NullString{String: email, Valid: true},
		IsTOTPAllowed: settings.TotpAllowed,
	})
	if err != nil {
		return "", err
	}

	return qrCode, nil
}

// getDecryptedTFASecret returns the TFA secret for the user, decrypted
func (r *mutationResolver) getDecryptedTFASecret(ctx context.Context, settings *generated.TFASetting) (string, error) {
	if !graphutils.CheckForRequestedField(ctx, "tfaSecret") {
		return "", nil
	}

	if !settings.TotpAllowed || settings.TfaSecret == nil {
		return "", nil
	}

	// generate a new QR code if it was requested
	secret, err := r.db.TOTP.Manager.TOTPDecryptedSecret(*settings.TfaSecret)
	if err != nil {
		return "", err
	}

	return secret, nil
}
