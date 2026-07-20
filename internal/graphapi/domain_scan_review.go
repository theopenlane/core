package graphapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrDomainScanReviewInvalidRef is returned when a platform/system entityRef or assetRef doesn't
	// match any ref among the accepted vendors/assets in the same review
	ErrDomainScanReviewInvalidRef = errors.New("domain scan review: unknown ref")

	// ErrDomainScanReviewNoScans is returned when none of the submitted scan IDs resolve to a Scan
	// the caller can see
	ErrDomainScanReviewNoScans = errors.New("domain scan review: no matching scans found")

	// ErrDomainScanReviewMixedOrganizations is returned when the submitted scan IDs belong to more
	// than one organization
	ErrDomainScanReviewMixedOrganizations = errors.New("domain scan review: scans belong to more than one organization")

	// ErrDomainScanReviewUnknownScan is returned when one or more submitted scan IDs don't resolve to
	// a Scan the caller can see
	ErrDomainScanReviewUnknownScan = errors.New("domain scan review: one or more scan IDs could not be resolved")

	// ErrDomainScanReviewImportFailed is returned when the accepted review could not be queued for
	// import; the underlying cause is internal and gets logged, not surfaced
	ErrDomainScanReviewImportFailed = errors.New("domain scan review: failed to queue import")
)

// importDomainScanReview validates the accepted review, resolves the organization it belongs to
// from the referenced scans, and emits it for asynchronous processing
func (r *mutationResolver) importDomainScanReview(ctx context.Context, input model.ImportDomainScanReviewInput) (*model.ImportDomainScanReviewPayload, error) {
	if err := validateDomainScanReviewRefs(input); err != nil {
		return nil, err
	}

	client := withTransactionalMutation(ctx)

	scans, err := client.Scan.Query().Where(scan.IDIn(input.ScanIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(scans) == 0 {
		return nil, ErrDomainScanReviewNoScans
	}

	knownScanIDs := make(map[string]bool, len(input.ScanIDs))
	for _, id := range input.ScanIDs {
		knownScanIDs[id] = true
	}

	if len(scans) != len(knownScanIDs) {
		return nil, ErrDomainScanReviewUnknownScan
	}

	organizationID := scans[0].OwnerID

	for _, s := range scans[1:] {
		if s.OwnerID != organizationID {
			return nil, ErrDomainScanReviewMixedOrganizations
		}
	}

	ctx, err = common.SetOrganizationInAuthContext(ctx, &organizationID)
	if err != nil {
		return nil, err
	}

	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return nil, ErrCampaignDispatchRuntimeRequired
	}

	receipt := rt.Gala().EmitWithHeaders(ctx, operations.DomainScanImportTopic, buildDomainScanImportEnvelope(organizationID, input), gala.Headers{})
	if receipt.Err != nil {
		logx.FromContext(ctx).Error().Err(receipt.Err).Msg("domain scan review: failed to queue import")
		return nil, ErrDomainScanReviewImportFailed
	}

	return &model.ImportDomainScanReviewPayload{Accepted: true}, nil
}

// validateDomainScanReviewRefs checks that every entityRef/assetRef referenced by a platform or
// a system resolves to a ref among the review's accepted vendors/assets, and that every
// platformRef referenced by a system resolves to a ref among the review's accepted platforms
func validateDomainScanReviewRefs(input model.ImportDomainScanReviewInput) error {
	knownEntityRefs := make(map[string]bool, len(input.Vendors))
	for _, vendor := range input.Vendors {
		knownEntityRefs[vendor.Ref] = true
	}

	knownAssetRefs := make(map[string]bool, len(input.Assets))
	for _, asset := range input.Assets {
		knownAssetRefs[asset.Ref] = true
	}

	knownPlatformRefs := make(map[string]bool, len(input.Platforms))
	for _, platform := range input.Platforms {
		knownPlatformRefs[platform.Ref] = true
	}

	checkRefs := func(entityRefs, assetRefs []string) error {
		for _, ref := range entityRefs {
			if !knownEntityRefs[ref] {
				return fmt.Errorf("%w: vendor ref %q", ErrDomainScanReviewInvalidRef, ref)
			}
		}

		for _, ref := range assetRefs {
			if !knownAssetRefs[ref] {
				return fmt.Errorf("%w: asset ref %q", ErrDomainScanReviewInvalidRef, ref)
			}
		}

		return nil
	}

	for _, platform := range input.Platforms {
		if err := checkRefs(platform.EntityRefs, platform.AssetRefs); err != nil {
			return err
		}
	}

	for _, system := range input.Systems {
		if err := checkRefs(system.EntityRefs, system.AssetRefs); err != nil {
			return err
		}

		for _, ref := range system.PlatformRefs {
			if !knownPlatformRefs[ref] {
				return fmt.Errorf("%w: platform ref %q", ErrDomainScanReviewInvalidRef, ref)
			}
		}
	}

	return nil
}

// buildDomainScanImportEnvelope translates the GraphQL input into the envelope the async import
// handler consumes
func buildDomainScanImportEnvelope(organizationID string, input model.ImportDomainScanReviewInput) operations.ImportDomainScanReviewEnvelope {
	envelope := operations.ImportDomainScanReviewEnvelope{
		OrganizationID: organizationID,
		ScanIDs:        input.ScanIDs,
		Vendors:        make([]operations.ImportDomainScanReviewVendor, 0, len(input.Vendors)),
		Assets:         make([]operations.ImportDomainScanReviewAsset, 0, len(input.Assets)),
	}

	for _, vendor := range input.Vendors {
		v := operations.ImportDomainScanReviewVendor{
			Ref:        vendor.Ref,
			Name:       vendor.Name,
			Categories: vendor.Categories,
		}

		if vendor.Domain != nil {
			v.Domain = *vendor.Domain
		}

		envelope.Vendors = append(envelope.Vendors, v)
	}

	for _, asset := range input.Assets {
		a := operations.ImportDomainScanReviewAsset{
			Ref:        asset.Ref,
			Name:       asset.Name,
			Categories: asset.Categories,
		}

		if asset.Identifier != nil {
			a.Identifier = *asset.Identifier
		}

		if asset.Website != nil {
			a.Website = *asset.Website
		}

		envelope.Assets = append(envelope.Assets, a)
	}

	for _, platform := range input.Platforms {
		p := operations.ImportDomainScanReviewPlatform{
			Ref:        platform.Ref,
			Name:       platform.Name,
			EntityRefs: platform.EntityRefs,
			AssetRefs:  platform.AssetRefs,
		}

		if platform.Description != nil {
			p.Description = *platform.Description
		}

		envelope.Platforms = append(envelope.Platforms, p)
	}

	for _, system := range input.Systems {
		s := operations.ImportDomainScanReviewSystem{
			Name:         system.Name,
			EntityRefs:   system.EntityRefs,
			AssetRefs:    system.AssetRefs,
			PlatformRefs: system.PlatformRefs,
		}

		if system.Description != nil {
			s.Description = *system.Description
		}

		envelope.Systems = append(envelope.Systems, s)
	}

	for _, finding := range input.Findings {
		f := operations.ImportDomainScanReviewFinding{}

		if finding.Category != nil {
			f.Category = *finding.Category
		}

		if finding.Description != nil {
			f.Description = *finding.Description
		}

		if finding.Severity != nil {
			f.Severity = *finding.Severity
		}

		envelope.Findings = append(envelope.Findings, f)
	}

	return envelope
}
