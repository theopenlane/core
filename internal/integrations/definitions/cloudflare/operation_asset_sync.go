package cloudflare

import (
	"context"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/packages/pagination"
	"github.com/cloudflare/cloudflare-go/v7/registrar"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// AssetCollect collects Cloudflare domain registrations and stores them as assets
type AssetCollect struct{}

// IngestHandle adapts asset collection to the ingest operation registration boundary
func (a AssetCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(cloudflareClient, func(ctx context.Context, request types.OperationRequest, client *cf.Client) ([]types.IngestPayloadSet, error) {
		return a.Run(ctx, request.Credentials, client)
	})
}

// Run collects Cloudflare domain registrations and emits asset ingest payloads
func (AssetCollect) Run(ctx context.Context, credentials types.CredentialBindings, client *cf.Client) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		return nil, err
	}

	if meta.AccountID == "" {
		return nil, ErrAccountIDMissing
	}

	registrations, err := fetchRegistrarRegistrations(ctx, client, meta.AccountID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: error fetching Registrar registrations")

		return nil, ErrAssetsFetchFailed
	}

	envelopes := make([]types.MappingEnvelope, 0, len(registrations))
	for _, registration := range registrations {
		envelope, err := providerkit.MarshalEnvelope(meta.AccountID, registration, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: envelopes,
		},
	}, nil
}

type cloudflareRegistrationsResponse struct {
	Result     []registrar.Registration              `json:"result"`
	ResultInfo pagination.CursorPaginationResultInfo `json:"result_info"`
}

func fetchRegistrarRegistrations(ctx context.Context, client *cf.Client, accountID string) ([]registrar.Registration, error) {
	domains := make([]registrar.Registration, 0)

	for cursor := ""; ; {
		var response cloudflareRegistrationsResponse
		params := registrar.RegistrationListParams{
			AccountID: cf.F(accountID),
			PerPage:   cf.F(int64(assetSyncRegistrarPageSize)),
		}

		if cursor != "" {
			params.Cursor = cf.F(cursor)
		}

		if _, err := client.Registrar.Registrations.List(ctx, params, option.WithResponseBodyInto(&response)); err != nil {
			return nil, err
		}

		domains = append(domains, response.Result...)

		cursor = response.ResultInfo.Cursor
		if cursor == "" {
			break
		}
	}

	return domains, nil
}
