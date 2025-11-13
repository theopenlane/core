package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/oklog/ulid/v2"
	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects/storage"
	dbprovider "github.com/theopenlane/core/pkg/objects/storage/providers/database"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
)

type Download struct {
	Token string
	DownloadToken
}

// DownloadToken holds the persisted token details required for verification.
type DownloadToken struct {
	Expires sql.NullString
	Token   sql.NullString
	Secret  []byte
}

// FileDownloadHandler serves files that are stored in the database backend using a presigned token.
func (h *Handler) FileDownloadHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.FileDownload{}, models.ExampleFileDownloadRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	if h.ObjectStore == nil {
		return h.InternalServerError(ctx, ErrObjectStoreUnavailable, openapi)
	}

	requestCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithDownloadToken(requestCtx, in.Token)

	downloadFile, downloadTokenRecord, err := h.getFilebyDownloadToken(ctxWithToken, in.Token)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.BadRequest(ctx, err, openapi)
		}

		logx.FromContext(requestCtx).Error().Err(err).Msg("error retrieving user token")
		return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
	}

	logx.FromContext(requestCtx).Debug().Msg("able to fetch file by download token")

	d := &Download{
		Token: in.Token,
	}

	if err := d.setDownloadTokens(downloadTokenRecord, d.Token); err != nil {
		logx.FromContext(requestCtx).Debug().Err(err).Msg("download token mismatch")
		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	expiresAt, err := d.GetDownloadExpires()
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Msg("unable to parse download token expiration")
		return h.InternalServerError(ctx, ErrUnableToVerifyEmail, openapi)
	}

	if expiresAt.IsZero() {
		logx.FromContext(requestCtx).Debug().Msg("download token missing expiration")
		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	storFile := buildStorageFile(downloadFile)

	objectURI := buildObjectURI(storFile, downloadFile)

	downloadToken := &tokens.DownloadToken{
		ObjectURI: objectURI,
		SigningInfo: tokens.SigningInfo{
			ExpiresAt: expiresAt,
		},
	}

	if downloadTokenRecord != nil {
		if downloadTokenRecord.UserID != nil && *downloadTokenRecord.UserID != "" {
			if userID, parseErr := ulid.Parse(*downloadTokenRecord.UserID); parseErr == nil {
				downloadToken.UserID = userID
			}
		}
		if downloadTokenRecord.OrganizationID != nil && *downloadTokenRecord.OrganizationID != "" {
			if orgID, parseErr := ulid.Parse(*downloadTokenRecord.OrganizationID); parseErr == nil {
				downloadToken.OrgID = orgID
			}
		}
	}

	if err := downloadToken.Verify(d.GetDownloadToken(), d.Secret); err != nil {
		logx.FromContext(requestCtx).Debug().Err(err).Msg("download token verification failed")

		if errors.Is(err, tokens.ErrTokenExpired) {
			return h.BadRequest(ctx, err, openapi)
		}

		return h.Unauthorized(ctx, ErrUnauthorized, openapi)
	}

	if err := validateTokenAuthorization(requestCtx, downloadToken); err != nil {
		logx.FromContext(requestCtx).Debug().Err(err).Msg("token authorization failed")
		return h.Unauthorized(ctx, err, openapi)
	}

	download, err := h.ObjectStore.Download(requestCtx, nil, storFile, &storage.DownloadOptions{})
	if err != nil {
		if errors.Is(err, dbprovider.ErrFileNotFound) || ent.IsNotFound(err) {
			logx.FromContext(requestCtx).Debug().Err(err).Msg("file not found in storage")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		logx.FromContext(requestCtx).Error().Err(err).Msg("error downloading file from storage")
		return h.InternalServerError(ctx, err, openapi)
	}

	headers := ctx.Response().Header()
	headers.Set(echo.HeaderContentType, downloadFile.DetectedContentType)
	headers.Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", downloadFile.ProvidedFileName))

	return ctx.Blob(http.StatusOK, downloadFile.DetectedContentType, download.File)
}

// buildObjectURI builds the object URI from the storagetypes.File or ent.File
func buildObjectURI(file *storagetypes.File, entFile *ent.File) string {
	if entFile != nil && entFile.URI != "" {
		return entFile.URI
	}

	if file != nil {
		return file.FullURI
	}

	return ""
}

// validateTokenAuthorization checks if the token's user ID matches the authenticated user in the request context
func validateTokenAuthorization(requestCtx context.Context, downloadToken *tokens.DownloadToken) error {
	if !ulids.IsZero(downloadToken.UserID) {
		user, ok := auth.AuthenticatedUserFromContext(requestCtx)
		if !ok || user == nil {
			return ErrUnauthorized
		}

		userULID, err := ulid.Parse(user.SubjectID)
		if err != nil || userULID != downloadToken.UserID {
			return ErrUnauthorized
		}
	}

	return nil
}

// buildStorageFile constructs a storagetypes.File from an ent.File entity
func buildStorageFile(fileEntity *ent.File) *storagetypes.File {
	storFile := &storagetypes.File{
		ID:           fileEntity.ID,
		OriginalName: fileEntity.ProvidedFileName,
		ProviderType: storagetypes.ProviderType(fileEntity.StorageProvider),
		FileMetadata: storagetypes.FileMetadata{
			Key:         fileEntity.StoragePath,
			Bucket:      fileEntity.StorageVolume,
			ContentType: fileEntity.DetectedContentType,
			FullURI:     fileEntity.URI,
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storagetypes.ProviderType(fileEntity.StorageProvider),
			},
		},
	}

	return storFile
}

// setDownloadTokens sets the download token details from the database record
func (d *Download) setDownloadTokens(tokenRecord *ent.FileDownloadToken, downloadToken string) error {
	if tokenRecord == nil || tokenRecord.Token != downloadToken {
		return ErrNotFound
	}

	if tokenRecord.Secret == nil || len(*tokenRecord.Secret) == 0 {
		return ErrNotFound
	}

	d.DownloadToken.Token = sql.NullString{String: tokenRecord.Token, Valid: true}
	d.Secret = *tokenRecord.Secret

	if tokenRecord.TTL != nil {
		d.Expires = sql.NullString{String: tokenRecord.TTL.UTC().Format(time.RFC3339Nano), Valid: true}
	} else {
		d.Expires = sql.NullString{}
	}

	return nil
}

// GetDownloadToken returns the download token string
func (d *Download) GetDownloadToken() string {
	if d.DownloadToken.Token.Valid {
		return d.DownloadToken.Token.String
	}

	return ""
}

// GetDownloadExpires returns the expiration time of the download token
func (d *Download) GetDownloadExpires() (time.Time, error) {
	if d.Expires.Valid {
		return time.Parse(time.RFC3339Nano, d.Expires.String)
	}

	return time.Time{}, nil
}
