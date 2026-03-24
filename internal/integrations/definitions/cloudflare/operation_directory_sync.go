package cloudflare

import (
	"context"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// cloudflareMemberPayload is the normalized payload for a Cloudflare account member
type cloudflareMemberPayload struct {
	// ID is the Cloudflare membership identifier
	ID string `json:"id,omitempty"`
	// Email is the contact email of the member
	Email string `json:"email,omitempty"`
	// Status is the membership status (accepted, pending)
	Status string `json:"status,omitempty"`
	// UserID is the Cloudflare user identifier
	UserID string `json:"user_id,omitempty"`
	// FirstName is the user's first name
	FirstName string `json:"first_name,omitempty"`
	// LastName is the user's last name
	LastName string `json:"last_name,omitempty"`
	// TwoFactorEnabled indicates whether 2FA is enabled for the user
	TwoFactorEnabled bool `json:"two_factor_enabled,omitempty"`
}

// DirectorySync collects Cloudflare account members for directory account ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(cloudflareClient, func(ctx context.Context, request types.OperationRequest, client *cf.Client) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, client, cfg)
	})
}

// Run collects Cloudflare account members and emits directory account ingest payloads
func (DirectorySync) Run(ctx context.Context, client *cf.Client, cfg UserInput) ([]types.IngestPayloadSet, error) {
	if cfg.AccountID == "" {
		return nil, ErrAccountIDMissing
	}

	iter := client.Accounts.Members.ListAutoPaging(ctx, accounts.MemberListParams{
		AccountID: cf.F(cfg.AccountID),
	})

	envelopes := make([]types.MappingEnvelope, 0)

	for iter.Next() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		member := iter.Current()

		payload := cloudflareMemberPayload{
			ID:               member.ID,
			Email:            member.User.Email,
			Status:           string(member.Status),
			UserID:           member.User.ID,
			FirstName:        member.User.FirstName,
			LastName:         member.User.LastName,
			TwoFactorEnabled: member.User.TwoFactorAuthenticationEnabled,
		}

		resource := cfg.AccountID + "/" + member.User.Email

		envelope, err := providerkit.MarshalEnvelope(resource, payload, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	if err := iter.Err(); err != nil {
		return nil, ErrMembersFetchFailed
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: envelopes,
		},
	}, nil
}
