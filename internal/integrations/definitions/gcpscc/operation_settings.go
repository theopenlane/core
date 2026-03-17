package gcpscc

import (
	"context"
	"encoding/json"
	"errors"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/api/iterator"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// settingsPageSize is the number of notification configs requested per paginated API call
	settingsPageSize = 10
	// sampleConfigsCapacity is the pre-allocated capacity for the notification config sample slice
	sampleConfigsCapacity = 5
)

// NotificationConfigSample holds a single SCC notification config entry
type NotificationConfigSample struct {
	// Name is the notification config resource name
	Name string `json:"name"`
	// Description is the notification config description
	Description string `json:"description"`
	// PubSubTopic is the Pub/Sub topic for the notification config
	PubSubTopic string `json:"pubsubTopic"`
	// Parent is the parent resource for the notification config
	Parent string `json:"parent"`
}

// SettingsScan scans GCP SCC notification settings
type SettingsScan struct {
	// Parents is the list of SCC parent resources that were scanned
	Parents []string `json:"parents"`
	// NotificationConfigCount is the total count of notification configs found
	NotificationConfigCount int `json:"notificationConfigCount"`
	// SampleNotificationConfigs holds a representative subset of notification configs
	SampleNotificationConfigs []NotificationConfigSample `json:"sampleNotificationConfigs"`
}

// Handle adapts settings scan to the generic operation registration boundary
func (s SettingsScan) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return s.Run(ctx, request.Credential, c)
	}
}

// Run scans GCP SCC notification configs
func (SettingsScan) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client) (json.RawMessage, error) {
	meta, err := metadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveParents(meta)
	if err != nil {
		return nil, err
	}

	configs := make([]NotificationConfigSample, 0, sampleConfigsCapacity)
	count := 0

	for _, parent := range parents {
		req := &securitycenterpb.ListNotificationConfigsRequest{
			Parent:   parent,
			PageSize: settingsPageSize,
		}

		it := c.ListNotificationConfigs(ctx, req)

		for {
			cfg, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				return nil, ErrNotificationConfigScanFailed
			}

			count++

			if len(configs) < cap(configs) {
				configs = append(configs, NotificationConfigSample{
					Name:        cfg.GetName(),
					Description: cfg.GetDescription(),
					PubSubTopic: cfg.GetPubsubTopic(),
					Parent:      parent,
				})
			}
		}
	}

	return providerkit.EncodeResult(SettingsScan{
		Parents:                   parents,
		NotificationConfigCount:   count,
		SampleNotificationConfigs: configs,
	}, ErrResultEncode)
}
