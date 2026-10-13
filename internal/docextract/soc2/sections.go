package soc2

import (
	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/docextract/schema"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// The keys of the extras a section carries to link its records, which the entities have no field for
const (
	// refCodesKey lists the control ref codes a review or finding is recorded against
	refCodesKey = "ref_codes"
	// testDetailsKey is the field on a finding carrying the test procedure printed in its table row
	testDetailsKey = "test_details"
	// ReviewExternalIDKey is the field on a finding naming the review whose test produced it
	ReviewExternalIDKey = "review_external_id"
	// soc2MappingKey lists the SOC 2 criteria a control maps to
	soc2MappingKey = "soc2_mapping"
	// entityNameKey names the vendor an asset belongs to
	entityNameKey = "entity_name"
	// entityNamesKey names the vendors a contact works for
	entityNamesKey = "entity_names"
	// platformNameKey names the platform a system belongs to
	platformNameKey = "platform_name"
)

// ProgramSchema asks for the audited program the report covers
var ProgramSchema = schema.Section("programs",
	schema.From(entityops.ProgramFields.Name).Required(),
	schema.From(entityops.ProgramFields.Description).Required(),
	schema.From(entityops.ProgramFields.AuditFirm),
	schema.From(entityops.ProgramFields.Auditor),
	schema.From(entityops.ProgramFields.AuditorEmail),
	schema.From(entityops.ProgramFields.Tags),
)

// DomainSchema asks for the trust services categories the audit tested
var DomainSchema = schema.Section("domains",
	schema.Extra("name", docextract.TypeString).Required(),
)

// EntitySchema asks for the vendors and subservice organizations the report names
var EntitySchema = schema.Section("entities",
	schema.From(entityops.EntityFields.Name).Required(),
	schema.From(entityops.EntityFields.DisplayName).Required(),
	schema.From(entityops.EntityFields.Description).Required(),
	schema.From(entityops.EntityFields.ApprovedForUse).Required(),
	schema.From(entityops.EntityFields.HasSoc2),
	schema.From(entityops.EntityFields.Domains),
	schema.From(entityops.EntityFields.SSOEnforced),
	schema.From(entityops.EntityFields.MfaSupported),
	schema.From(entityops.EntityFields.MfaEnforced),
	schema.From(entityops.EntityFields.StatusPageURL),
	schema.From(entityops.EntityFields.ProvidedServices),
)

// AssetSchema asks for the systems and infrastructure in the audited environment
var AssetSchema = schema.Section("assets",
	schema.From(entityops.AssetFields.Name).Required(),
	schema.From(entityops.AssetFields.DisplayName).Required(),
	schema.From(entityops.AssetFields.Description).Required(),
	schema.From(entityops.AssetFields.ContainsPii).Required(),
	schema.From(entityops.AssetFields.AssetType).Required().Enum(enums.AssetTypeTechnology.String(), enums.AssetTypeDomain.String(), enums.AssetTypeDevice.String(), enums.AssetTypeTelephone.String()),
	schema.From(entityops.AssetFields.PhysicalLocation),
	schema.From(entityops.AssetFields.Region),
	schema.From(entityops.AssetFields.Categories),
	schema.Extra(entityNameKey, docextract.TypeString),
)

// ControlSchema asks for the controls the auditor tested
var ControlSchema = schema.Section("controls",
	schema.From(entityops.ControlFields.RefCode).Required(),
	schema.From(entityops.ControlFields.Title).Required(),
	schema.From(entityops.ControlFields.Description).Required(),
	schema.From(entityops.ControlFields.AuditorReferenceID),
	schema.From(entityops.ControlFields.Category).Required(),
	schema.From(entityops.ControlFields.Subcategory).Required(),
	schema.Extra(soc2MappingKey, docextract.TypeArray),
)

// GroupSchema asks for the teams and roles responsible for the controls
var GroupSchema = schema.Section("groups",
	schema.From(entityops.GroupFields.Name).Required(),
	schema.From(entityops.GroupFields.DisplayName).Required(),
)

// PlatformSchema asks for the platforms the audited system runs on
var PlatformSchema = schema.Section("platforms",
	schema.From(entityops.PlatformFields.Name).Required(),
	schema.From(entityops.PlatformFields.Description).Required(),
	schema.From(entityops.PlatformFields.BusinessPurpose).Required(),
	schema.From(entityops.PlatformFields.ContainsPii).Required(),
	schema.From(entityops.PlatformFields.PhysicalLocation),
	schema.From(entityops.PlatformFields.Region),
)

// SystemDetailsSchema asks for the systems described in the report and the platform each runs on
var SystemDetailsSchema = schema.Section("systemdetails",
	schema.From(entityops.SystemDetailFields.SystemName).Required(),
	schema.From(entityops.SystemDetailFields.Description).Required(),
	schema.Extra(platformNameKey, docextract.TypeString).Required(),
)

// ContactSchema asks for the people the report names and the vendors they work for
var ContactSchema = schema.Section("contacts",
	schema.From(entityops.ContactFields.Company).Required(),
	schema.From(entityops.ContactFields.Email).Required(),
	schema.From(entityops.ContactFields.FullName).Required(),
	schema.From(entityops.ContactFields.Title).Required(),
	schema.Extra(entityNamesKey, docextract.TypeArray),
)

// ReviewSchema asks for the auditor's tests of each control and their results
var ReviewSchema = schema.Section("reviews",
	schema.From(entityops.ReviewFields.Title).Required(),
	schema.From(entityops.ReviewFields.Summary).Required(),
	schema.From(entityops.ReviewFields.Details).Required(),
	schema.From(entityops.ReviewFields.Source),
	schema.From(entityops.ReviewFields.ReportedAt),
	schema.From(entityops.ReviewFields.ReviewedAt),
	schema.From(entityops.ReviewFields.Reporter),
	schema.From(entityops.ReviewFields.ApprovedAt),
	schema.From(entityops.ReviewFields.Approved).Required(),
	schema.From(entityops.ReviewFields.ExternalID),
	schema.Extra(refCodesKey, docextract.TypeArray),
)

// FindingSchema asks for the exceptions and deviations the auditor noted against controls
var FindingSchema = schema.Section("findings",
	schema.From(entityops.FindingFields.Open).Required(),
	schema.From(entityops.FindingFields.Description).Required(),
	schema.From(entityops.FindingFields.Severity).Required().Enum("high", "medium", "low"),
	schema.From(entityops.FindingFields.Source),
	schema.From(entityops.FindingFields.ReportedAt),
	schema.Extra("reporter", docextract.TypeString),
	schema.Extra(refCodesKey, docextract.TypeArray).Required(),
	schema.Extra(testDetailsKey, docextract.TypeString).Describe("The verbatim text of the test procedure printed in the same table row as this exception, copied exactly and not summarized"),
)
