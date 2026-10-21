// Package soc2schema holds the response schema of each SOC 2 report section for passing to soc2.NewKind
// record sections are built from the Openlane entity fields they import into
// only judgment sections such as the assessment are written by hand
package soc2schema

import (
	"github.com/invopop/jsonschema"
	"github.com/theopenlane/precis/docextract/kinds/soc2"
	"github.com/theopenlane/precis/docextract/schema"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// ProgramSchema asks for the audited program the report covers
var ProgramSchema = schema.Section(soc2.ProgramsPart,
	from(entityops.ProgramFields.Name).Required(),
	from(entityops.ProgramFields.Description).Required(),
	from(entityops.ProgramFields.AuditFirm),
	from(entityops.ProgramFields.Auditor),
	from(entityops.ProgramFields.AuditorEmail),
	from(entityops.ProgramFields.Tags),
)

// EntitySchema asks for the vendors and subservice organizations the report names
var EntitySchema = schema.Section(soc2.EntitiesPart,
	from(entityops.EntityFields.Name).Required(),
	from(entityops.EntityFields.DisplayName).Required(),
	from(entityops.EntityFields.Description).Required(),
	from(entityops.EntityFields.ApprovedForUse).Required(),
	from(entityops.EntityFields.HasSoc2),
	from(entityops.EntityFields.Domains),
	from(entityops.EntityFields.SSOEnforced),
	from(entityops.EntityFields.MfaSupported),
	from(entityops.EntityFields.MfaEnforced),
	from(entityops.EntityFields.StatusPageURL),
	from(entityops.EntityFields.ProvidedServices),
)

// AssetSchema asks for the systems and infrastructure in the audited environment
var AssetSchema = schema.Section(soc2.AssetsPart,
	from(entityops.AssetFields.Name).Required(),
	from(entityops.AssetFields.DisplayName).Required(),
	from(entityops.AssetFields.Description).Required(),
	from(entityops.AssetFields.ContainsPii).Required(),
	from(entityops.AssetFields.AssetType).Required().Enum(enums.AssetTypeTechnology.String(), enums.AssetTypeDomain.String(), enums.AssetTypeDevice.String(), enums.AssetTypeTelephone.String()),
	from(entityops.AssetFields.PhysicalLocation),
	from(entityops.AssetFields.Region),
	from(entityops.AssetFields.Categories),
	schema.Extra(soc2.EntityNameKey, schema.TypeString),
)

// ControlSchema asks for the controls the auditor tested
var ControlSchema = schema.Section(soc2.ControlsPart,
	from(entityops.ControlFields.RefCode).Required(),
	from(entityops.ControlFields.Title).Required(),
	from(entityops.ControlFields.Description).Required(),
	from(entityops.ControlFields.AuditorReferenceID),
	from(entityops.ControlFields.Category).Required(),
	from(entityops.ControlFields.Subcategory).Required(),
	schema.Extra(soc2.MappingKey, schema.TypeArray),
)

// GroupSchema asks for the teams and roles responsible for the controls
var GroupSchema = schema.Section(soc2.GroupsPart,
	from(entityops.GroupFields.Name).Required(),
	from(entityops.GroupFields.DisplayName).Required(),
)

// PlatformSchema asks for the platforms the audited system runs on
var PlatformSchema = schema.Section(soc2.PlatformsPart,
	from(entityops.PlatformFields.Name).Required(),
	from(entityops.PlatformFields.Description).Required(),
	from(entityops.PlatformFields.BusinessPurpose).Required(),
	from(entityops.PlatformFields.ContainsPii).Required(),
	from(entityops.PlatformFields.PhysicalLocation),
	from(entityops.PlatformFields.Region),
)

// SystemDetailsSchema asks for the systems described in the report and the platform each runs on
var SystemDetailsSchema = schema.Section(soc2.SystemDetailsPart,
	from(entityops.SystemDetailFields.SystemName).Required(),
	from(entityops.SystemDetailFields.Description).Required(),
	schema.Extra(soc2.PlatformNameKey, schema.TypeString).Required(),
)

// ContactSchema asks for the people the report names and the vendors they work for
var ContactSchema = schema.Section(soc2.ContactsPart,
	from(entityops.ContactFields.Company).Required(),
	from(entityops.ContactFields.Email).Required(),
	from(entityops.ContactFields.FullName).Required(),
	from(entityops.ContactFields.Title).Required(),
	schema.Extra(soc2.EntityNamesKey, schema.TypeArray),
)

// ReviewSchema asks for the auditor's tests of each control and their results
var ReviewSchema = schema.Section(soc2.ReviewsPart,
	from(entityops.ReviewFields.Title).Required(),
	from(entityops.ReviewFields.Summary).Required(),
	from(entityops.ReviewFields.Details).Required(),
	from(entityops.ReviewFields.Source),
	from(entityops.ReviewFields.ReportedAt),
	from(entityops.ReviewFields.ReviewedAt),
	from(entityops.ReviewFields.Reporter),
	from(entityops.ReviewFields.ApprovedAt),
	from(entityops.ReviewFields.Approved).Required(),
	from(entityops.ReviewFields.ExternalID),
	schema.Extra(soc2.RefCodesKey, schema.TypeArray),
)

// FindingSchema asks for the exceptions and deviations the auditor noted against controls
var FindingSchema = schema.Section(soc2.FindingsPart,
	from(entityops.FindingFields.Open).Required(),
	from(entityops.FindingFields.Description).Required(),
	from(entityops.FindingFields.Severity).Required().Enum("high", "medium", "low"),
	from(entityops.FindingFields.Source),
	from(entityops.FindingFields.ReportedAt),
	schema.Extra("reporter", schema.TypeString),
	schema.Extra(soc2.RefCodesKey, schema.TypeArray).Required(),
	schema.Extra(soc2.TestDetailsKey, schema.TypeString).Describe("The verbatim text of the test procedure printed in the same table row as this exception, copied exactly and not summarized"),
	schema.Extra(soc2.ManagementResponseKey, schema.TypeString).Describe("Management's response to the exception when the report gives one, copied from the report"),
)

// Schemas maps each SOC 2 part to the response schema the model must follow
var Schemas = map[string]*jsonschema.Schema{
	soc2.EntitiesPart:      EntitySchema,
	soc2.AssetsPart:        AssetSchema,
	soc2.ControlsPart:      ControlSchema,
	soc2.GroupsPart:        GroupSchema,
	soc2.ReviewsPart:       ReviewSchema,
	soc2.FindingsPart:      FindingSchema,
	soc2.PlatformsPart:     PlatformSchema,
	soc2.SystemDetailsPart: SystemDetailsSchema,
	soc2.ContactsPart:      ContactSchema,
	soc2.ProgramsPart:      ProgramSchema,
	soc2.AssessmentPart:    AssessmentSchema,
}

// fields reads each section's field list from the schemas alone, without the configured prompts
var fields = soc2.NewKind(soc2.Prompts{}, Schemas)

// SectionFields lists the fields the named section's items carry, so an import keeps only what the model was asked for
func SectionFields(name string) []string {
	return fields.SectionFields(name)
}
