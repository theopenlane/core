package schema

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

func TestSection(t *testing.T) {
	section := Section("controls",
		From(entityops.ControlFields.RefCode).Required(),
		From(entityops.ControlFields.Tags),
		Extra("soc2_mapping", docextract.TypeArray),
		Extra("note", docextract.TypeString).Describe("free text").Enum("a", "b"),
	)

	assert.Check(t, cmp.DeepEqual(section.Required, []string{"controls"}))

	list := section.Properties["controls"]
	assert.Assert(t, list != nil)
	assert.Check(t, cmp.Equal(list.Type, docextract.TypeArray))

	item := list.Items
	assert.Check(t, cmp.DeepEqual(item.Required, []string{"ref_code"}))
	assert.Check(t, cmp.Equal(item.Properties["ref_code"].Type, docextract.TypeString))
	assert.Check(t, cmp.Equal(item.Properties["tags"].Type, docextract.TypeArray))
	assert.Check(t, cmp.Equal(item.Properties["tags"].Items.Type, docextract.TypeString))
	assert.Check(t, cmp.Equal(item.Properties["soc2_mapping"].Items.Type, docextract.TypeString))
	assert.Check(t, cmp.Equal(item.Properties["note"].Description, "free text"))
	assert.Check(t, cmp.DeepEqual(item.Properties["note"].Enum, []string{"a", "b"}))
}

func TestFromBoolField(t *testing.T) {
	field := From(entityops.EntityFields.ApprovedForUse)

	assert.Check(t, cmp.Equal(field.Name, "approved_for_use"))
	assert.Check(t, cmp.Equal(field.Schema.Type, docextract.TypeBoolean))
}

func TestDescribeDoesNotShareSchema(t *testing.T) {
	base := Extra("note", docextract.TypeString)
	described := base.Describe("changed")

	assert.Check(t, cmp.Equal(base.Schema.Description, ""))
	assert.Check(t, cmp.Equal(described.Schema.Description, "changed"))
}
