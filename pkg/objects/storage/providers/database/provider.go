package database

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/metrics"
	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
)

const (
	defaultDatabaseBucket = "default"
)

// Provider persists file bytes directly into the database.
type Provider struct {
	options       *storage.ProviderOptions
	tokenManager  *tokens.TokenManager
	tokenAudience string
	tokenIssuer   string
}

// Upload stores file contents in the database for the ent file identified by provider hints.
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	client, err := p.entClient(ctx)
	if err != nil {
		return nil, err
	}

	fileID := extractFileIdentifier(opts)
	if fileID == "" {
		return nil, ErrMissingFileIdentifier
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if err := client.File.UpdateOneID(fileID).
		SetFileContents(data).
		Exec(allowCtx); err != nil {
		return nil, err
	}

	size := int64(len(data))
	metrics.RecordStorageUpload(string(storagetypes.DatabaseProvider), size)

	bucket := p.bucket()

	metadata := &storagetypes.UploadedFileMetadata{
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileID,
			Size:         size,
			Bucket:       bucket,
			ProviderType: storagetypes.DatabaseProvider,
			FullURI:      fmt.Sprintf("database://%s/%s", bucket, fileID),
		},
	}

	return metadata, nil
}

// Download retrieves file contents from the database.
func (p *Provider) Download(ctx context.Context, fileRef *storagetypes.File, _ *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	client, err := p.entClient(ctx)
	if err != nil {
		return nil, err
	}

	fileID := fileRef.ID
	if fileID == "" {
		return nil, ErrMissingFileIdentifier
	}

	record, err := client.File.Get(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if len(record.FileContents) == 0 {
		return nil, ErrFileNotFound
	}

	size := int64(len(record.FileContents))
	metrics.RecordStorageDownload(string(storagetypes.DatabaseProvider), size)

	metadata := &storagetypes.DownloadedFileMetadata{
		File:           record.FileContents,
		Size:           size,
		TimeDownloaded: time.Now(),
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileID,
			Bucket:       p.bucket(),
			ContentType:  record.DetectedContentType,
			Name:         record.ProvidedFileName,
			ProviderType: storagetypes.DatabaseProvider,
		},
	}

	return metadata, nil
}

// Delete removes file bytes from the database while retaining metadata.
func (p *Provider) Delete(ctx context.Context, fileRef *storagetypes.File, _ *storagetypes.DeleteFileOptions) error {
	client, err := p.entClient(ctx)
	if err != nil {
		return err
	}

	fileID := fileRef.ID
	if fileID == "" {
		return ErrMissingFileIdentifier
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if err := client.File.UpdateOneID(fileID).
		ClearFileContents().
		Exec(allowCtx); err != nil {
		return err
	}

	metrics.RecordStorageDelete(string(storagetypes.DatabaseProvider))

	return nil
}

// Exists checks whether file bytes are present in the database.
func (p *Provider) Exists(ctx context.Context, fileRef *storagetypes.File) (bool, error) {
	client, err := p.entClient(ctx)
	if err != nil {
		return false, err
	}

	fileID := fileRef.ID
	if fileID == "" {
		return false, ErrMissingFileIdentifier
	}

	return client.File.Query().
		Where(
			file.IDEQ(fileID),
			file.FileContentsNotNil(),
		).
		Exist(ctx)
}

// GetPresignedURL returns a signed URL that proxies through the application for download.
func (p *Provider) GetPresignedURL(ctx context.Context, fileRef *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	if p.tokenManager == nil {
		return "", ErrTokenManagerRequired
	}

	fileID := fileRef.ID
	if fileID == "" {
		return "", ErrMissingFileIdentifier
	}

	duration := opts.Duration
	if duration <= 0 {
		duration = 15 * time.Minute // nolint:mnd
	}

	now := time.Now()
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			ID:        ulids.New().String(),
			Subject:   fileID,
		},
		Scopes: tokens.PermissionScopes{
			Read: []string{fileID},
		},
	}

	if orgID, err := auth.GetOrganizationIDFromContext(ctx); err == nil && orgID != "" {
		claims.OrgID = orgID
	}

	if user, err := auth.GetAuthenticatedUserFromContext(ctx); err == nil && user != nil {
		claims.UserID = user.SubjectID
	}

	if p.tokenAudience != "" {
		claims.Audience = jwt.ClaimStrings{p.tokenAudience}
	}
	if p.tokenIssuer != "" {
		claims.Issuer = p.tokenIssuer
	}

	token := p.tokenManager.CreateToken(claims)
	signed, err := p.tokenManager.Sign(token)
	if err != nil {
		return "", err
	}

	encodedToken := url.QueryEscape(signed)

	base := strings.TrimRight(p.options.Endpoint, "/")
	path := fmt.Sprintf("/v1/files/%s/download", url.PathEscape(fileID))

	if base == "" {
		return fmt.Sprintf("%s?token=%s", path, encodedToken), nil
	}

	return fmt.Sprintf("%s%s?token=%s", base, path, encodedToken), nil
}

// ListBuckets returns the configured logical bucket for the provider.
func (p *Provider) ListBuckets() ([]string, error) {
	return []string{p.bucket()}, nil
}

// GetScheme identifies the provider URI scheme.
func (p *Provider) GetScheme() *string {
	scheme := "database://"
	return &scheme
}

// Close satisfies the Provider interface. No resources to release.
func (p *Provider) Close() error {
	return nil
}

// ProviderType returns the database provider identifier.
func (p *Provider) ProviderType() storagetypes.ProviderType {
	return storagetypes.DatabaseProvider
}

func (p *Provider) entClient(ctx context.Context) (*ent.Client, error) {
	if client := ent.FromContext(ctx); client != nil {
		return client, nil
	}

	if extra, ok := p.options.Extra("ent_client"); ok {
		if client, ok := extra.(*ent.Client); ok && client != nil {
			return client, nil
		}
	}

	return nil, ErrMissingEntClient
}

// bucket returns the configured bucket or a default value
func (p *Provider) bucket() string {
	if p.options == nil || p.options.Bucket == "" {
		return defaultDatabaseBucket
	}

	return p.options.Bucket
}

// extractFileIdentifier retrieves the file ID from provider hints in upload options
func extractFileIdentifier(opts *storagetypes.UploadFileOptions) string {
	if opts == nil {
		return ""
	}

	if opts.ProviderHints != nil && opts.ProviderHints.Metadata != nil {
		if id, ok := opts.ProviderHints.Metadata["file_id"]; ok && id != "" {
			return id
		}
	}

	if opts.FileMetadata.ProviderHints != nil && opts.FileMetadata.ProviderHints.Metadata != nil { // nolint:staticcheck
		if id, ok := opts.FileMetadata.ProviderHints.Metadata["file_id"]; ok && id != "" { // nolint:staticcheck
			return id
		}
	}

	return ""
}
