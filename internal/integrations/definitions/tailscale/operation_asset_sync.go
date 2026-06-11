package tailscale

import (
	"context"
	"fmt"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// repositoryAssetVariant is the mapping variant for device asset payloads
const deviceAssetVariant = "DEVICE"

// IngestHandle adapts Tailscale asset sync to the ingest operation registration boundary
func (a AssetSync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(tailscaleClient, func(ctx context.Context, request types.OperationRequest, client *tsclient.Client) ([]types.IngestPayloadSet, error) {
		return a.Run(ctx, client)
	})
}

// Run collects Tailscale devices and emits asset ingest payloads
func (AssetSync) Run(ctx context.Context, client *tsclient.Client) ([]types.IngestPayloadSet, error) {
	devices, err := listTailscaleDevices(ctx, client)
	if err != nil {
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0, len(devices))

	for _, device := range devices {
		envelope, err := providerkit.MarshalEnvelopeVariant(deviceAssetVariant, device.ID, device, ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("device", device.Name).Msg("tailscale: failed to marshal device")
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	logx.FromContext(ctx).Debug().Int("device_count", len(envelopes)).Msg("tailscale: collected devices")

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: envelopes,
		},
	}, nil
}

// listTailscaleDevices fetches all devices from the Tailscale API and maps them to payloads
func listTailscaleDevices(ctx context.Context, client *tsclient.Client) ([]tsclient.Device, error) {
	devices, err := client.Devices().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDevicesFetchFailed, err)
	}

	return devices, nil
}
