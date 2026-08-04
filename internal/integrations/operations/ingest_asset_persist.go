package operations

import (
	"context"
	"net/mail"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/logx"
)

// persistAssetInput upserts one Asset record through the catalog-driven entityops upsert
func persistAssetInput(ctx context.Context, db *ent.Client, integration *ent.Integration, createInput ent.CreateAssetInput) (string, error) {
	if createInput.SourceType == nil {
		createInput.SourceType = &enums.SourceTypeImported
	}

	if createInput.IntegrationID == nil {
		createInput.IntegrationID = &integration.ID
	}

	// if user is in system, replace internal owner with
	// internal owner user id
	if createInput.InternalOwner != nil {
		userID, err := resolveInternalOwner(ctx, db, *createInput.InternalOwner)
		if err == nil && userID != nil {
			createInput.InternalOwnerUserID = userID
			createInput.InternalOwner = nil
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error converting internal owner to user, keeping original internal owner")
		}
	}

	return persistCatalogUpsert(ctx, db, entityops.SchemaAsset, integration.OwnerID, createInput)
}

// resolveInternalOwner resolves internal owner
func resolveInternalOwner(ctx context.Context, client *ent.Client, rawValue string) (*string, error) {
	if _, err := mail.ParseAddress(rawValue); err == nil {
		userID, err := client.User.Query().
			Where(
				user.EmailEqualFold(rawValue),
			).
			OnlyID(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}

		if userID != "" {
			return &userID, nil
		}
	}

	return nil, nil
}
