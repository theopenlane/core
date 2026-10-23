package googledrive

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// PickerSession holds what the Google Picker needs to open for an installation
type PickerSession struct {
	// AccessToken is the short-lived OAuth access token scoped to drive.file (populated in the response)
	AccessToken string `json:"accessToken,omitempty"`
	// Expiry is when the access token expires (populated in the response)
	Expiry *time.Time `json:"expiry,omitempty"`
	// APIKey is the browser API key the Picker loads with (populated in the response)
	APIKey string `json:"apiKey,omitempty"`
	// AppID is the Google Cloud project number the Picker grants file access to, derived from the OAuth client ID (populated in the response)
	AppID string `json:"appId,omitempty"`
}

// pickerSessionOperation returns the inline operation ref that issues Picker sessions under the operator config
func pickerSessionOperation(cfg Config) types.OperationRef[PickerSession] {
	return types.OperationPayloadOf[PickerSession]().
		Handles(pickerSession(cfg)).
		Policy(types.ExecutionPolicy{Inline: true})
}

// pickerSession returns the handler that refreshes the installation token and returns it with the Picker settings
func pickerSession(cfg Config) func(context.Context, types.OperationRequest, DriveClient, PickerSession) (json.RawMessage, error) {
	return func(ctx context.Context, _ types.OperationRequest, c DriveClient, _ PickerSession) (json.RawMessage, error) {
		// legacy tokens carry drive.readonly, which must never be handed to the browser
		if c.Legacy {
			return nil, ErrPickerLegacyInstallation
		}

		appID, ok := projectNumber(cfg.ClientID)
		if !ok {
			logx.FromContext(ctx).Error().Msg("googledrive: client id does not start with a google cloud project number")

			return nil, ErrPickerAppIDUnresolved
		}

		token, err := c.TokenSource.Token()
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("googledrive: failed to refresh token for picker session")

			return nil, ErrPickerTokenFailed
		}

		session := PickerSession{
			AccessToken: token.AccessToken,
			APIKey:      cfg.PickerAPIKey,
			AppID:       appID,
		}

		if !token.Expiry.IsZero() {
			session.Expiry = &token.Expiry
		}

		return providerkit.EncodeResult(session, ErrPayloadEncode)
	}
}

// projectNumber returns the Google Cloud project number an OAuth client ID is prefixed with, as in <number>-<hash>.apps.googleusercontent.com
func projectNumber(clientID string) (string, bool) {
	number, _, found := strings.Cut(clientID, "-")
	if !found || number == "" || strings.Trim(number, "0123456789") != "" {
		return "", false
	}

	return number, true
}
