package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/theopenlane/precis/docextract/kinds/soc2"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/docextract/soc2schema"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/ent/generated/platform"
	"github.com/theopenlane/core/v2/internal/ent/generated/review"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/utils/mapx"
)

const (
	// reportScanImportVariant is the mapping variant a report import ingests its records under
	reportScanImportVariant = "SOC2_REPORT"
	// soc2StandardShortName is the short name of the standard the program's framework controls are copied from
	soc2StandardShortName = "SOC 2"
	// reportImportVendorEntityType is the entity type imported vendors are created as
	reportImportVendorEntityType = "vendor"
	// reportImportEnvironment is the environment imported records are placed in
	reportImportEnvironment = "production"
	// reportImportScope is the scope imported records are placed in
	reportImportScope = "in-scope"
	// reportImportReviewState is the state imported reviews are created in
	reportImportReviewState = "COMPLETED"
	// reportImportReviewCategory is the category imported reviews are created in
	reportImportReviewCategory = "audit-review"
	// mappedControlKeySeparator joins a report control's ref code and a criterion into a mapping's external id
	mappedControlKeySeparator = ">"
)

// reportImportProgramNamespace seeds the program's external uuid, so a report always resolves to the same program
var reportImportProgramNamespace = uuid.MustParse("3c4a1f6e-9d2b-5a7c-8e1f-0b6d2c9a4e57")

// reportImportKeys names the field the reviewer accepts and edits each section's items by
var reportImportKeys = map[string]string{
	soc2.ProgramsPart:      entityops.ProgramFields.Name.Name,
	soc2.EntitiesPart:      entityops.EntityFields.Name.Name,
	soc2.AssetsPart:        entityops.AssetFields.Name.Name,
	soc2.ControlsPart:      entityops.ControlFields.RefCode.Name,
	soc2.GroupsPart:        entityops.GroupFields.Name.Name,
	soc2.PlatformsPart:     entityops.PlatformFields.Name.Name,
	soc2.SystemDetailsPart: entityops.SystemDetailFields.SystemName.Name,
	soc2.ContactsPart:      entityops.ContactFields.Email.Name,
	soc2.ReviewsPart:       entityops.ReviewFields.ExternalID.Name,
	soc2.FindingsPart:      entityops.FindingFields.ExternalID.Name,
}

// reportImportComputedFields are the fields the saga adds to a section after extraction, kept alongside the
// fields the section's schema defines
var reportImportComputedFields = map[string][]string{
	soc2.FindingsPart: {entityops.FindingFields.ExternalID.Name, soc2.ReviewExternalIDKey},
}

// reportScanImportContracts are the schemas a report import ingests, in dependency order
var reportScanImportContracts = []types.IngestContract{
	{Schema: entityops.SchemaControl.Name},
	{Schema: entityops.SchemaEntity.Name},
	{Schema: entityops.SchemaAsset.Name},
	{Schema: entityops.SchemaGroup.Name},
	{Schema: entityops.SchemaPlatform.Name},
	{Schema: entityops.SchemaSystemDetail.Name},
	{Schema: entityops.SchemaContact.Name},
	{Schema: entityops.SchemaReview.Name},
	{Schema: entityops.SchemaFinding.Name},
	{Schema: entityops.SchemaMappedControl.Name},
}

// The payload keys listing the target keys each link resolves to ids at ingest
const (
	controlRefCodesField     = "control_ref_codes"
	vendorKeysField          = "vendor_keys"
	platformNamesField       = "platform_names"
	reviewExternalIDsField   = "review_external_ids"
	fromControlRefCodesField = "from_control_ref_codes"
)

// reportScanImportMappings returns the identity mappings a report import ingests through; each payload carries
// its links as lists of target keys, resolved to ids at ingest
func reportScanImportMappings() []types.MappingRegistration {
	toControls := types.LinkRule{TargetSchema: entityops.SchemaControl.Name, Edge: "controls", TargetField: control.FieldRefCode, SourceList: controlRefCodesField, PayloadKeys: true}
	toVendors := types.LinkRule{TargetSchema: entityops.SchemaEntity.Name, Edge: "entities", TargetField: entity.FieldExternalID, SourceList: vendorKeysField, PayloadKeys: true}

	return []types.MappingRegistration{
		reportScanImportMapping(entityops.SchemaControl.Name),
		reportScanImportMapping(entityops.SchemaEntity.Name),
		reportScanImportMapping(entityops.SchemaAsset.Name, toVendors),
		reportScanImportMapping(entityops.SchemaGroup.Name),
		reportScanImportMapping(entityops.SchemaPlatform.Name),
		reportScanImportMapping(entityops.SchemaSystemDetail.Name, types.LinkRule{TargetSchema: entityops.SchemaPlatform.Name, Edge: "platforms", TargetField: platform.FieldName, SourceList: platformNamesField, PayloadKeys: true}),
		reportScanImportMapping(entityops.SchemaContact.Name, toVendors),
		reportScanImportMapping(entityops.SchemaReview.Name, toControls),
		reportScanImportMapping(entityops.SchemaFinding.Name, toControls, types.LinkRule{TargetSchema: entityops.SchemaReview.Name, Edge: "reviews", TargetField: review.FieldExternalID, SourceList: reviewExternalIDsField, PayloadKeys: true}),
		reportScanImportMapping(entityops.SchemaMappedControl.Name, types.LinkRule{TargetSchema: entityops.SchemaControl.Name, Edge: "from_controls", TargetField: control.FieldRefCode, SourceList: fromControlRefCodesField, PayloadKeys: true}),
	}
}

// reportScanImportMapping registers one schema's report import variant, passing its payload through as-is
func reportScanImportMapping(schema string, links ...types.LinkRule) types.MappingRegistration {
	return types.MappingRegistration{
		Schema:  schema,
		Variant: reportScanImportVariant,
		Spec:    types.MappingOverride{FilterExpr: "true", Links: links},
	}
}

// reportImport carries what every section's payloads need once the program and its framework controls exist
type reportImport struct {
	// programID is the program imported records belong to, empty when the reviewer accepted none
	programID string
	// vendorTypeID is the entity type imported vendors are created as
	vendorTypeID string
	// vendorKeys maps a vendor's name and display name to the external id it is imported by
	vendorKeys map[string]string
	// existingVendors maps a vendor domain to the external id of the organization entity already carrying it
	existingVendors map[string]string
	// frameworkIDs maps a SOC 2 criterion to the framework control copied into the program for it
	frameworkIDs map[string]string
}

// ingestReportScanImport sets up the program and its framework controls, then builds the payload sets for the
// accepted items in dependency order, ending with the mappings from report controls to the framework
func ingestReportScanImport(ctx context.Context, request types.OperationRequest, cfg ReportScanImport) ([]types.IngestPayloadSet, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{"organization_id": cfg.OrganizationID, "scan_id": cfg.ScanID})

	client := request.DB

	scanRecord, err := client.Scan.Get(scanSystemContext(ctx, cfg.OrganizationID), cfg.ScanID)
	if err != nil {
		return nil, err
	}

	reportFileID, ok := ReportFileID(scanRecord.Metadata)
	if !ok {
		return nil, ErrReportFileMissing
	}

	report, _ := scanRecord.Metadata[reportMetadataKey].(map[string]any)
	accepted := acceptedSections(report, cfg)

	programID, err := upsertImportProgram(ctx, client, cfg.OrganizationID, reportFileID, accepted[soc2.ProgramsPart])
	if err != nil {
		return nil, err
	}

	frameworkIDs, err := copyFrameworkControls(ctx, client, cfg.OrganizationID, programID, accepted[soc2.ControlsPart])
	if err != nil {
		return nil, err
	}

	vendorTypeID, err := client.EntityType.Query().Where(entitytype.NameEqualFold(reportImportVendorEntityType)).OnlyID(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up vendor entity type: %w", err)
	}

	existingVendors, err := operations.EntityExternalIDsByDomain(ctx, client, lo.Uniq(lo.FilterMap(accepted[soc2.EntitiesPart], func(vendor map[string]any, _ int) (string, bool) {
		domain := vendorDomain(vendor)

		return domain, domain != ""
	})))
	if err != nil {
		return nil, err
	}

	imp := reportImport{programID: programID, vendorTypeID: vendorTypeID, vendorKeys: importVendorKeys(accepted[soc2.EntitiesPart], existingVendors), existingVendors: existingVendors, frameworkIDs: frameworkIDs}

	sections := []struct {
		schema  string
		section string
		build   func(map[string]any) map[string]any
	}{
		{entityops.SchemaControl.Name, soc2.ControlsPart, imp.control},
		{entityops.SchemaEntity.Name, soc2.EntitiesPart, imp.vendor},
		{entityops.SchemaAsset.Name, soc2.AssetsPart, imp.asset},
		{entityops.SchemaGroup.Name, soc2.GroupsPart, imp.group},
		{entityops.SchemaPlatform.Name, soc2.PlatformsPart, imp.platform},
		{entityops.SchemaSystemDetail.Name, soc2.SystemDetailsPart, imp.systemDetail},
		{entityops.SchemaContact.Name, soc2.ContactsPart, imp.contact},
		{entityops.SchemaReview.Name, soc2.ReviewsPart, imp.review},
		{entityops.SchemaFinding.Name, soc2.FindingsPart, imp.finding},
	}

	sets := make([]types.IngestPayloadSet, 0, len(sections)+1)

	for _, s := range sections {
		set, err := importPayloadSet(s.schema, s.section, accepted[s.section], reportFileID, s.build)
		if err != nil {
			return nil, err
		}

		sets = append(sets, set)
	}

	mappings, err := imp.mappedControls(accepted[soc2.ControlsPart], reportFileID)
	if err != nil {
		return nil, err
	}

	return append(sets, mappings), nil
}

// afterReportScanImport notifies the organization of the import outcome
func afterReportScanImport(ctx context.Context, request types.OperationRequest, cfg ReportScanImport, outcome types.IngestOutcome) error {
	return operations.NotifyImportOutcome(scanSystemContext(ctx, cfg.OrganizationID), request.DB, cfg.OrganizationID, "Report", outcome, "")
}

// acceptedSections returns, per section, the stored items the reviewer accepted with their edits applied, keeping
// only the fields the section's schema defines so neither the model nor an edit can set anything else
func acceptedSections(report map[string]any, cfg ReportScanImport) map[string][]map[string]any {
	accepted := make(map[string][]map[string]any, len(reportImportKeys))

	for section, keyField := range reportImportKeys {
		keys := cfg.Accepted[section]
		if len(keys) == 0 {
			continue
		}

		fields := append(soc2schema.SectionFields(section), reportImportComputedFields[section]...)
		edited := lo.SliceToMap(cfg.Edited[section], func(item map[string]any) (string, map[string]any) { return mapx.Value[string](item, keyField), item })
		for _, item := range mapx.Maps(report, section) {
			key := mapx.Value[string](item, keyField)
			if !lo.Contains(keys, key) {
				continue
			}

			if replacement, ok := edited[key]; ok {
				item = replacement
			}

			accepted[section] = append(accepted[section], lo.PickByKeys(item, fields))
		}
	}

	return accepted
}

// addProgram links a payload to the imported program when there is one
func (r reportImport) addProgram(payload map[string]any) {
	if r.programID != "" {
		payload["program_ids"] = []string{r.programID}
	}
}

// upsertImportProgram creates the program the report covers, or finds the one an earlier import of the same report
// created, through the same ingest persistence the other records use; returns empty when no program was accepted
func upsertImportProgram(ctx context.Context, client *generated.Client, organizationID, reportFileID string, programs []map[string]any) (string, error) {
	if len(programs) == 0 {
		return "", nil
	}

	payload := programs[0]
	payload[entityops.ProgramFields.ExternalUUID.Name] = uuid.NewSHA1(reportImportProgramNamespace, []byte(reportFileID)).String()
	payload[entityops.ProgramFields.FrameworkName.Name] = soc2StandardShortName
	payload[entityops.ProgramFields.ProgramKindName.Name] = enums.ProgramTypeFramework.String()
	payload[entityops.ProgramFields.Status.Name] = enums.ProgramStatusNotStarted
	payload[entityops.SchemaProgram.OwnerField] = organizationID
	payload[entityops.FieldSourceDefinitionID] = DefinitionID.ID()
	payload[entityops.FieldSourceInstanceID] = reportFileID

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", ErrPayloadEncode
	}

	id, _, _, err := entityops.SchemaProgram.PersistIngest(ctx, client, nil, raw)

	return id, err
}

// copyFrameworkControls copies the SOC 2 controls the accepted report controls map to into the org, linked to the
// program when there is one, returning each copied control's id by its criterion
func copyFrameworkControls(ctx context.Context, client *generated.Client, organizationID, programID string, reportControls []map[string]any) (map[string]string, error) {
	criteria := lo.Uniq(lo.FlatMap(reportControls, func(item map[string]any, _ int) []string { return mapx.Strings(item, soc2.MappingKey) }))
	if len(criteria) == 0 {
		return map[string]string{}, nil
	}

	shortName := soc2StandardShortName
	cloner := controls.Cloner{DB: client, OrgID: organizationID, SkipProgramAccess: true}

	copied, err := cloner.FromStandard(ctx, controls.CloneFilterOptions{StandardShortName: &shortName, RefCodes: criteria}, lo.EmptyableToPtr(programID))
	if err != nil {
		return nil, err
	}

	return lo.SliceToMap(copied, func(c *generated.Control) (string, string) { return c.RefCode, c.ID }), nil
}

// importVendorKeys maps each accepted vendor's name and display name to the external id it is imported by
func importVendorKeys(vendors []map[string]any, existing map[string]string) map[string]string {
	keys := make(map[string]string, len(vendors)*2) //nolint:mnd

	for _, vendor := range vendors {
		key := vendorExternalID(vendor, existing)

		for _, field := range []string{entityops.EntityFields.Name.Name, entityops.EntityFields.DisplayName.Name} {
			if name := mapx.Value[string](vendor, field); name != "" {
				keys[strings.ToLower(name)] = key
			}
		}
	}

	return keys
}

// vendorExternalID is the external id a vendor is imported by: the existing entity's for its first domain, else the
// domain itself, else its name, lowercased
func vendorExternalID(vendor map[string]any, existing map[string]string) string {
	if domain := vendorDomain(vendor); domain != "" {
		return lo.CoalesceOrEmpty(existing[domain], domain)
	}

	name := mapx.Value[string](vendor, entityops.EntityFields.Name.Name)

	return strings.ToLower(strings.TrimSpace(name))
}

// vendorDomain returns a vendor's first domain, normalized, empty when it lists none
func vendorDomain(vendor map[string]any) string {
	domains := mapx.Strings(vendor, entityops.EntityFields.Domains.Name)
	if len(domains) == 0 {
		return ""
	}

	return operations.NormalizeDomain(domains[0])
}

// importPayloadSet builds one section's payload set, each accepted item shaped by build into its create payload
func importPayloadSet(schema, section string, items []map[string]any, reportFileID string, build func(map[string]any) map[string]any) (types.IngestPayloadSet, error) {
	envelopes := make([]types.MappingEnvelope, 0, len(items))

	for _, item := range items {
		envelope, err := providerkit.MarshalEnvelopeVariant(reportScanImportVariant, mapx.Value[string](item, reportImportKeys[section]), build(item), ErrPayloadEncode)
		if err != nil {
			return types.IngestPayloadSet{}, err
		}

		envelopes = append(envelopes, envelope)
	}

	return types.IngestPayloadSet{Schema: schema, Envelopes: envelopes, SourceInstanceID: reportFileID}, nil
}

// control places an auditor's control in the program as an approved, imported control
func (r reportImport) control(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{soc2.MappingKey})
	payload[entityops.ControlFields.Status.Name] = enums.ControlStatusApproved
	payload[entityops.ControlFields.Source.Name] = enums.ControlSourceImport
	payload[entityops.ControlFields.EnvironmentName.Name] = reportImportEnvironment
	payload[entityops.ControlFields.ScopeName.Name] = reportImportScope
	r.addProgram(payload)

	return payload
}

// vendor imports a vendor as an active, in-scope vendor entity keyed by its external id, which for a vendor the
// organization already has is that entity's, so the lookup updates it instead of creating a second one
func (r reportImport) vendor(item map[string]any) map[string]any {
	payload := lo.Assign(item)
	payload[entityops.EntityFields.ExternalID.Name] = vendorExternalID(item, r.existingVendors)

	// the existing entity already carries the domain, setting the list would replace its others
	if _, ok := r.existingVendors[vendorDomain(item)]; ok {
		delete(payload, entityops.EntityFields.Domains.Name)
	}

	payload[entityops.EntityFields.Status.Name] = enums.EntityStatusActive
	payload[entity.FieldEntityTypeID] = r.vendorTypeID
	payload[entityops.EntityFields.EnvironmentName.Name] = reportImportEnvironment
	payload[entityops.EntityFields.ScopeName.Name] = reportImportScope

	return payload
}

// asset imports an asset keyed by its name and linked to the vendor that provides it
func (r reportImport) asset(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{"entity_name"})
	payload[entityops.AssetFields.SourceIdentifier.Name] = item[entityops.AssetFields.Name.Name]
	payload[entityops.AssetFields.EnvironmentName.Name] = reportImportEnvironment
	payload[entityops.AssetFields.ScopeName.Name] = reportImportScope

	if vendor := mapx.Value[string](item, "entity_name"); vendor != "" {
		if key, ok := r.vendorKeys[strings.ToLower(vendor)]; ok {
			payload[vendorKeysField] = []string{key}
		}
	}

	return payload
}

// group imports a group as extracted
func (reportImport) group(item map[string]any) map[string]any {
	return lo.Assign(item)
}

// platform imports a platform in the production, in-scope environment
func (reportImport) platform(item map[string]any) map[string]any {
	payload := lo.Assign(item)
	payload[entityops.PlatformFields.EnvironmentName.Name] = reportImportEnvironment
	payload[entityops.PlatformFields.ScopeName.Name] = reportImportScope

	return payload
}

// systemDetail imports a system in the program, parented to the platform it runs on
func (r reportImport) systemDetail(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{"platform_name"})
	r.addProgram(payload)

	if name := mapx.Value[string](item, "platform_name"); name != "" {
		payload[platformNamesField] = []string{name}
	}

	return payload
}

// contact imports a contact linked to the vendors they work for
func (r reportImport) contact(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{"entity_names"})
	names := mapx.Strings(item, "entity_names")

	keys := lo.Uniq(lo.FilterMap(names, func(name string, _ int) (string, bool) {
		key, ok := r.vendorKeys[strings.ToLower(name)]

		return key, ok
	}))

	if len(keys) > 0 {
		payload[vendorKeysField] = keys
	}

	return payload
}

// review imports the auditor's test as a completed audit review of the controls it covered
func (r reportImport) review(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{"ref_codes"})
	payload[entityops.ReviewFields.State.Name] = reportImportReviewState
	payload[entityops.ReviewFields.Category.Name] = reportImportReviewCategory
	payload[entityops.ReviewFields.EnvironmentName.Name] = reportImportEnvironment
	payload[entityops.ReviewFields.ScopeName.Name] = reportImportScope
	r.addProgram(payload)

	if refCodes, ok := item["ref_codes"]; ok {
		payload[controlRefCodesField] = refCodes
	}

	return payload
}

// finding imports an exception against the controls it was raised on and the review whose test found it
func (r reportImport) finding(item map[string]any) map[string]any {
	payload := lo.OmitByKeys(item, []string{"ref_codes", "test_details", "reporter", soc2.ReviewExternalIDKey})
	payload[entityops.FindingFields.EnvironmentName.Name] = reportImportEnvironment
	r.addProgram(payload)

	if refCodes, ok := item["ref_codes"]; ok {
		payload[controlRefCodesField] = refCodes
	}

	if reviewID := mapx.Value[string](item, soc2.ReviewExternalIDKey); reviewID != "" {
		payload[reviewExternalIDsField] = []string{reviewID}
	}

	return payload
}

// mappedControls builds one subset mapping from each accepted report control to every framework control its SOC 2
// criteria were copied as, keyed by the pair so a re-import finds the mappings it created
func (r reportImport) mappedControls(controls []map[string]any, reportFileID string) (types.IngestPayloadSet, error) {
	var envelopes []types.MappingEnvelope

	for _, item := range controls {
		refCode := mapx.Value[string](item, entityops.ControlFields.RefCode.Name)

		for _, criterion := range mapx.Strings(item, soc2.MappingKey) {
			frameworkID, ok := r.frameworkIDs[criterion]
			if !ok || refCode == "" {
				continue
			}

			key := refCode + mappedControlKeySeparator + criterion

			envelope, err := providerkit.MarshalEnvelopeVariant(reportScanImportVariant, key, map[string]any{
				entityops.MappedControlFields.ExternalID.Name:  key,
				entityops.MappedControlFields.MappingType.Name: enums.MappingTypeSubset,
				entityops.MappedControlFields.Source.Name:      enums.MappingSourceImported,
				fromControlRefCodesField:                       []string{refCode},
				"to_control_ids":                               []string{frameworkID},
			}, ErrPayloadEncode)
			if err != nil {
				return types.IngestPayloadSet{}, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return types.IngestPayloadSet{Schema: entityops.SchemaMappedControl.Name, Envelopes: envelopes, SourceInstanceID: reportFileID}, nil
}
