package objectstore

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/iam/auth"
)

// demoPrefix is the bucket directory holding the demo datasets
const demoPrefix = "demo"

// demoDatasetPattern keeps a dataset name to a single directory under the demo prefix
var demoDatasetPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// Prefix returns the dataset's bucket directory, failing when the name is not a single lowercase directory name
func (d DemoDataset) Prefix() (string, error) {
	if !demoDatasetPattern.MatchString(string(d)) {
		return "", fmt.Errorf("%w: %q", ErrDemoDatasetInvalid, d)
	}

	return path.Join(demoPrefix, string(d)), nil
}

// demoProfileName returns the dataset manifest profile the demo profile imports under, the demo profile when unset
func demoProfileName(profile enums.DemoDataProfile) (string, error) {
	switch profile {
	case "", enums.DemoDataProfileDemo:
		return enums.DemoDataProfileDemo.String(), nil
	case enums.DemoDataProfileDraft:
		return enums.DemoDataProfileDraft.String(), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrDemoProfileUnknown, profile)
	}
}

// DemoSeedUniqueKey returns the dispatch key that drops a second seed of a dataset into an organization under one
// profile while the first is queued or running
func DemoSeedUniqueKey(ownerID string, dataset DemoDataset, profile string) string {
	return "demo-seed:" + ownerID + ":" + dataset.String() + ":" + profile
}

// SeedDemoRequest returns the runtime dispatch that seeds the demo dataset, under the profile, into the caller's organization
func SeedDemoRequest(ctx context.Context, dataset DemoDataset, profile enums.DemoDataProfile) (types.DispatchRequest, error) {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return types.DispatchRequest{}, err
	}

	profileName, err := demoProfileName(profile)
	if err != nil {
		return types.DispatchRequest{}, err
	}

	prefix, err := dataset.Prefix()
	if err != nil {
		return types.DispatchRequest{}, err
	}

	cfg := ImportRecords{Prefix: prefix, Manifest: true, Profile: profileName}

	return demoRequest(orgID, ImportRecordsOp.Name(), cfg, DemoSeedUniqueKey(orgID, dataset, profileName))
}

// DeleteDemoRequest returns the runtime dispatch that removes the demo dataset's seeded rows from the caller's
// organization, leaving the schemas listed in keep untouched
func DeleteDemoRequest(ctx context.Context, dataset DemoDataset, keep []string) (types.DispatchRequest, error) {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return types.DispatchRequest{}, err
	}

	prefix, err := dataset.Prefix()
	if err != nil {
		return types.DispatchRequest{}, err
	}

	// no unique key, a repeated removal only finds what is left
	return demoRequest(orgID, operations.DeleteSourceOp.Name(), operations.DeleteSource{SourceKey: prefix, Keep: keep}, "")
}

// demoRequest builds a manual runtime dispatch of one operation of the definition for the organization
func demoRequest(orgID, operation string, cfg any, uniqueKey string) (types.DispatchRequest, error) {
	config, err := json.Marshal(cfg)
	if err != nil {
		return types.DispatchRequest{}, fmt.Errorf("%w: %w", ErrContentEncode, err)
	}

	return types.DispatchRequest{
		DefinitionID: DefinitionID.ID(),
		OwnerID:      orgID,
		Operation:    operation,
		Config:       config,
		RunType:      enums.IntegrationRunTypeManual,
		Runtime:      true,
		UniqueKey:    uniqueKey,
	}, nil
}
