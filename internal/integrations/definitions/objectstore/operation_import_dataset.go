package objectstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/placeholder"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/jsonx"
)

const (
	// datasetManifestName is the object describing a dataset's standards and ingest order
	datasetManifestName = "manifest.json"
	// datasetNotificationSubject names the import in the outcome notification when the manifest has no name
	datasetNotificationSubject = "Dataset"
	// datasetRetryDetail tells the user how to retry an import that did not fully complete
	datasetRetryDetail = "remove the imported data and import it again"
	// datasetRemovalDetail tells the user where a completed import can be removed
	datasetRemovalDetail = "you can remove this data from your organization settings"
	// datasetRemovalURL is the console path holding the dataset removal action
	datasetRemovalURL = "/organization-settings/general-settings#demo-data"
	// datasetDefaultDomain is the domain variable's value for an organization without a domain
	datasetDefaultDomain = "example.com"
)

// datasetPattern keeps a dataset path to lowercase directory segments inside the bucket
var datasetPattern = regexp.MustCompile(`^[a-z0-9_]+(/[a-z0-9_]+)*$`)

// datasetManifest describes the standards a dataset clones and the order its schema files ingest in
type datasetManifest struct {
	// Name names the dataset in the outcome notifications
	Name string `json:"name"`
	// Standards are cloned into the organization before the records that link to their controls
	Standards []datasetStandard `json:"standards"`
	// Order lists the schema files to ingest, targets before the records linking to them
	Order []string `json:"order"`
	// Profiles name sets of variable values that fill the records' {{NAME|default}} placeholders over the built-in
	// variables, a record keeps each default when neither sets the variable
	Profiles map[string]placeholder.Variables `json:"profiles"`
}

// profileVariables returns the variables the named profile sets, none for the default profile, failing for a profile
// the manifest does not declare
func (m datasetManifest) profileVariables(profile string) (placeholder.Variables, error) {
	if profile == "" {
		return placeholder.Variables{}, nil
	}

	variables, ok := m.Profiles[profile]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrDatasetProfileUnknown, profile)
	}

	return variables, nil
}

// datasetStandard is a catalog standard the dataset clones and the program its controls join
type datasetStandard struct {
	// ShortName is the catalog standard short name
	ShortName string `json:"shortName"`
	// Categories limits the cloned controls to these categories, all when empty
	Categories []string `json:"categories"`
	// ProgramExternalID is the external id of the imported program the controls join
	ProgramExternalID string `json:"programExternalId"`
}

// runImportDataset imports a manifest-described dataset from the bucket as the dispatching organization and notifies the
// organization of the outcome
func runImportDataset(ctx context.Context, request types.OperationRequest, client *Client, cfg ImportRecords) (json.RawMessage, error) {
	name, result, importErr := importDataset(ctx, request, client, cfg)

	logx.FromContext(ctx).Debug().Str("prefix", cfg.Prefix).Int("persisted", result.Persisted).Int("failed", result.Failed).Msg("objectstore: dataset imported")

	detail := datasetRemovalDetail
	if importErr != nil || result.Failed > 0 {
		detail = datasetRetryDetail
	}

	notifyErr := operations.NotifyImportOutcome(ctx, lo.CoalesceOrEmpty(name, datasetNotificationSubject), result, importErr, detail, datasetRemovalURL)
	if importErr != nil {
		return nil, errors.Join(importErr, notifyErr)
	}

	if notifyErr != nil {
		return nil, notifyErr
	}

	return providerkit.EncodeResult(ImportRecordsResult{IngestResult: result, Objects: result.Attempted}, ErrResultEncode)
}

// importDataset loads the dataset's manifest and records and ingests them in manifest order, returning the dataset name
// for the outcome notification
func importDataset(ctx context.Context, request types.OperationRequest, client *Client, cfg ImportRecords) (string, operations.IngestResult, error) {
	prefix, err := datasetPrefix(cfg.Prefix)
	if err != nil {
		return "", operations.IngestResult{}, err
	}

	manifest, err := loadDatasetManifest(ctx, client.Provider, prefix)
	if err != nil {
		return "", operations.IngestResult{}, err
	}

	payloadSets, err := loadDatasetPayloadSets(ctx, request.DB, client, prefix, manifest, cfg.Profile)
	if err != nil {
		return manifest.Name, operations.IngestResult{}, err
	}

	if err := ensureEntityTypes(ctx, request.DB, false, recordEntityTypeNames(payloadSets)); err != nil {
		return manifest.Name, operations.IngestResult{}, err
	}

	oc, _ := gala.OperationContextFromContext(ctx)

	ingest := operations.IngestContext{
		Registry:     request.Services.Definitions(),
		DB:           request.DB,
		Runtime:      request.Services.Gala(),
		Integration:  request.Integration,
		DefinitionID: DefinitionID.ID(),
	}

	result, err := operations.ProcessPayloadSets(ctx, ingest, ImportRecordsOp.Name(), importContracts, importPolicy, payloadSets, operations.IngestOptionsFromOperationContext(oc))

	return manifest.Name, result, err
}

// loadDatasetPayloadSets clones the manifest's standards into the organization and returns the dataset's records in
// manifest order, with their placeholders filled under the profile and each program linked to its cloned controls
func loadDatasetPayloadSets(ctx context.Context, db *generated.Client, client *Client, prefix string, manifest datasetManifest, profile string) ([]types.IngestPayloadSet, error) {
	variables, err := manifest.profileVariables(profile)
	if err != nil {
		return nil, err
	}

	builtins, err := datasetBuiltinVariables(ctx, db)
	if err != nil {
		return nil, err
	}

	fill := builtins.With(variables)

	programRefCodes := map[string][]string{}

	for _, std := range manifest.Standards {
		refCodes, err := cloneDatasetStandard(ctx, db, std)
		if err != nil {
			return nil, err
		}

		if std.ProgramExternalID != "" {
			programRefCodes[std.ProgramExternalID] = append(programRefCodes[std.ProgramExternalID], refCodes...)
		}
	}

	payloadSets := make([]types.IngestPayloadSet, 0, len(manifest.Order))

	for _, name := range manifest.Order {
		payloadSet, err := loadDatasetPayloadSet(ctx, client.Provider, prefix, name, fill)
		if err != nil {
			return nil, err
		}

		payloadSets = append(payloadSets, payloadSet)
	}

	if err := linkStandardControls(payloadSets, programRefCodes); err != nil {
		return nil, err
	}

	return payloadSets, nil
}

// linkStandardControls appends the ref codes of the standard controls cloned for each program to that program record's
// control ref codes, keyed by the program's external id
func linkStandardControls(payloadSets []types.IngestPayloadSet, programRefCodes map[string][]string) error {
	for i := range payloadSets {
		if payloadSets[i].Schema != entityops.SchemaProgram.Name {
			continue
		}

		for j, envelope := range payloadSets[i].Envelopes {
			externalID, _ := jsonx.DecodeObjectKey[string](envelope.Payload, recordExternalIDField)

			refCodes := programRefCodes[externalID]
			if len(refCodes) == 0 {
				continue
			}

			existing, _ := jsonx.DecodeObjectKey[[]string](envelope.Payload, recordControlRefCodesKey)

			payload, _, err := jsonx.SetObjectKey(envelope.Payload, recordControlRefCodesKey, append(existing, refCodes...))
			if err != nil {
				return fmt.Errorf("%w: %w", ErrContentEncode, err)
			}

			payloadSets[i].Envelopes[j].Payload = payload
		}
	}

	return nil
}

// datasetPrefix returns the dataset's bucket directory without a trailing slash, failing when it is not lowercase
// directory segments
func datasetPrefix(prefix string) (string, error) {
	dataset := strings.TrimSuffix(prefix, "/")
	if !datasetPattern.MatchString(dataset) {
		return "", fmt.Errorf("%w: %q", ErrDatasetInvalid, prefix)
	}

	return dataset, nil
}

// recordEntityTypeNames returns the distinct entity type names the records' entities link to, the vendor type for an
// entity that names none
func recordEntityTypeNames(payloadSets []types.IngestPayloadSet) []string {
	names := []string{}

	for _, set := range payloadSets {
		if set.Schema != entityops.SchemaEntity.Name {
			continue
		}

		for _, envelope := range set.Envelopes {
			name, _ := jsonx.DecodeObjectKey[string](envelope.Payload, recordEntityTypeKey)
			names = append(names, lo.CoalesceOrEmpty(name, vendorEntityTypeName))
		}
	}

	return lo.Uniq(names)
}

// cloneDatasetStandard clones the standard's controls into the organization and returns the ref codes of its controls
func cloneDatasetStandard(ctx context.Context, db *generated.Client, std datasetStandard) ([]string, error) {
	cloner := controls.Cloner{DB: db}
	filters := controls.CloneFilterOptions{StandardShortName: &std.ShortName, Categories: std.Categories}

	cloned, err := cloner.FromStandard(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrDatasetStandardClone, std.ShortName, err)
	}

	return lo.Map(cloned, func(c *generated.Control, _ int) string { return c.RefCode }), nil
}

// datasetBuiltinVariables returns the built-in variables every dataset can use, filled from the importing organization
// and the requesting user
func datasetBuiltinVariables(ctx context.Context, db *generated.Client) (placeholder.Variables, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok {
		return nil, auth.ErrNoAuthUser
	}

	ownerID, ok := caller.ActiveOrg()
	if !ok {
		return nil, ErrDatasetOwnerRequired
	}

	org, err := db.Organization.Get(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDatasetOrganizationLookup, err)
	}

	domain := datasetDefaultDomain

	setting, err := db.OrganizationSetting.Query().Where(organizationsetting.OrganizationID(ownerID)).Only(ctx)
	switch {
	case err == nil && len(setting.Domains) > 0:
		domain = setting.Domains[0]
	case err != nil && !generated.IsNotFound(err):
		return nil, fmt.Errorf("%w: %w", ErrDatasetOrganizationLookup, err)
	}

	return DatasetBuiltinVariables{
		OrgName:     lo.CoalesceOrEmpty(org.DisplayName, org.Name),
		OrgSlug:     lo.CoalesceOrEmpty(org.SlugName, urlx.Slug(org.Name)),
		OrgDomain:   domain,
		OwnerUserID: caller.SubjectID,
		OwnerEmail:  caller.SubjectEmail,
		OwnerName:   lo.CoalesceOrEmpty(caller.SubjectName, caller.SubjectEmail),
		Today:       time.Now().UTC().Format(time.DateOnly),
	}.variables()
}

// variables returns the built-ins keyed by their variable names
func (b DatasetBuiltinVariables) variables() (placeholder.Variables, error) {
	var variables placeholder.Variables
	if err := jsonx.RoundTrip(b, &variables); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrContentEncode, err)
	}

	return variables, nil
}

// loadDatasetPayloadSet reads one schema file of the dataset, fills its placeholders, and wraps its records as that
// schema's payload set, keyed by the dataset prefix as its source key
func loadDatasetPayloadSet(ctx context.Context, provider storage.Provider, prefix, name string, fill placeholder.Variables) (types.IngestPayloadSet, error) {
	schema, ok := entityops.LookupSchema(name)
	if !ok {
		return types.IngestPayloadSet{}, fmt.Errorf("%w: %s", ErrDatasetSchemaUnknown, name)
	}

	key := path.Join(prefix, name+jsonObjectSuffix)

	records, err := loadRecords(ctx, provider, key)
	if err != nil {
		return types.IngestPayloadSet{}, err
	}

	envelopes := lo.Map(records, func(record json.RawMessage, idx int) types.MappingEnvelope {
		return providerkit.RawEnvelope(fmt.Sprintf("%s#%d", key, idx), json.RawMessage(fill.Fill(string(record))))
	})

	return types.IngestPayloadSet{Schema: schema.Name, Envelopes: envelopes, SourceKey: prefix}, nil
}

// loadDatasetManifest reads the dataset's manifest
func loadDatasetManifest(ctx context.Context, provider storage.Provider, prefix string) (datasetManifest, error) {
	key := path.Join(prefix, datasetManifestName)

	document, err := downloadObject(ctx, provider, key)
	if err != nil {
		return datasetManifest{}, err
	}

	var manifest datasetManifest
	if err := json.Unmarshal(document, &manifest); err != nil {
		return datasetManifest{}, fmt.Errorf("%w: %s: %w", ErrDatasetManifestInvalid, key, err)
	}

	return manifest, nil
}
