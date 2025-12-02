package proxy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/oklog/ulid/v2"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// DefaultPresignDuration is the default duration for presigned URLs
	DefaultPresignDurationMinutes = 15
	// SecretDivisor is used to split the secret into nonce and key components
	SecretDivisor = 2
)

var (
	// ErrTokenManagerRequired indicates proxy signing is impossible without a token manager.
	ErrTokenManagerRequired = errors.New("proxy presign requires token manager")
	// ErrEntClientRequired indicates storing secrets requires an ent client.
	ErrEntClientRequired = errors.New("storing secrets requires ent client")
)

// GenerateDownloadURL builds a proxy download URL and persists the signing secret for validation.
func GenerateDownloadURL(ctx context.Context, file *storagetypes.File, duration time.Duration, cfg *storage.ProxyPresignConfig) (string, error) {
	if file == nil || file.ID == "" {
		return "", ErrMissingFileID
	}

	if cfg == nil || cfg.TokenManager == nil {
		return "", ErrTokenManagerRequired
	}

	ctx = WithPresignInterceptorBypass(ctx)

	client := ent.FromContext(ctx)
	if client == nil {
		return "", ErrEntClientRequired
	}

	if duration <= 0 {
		duration = DefaultPresignDurationMinutes * time.Minute
	}

	objectURI := file.FullURI
	if objectURI == "" {
		return "", ErrMissingObjectURI
	}

	options := []tokens.DownloadTokenOption{
		tokens.WithDownloadTokenExpiresIn(duration),
	}

	authUser, ok := auth.AuthenticatedUserFromContext(ctx)
	if !ok {
		return "", ErrAuthenticatedUserRequired
	}

	if userID, err := ulid.Parse(authUser.SubjectID); err == nil {
		options = append(options, tokens.WithDownloadTokenUserID(userID))
	}

	if orgID, err := ulid.Parse(authUser.OrganizationID); err == nil {
		options = append(options, tokens.WithDownloadTokenOrgID(orgID))
	}

	downloadToken, err := tokens.NewDownloadToken(objectURI, options...)
	if err != nil {
		return "", err
	}

	tokenString, secret, err := downloadToken.Sign()
	if err != nil {
		return "", err
	}

	create := client.FileDownloadToken.Create().
		SetToken(tokenString).
		SetFileID(file.ID).
		SetSecret(secret)

	create.SetOwnerID(authUser.SubjectID)

	if !downloadToken.ExpiresAt.IsZero() {
		create.SetTTL(downloadToken.ExpiresAt.UTC().Truncate(time.Microsecond))
	}

	if !ulids.IsZero(downloadToken.UserID) {
		create.SetUserID(downloadToken.UserID.String())
	}

	if !ulids.IsZero(downloadToken.OrgID) {
		create.SetOrganizationID(downloadToken.OrgID.String())
	}

	if err := create.Exec(ctx); err != nil {
		return "", fmt.Errorf("failed to store download token: %w", err)
	}

	return composeDownloadURL(cfg.BaseURL, file.ID, url.QueryEscape(tokenString)), nil
}

// GenerateDownloadURLWithSecret builds a proxy download URL using the provided secret for testing.
func GenerateDownloadURLWithSecret(file *storagetypes.File, secret []byte, duration time.Duration, cfg *storage.ProxyPresignConfig) (string, error) {
	if file == nil || file.ID == "" {
		return "", ErrMissingFileID
	}

	if cfg == nil || cfg.TokenManager == nil {
		return "", ErrTokenManagerRequired
	}

	if len(secret) != 0 && len(secret) != 128 {
		return "", ErrInvalidSecretLength
	}

	if duration <= 0 {
		duration = DefaultPresignDurationMinutes * time.Minute
	}

	objectURI := file.FullURI
	if objectURI == "" {
		return "", ErrMissingObjectURI
	}

	options := []tokens.DownloadTokenOption{
		tokens.WithDownloadTokenExpiresIn(duration),
	}

	downloadToken, err := tokens.NewDownloadToken(objectURI, options...)
	if err != nil {
		return "", err
	}

	var tokenString string

	switch len(secret) {
	case 0:
		var err error

		tokenString, _, err = downloadToken.Sign()
		if err != nil {
			return "", err
		}
	default:
		nonceLen := len(secret) / SecretDivisor
		downloadToken.SetNonce(secret[:nonceLen])

		payload, err := msgpack.Marshal(downloadToken)
		if err != nil {
			return "", err
		}

		mac := hmac.New(sha256.New, secret[nonceLen:])
		if _, err = mac.Write(payload); err != nil {
			return "", err
		}

		tokenString = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	}

	return composeDownloadURL(cfg.BaseURL, file.ID, url.QueryEscape(tokenString)), nil
}

func composeDownloadURL(baseURL, fileID, escapedToken string) string {
	return fmt.Sprintf("%s/%s/download?token=%s", baseURL, url.PathEscape(fileID), escapedToken)
}
