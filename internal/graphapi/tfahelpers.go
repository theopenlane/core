package graphapi

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/utils"
)

// generateTFAQRCode generates a QR code for the user's TFA secret
func (r *mutationResolver) generateTFAQRCode(ctx context.Context, settings *generated.TFASetting, email, userID string) (string, error) {
	if !utils.CheckForRequestedField(ctx, "qrCode") {
		return "", nil
	}

	if !settings.TotpAllowed || settings.TfaSecret == nil {
		return "", nil
	}

	// generate a new QR code if it was requested
	qrCode, err := r.db.TOTP.TOTPManager.TOTPQRString(&totp.User{
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
	if !utils.CheckForRequestedField(ctx, "tfaSecret") {
		return "", nil
	}

	if !settings.TotpAllowed || settings.TfaSecret == nil {
		return "", nil
	}

	// generate a new QR code if it was requested
	secret, err := r.db.TOTP.TOTPManager.TOTPDecryptedSecret(*settings.TfaSecret)
	if err != nil {
		return "", err
	}

	return secret, nil
}
