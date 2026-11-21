package aikido

import (
	"context"
	"fmt"

	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// assetVariantCodeRepository is the mapping variant for code repository payloads
	assetVariantCodeRepository = "code_repository"
	// assetVariantContainer is the mapping variant for container image payloads
	assetVariantContainer = "container"
	// assetVariantCloud is the mapping variant for cloud environment payloads
	assetVariantCloud = "cloud"
	// assetVariantDomain is the mapping variant for domain payloads
	assetVariantDomain = "domain"
	// assetVariantVirtualMachine is the mapping variant for virtual machine payloads
	assetVariantVirtualMachine = "virtual_machine"
	// assetVariantDevice is the mapping variant for endpoint protection device payloads
	assetVariantDevice = "device"
	// assetVariantSubdomain is the mapping variant for subdomains discovered under a monitored domain
	assetVariantSubdomain = "subdomain"
)

// subdomainDomainIDKey is the subdomain payload key the parent domain id is attached under
const subdomainDomainIDKey = "domain_id"

// assetKind describes one Aikido inventory listing ingested as assets
type assetKind struct {
	// variant is the mapping variant applied to the listing's records
	variant string
	// path is the list endpoint
	path string
	// params are the query parameters requesting fields the listing omits by default
	params map[string]string
	// perPage is the page size requested from the endpoint, zero when the endpoint returns every record unpaged
	perPage int
	// nameKey is the record field carrying the asset name issues refer to
	nameKey string
	// optional kinds need a scope older installations did not grant, so a failure skips the kind
	optional bool
}

// assetKinds are the inventory listings collected by the asset sync
var assetKinds = []assetKind{
	{variant: assetVariantCodeRepository, path: pathCodeRepositories, params: map[string]string{"include_labels": "true"}, perPage: codeRepositoryPageSize, nameKey: "name"},
	{variant: assetVariantContainer, path: pathContainers, params: map[string]string{"include_labels": "true", "include_connectivity": "true", "include_sensitivity": "true"}, perPage: assetPageSize, nameKey: "name"},
	{variant: assetVariantCloud, path: pathClouds, perPage: assetPageSize, nameKey: "name"},
	{variant: assetVariantDomain, path: pathDomains, perPage: assetPageSize, nameKey: "domain"},
	{variant: assetVariantVirtualMachine, path: pathVirtualMachines, perPage: virtualMachinePageSize, nameKey: "name", optional: true},
	{variant: assetVariantDevice, path: pathDevices, nameKey: "name", optional: true},
}

// runAssetSync collects the Aikido inventory for asset ingestion
func runAssetSync(ctx context.Context, _ types.OperationRequest, client *APIClient, _ AssetSync) ([]types.IngestPayloadSet, error) {
	return CollectAssets(ctx, client)
}

// CollectAssets lists the Aikido inventory and the subdomains under each monitored domain, emitting one asset envelope per record
func CollectAssets(ctx context.Context, c *APIClient) ([]types.IngestPayloadSet, error) {
	envelopes := make([]types.MappingEnvelope, 0)

	var domains []map[string]any

	for _, kind := range assetKinds {
		records, err := c.listKind(ctx, kind)
		if err != nil {
			if !kind.optional || ctx.Err() != nil {
				return nil, fmt.Errorf("%w: %s", ErrAssetsFetchFailed, kind.variant)
			}

			logx.FromContext(ctx).Warn().Err(err).Str("variant", kind.variant).Msg("aikido: unable to list optional assets, skipping them")

			continue
		}

		if kind.variant == assetVariantDomain {
			domains = records
		}

		for _, record := range records {
			envelope, err := assetEnvelope(ctx, kind.variant, mapx.Value[string](record, kind.nameKey), record)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	subdomains, err := collectSubdomains(ctx, c, domains)
	if err != nil {
		return nil, err
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaAsset.Name,
			Envelopes: append(envelopes, subdomains...),
		},
	}, nil
}

// collectSubdomains lists the subdomains under each monitored domain, skipping the ones excluded in Aikido
func collectSubdomains(ctx context.Context, c *APIClient, domains []map[string]any) ([]types.MappingEnvelope, error) {
	envelopes := make([]types.MappingEnvelope, 0)

	for _, domain := range domains {
		domainID := recordID(domain)

		records, err := c.listAll(ctx, fmt.Sprintf(pathSubdomainsFormat, domainID), nil, assetPageSize)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrAssetsFetchFailed, assetVariantSubdomain)
		}

		for _, record := range records {
			if mapx.Value[bool](record, "is_excluded") {
				continue
			}

			record[subdomainDomainIDKey] = domainID

			envelope, err := assetEnvelope(ctx, assetVariantSubdomain, mapx.Value[string](record, "name"), record)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return envelopes, nil
}

// assetEnvelope marshals one inventory record into an asset envelope named after the asset issues refer to
func assetEnvelope(ctx context.Context, variant, name string, record map[string]any) (types.MappingEnvelope, error) {
	envelope, err := providerkit.MarshalEnvelopeVariant(variant, name, mapx.PruneMapZeroAny(record), ErrPayloadEncode)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Any("asset_id", record["id"]).Str("variant", variant).Msg("aikido: failed to marshal asset envelope")
	}

	return envelope, err
}

// listKind lists the records of one asset kind, in a single call when the endpoint is unpaged
func (c *APIClient) listKind(ctx context.Context, kind assetKind) ([]map[string]any, error) {
	if kind.perPage == 0 {
		var records []map[string]any

		_, err := c.get(ctx, kind.path, kind.params, &records)

		return records, err
	}

	return c.listAll(ctx, kind.path, kind.params, kind.perPage)
}
