package objectstore

import (
	"strconv"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// filterExprVendor admits only records that declare the vendor entity type under the vendor mapping variant
var filterExprVendor = `'entityTypeName' in payload && payload.entityTypeName == ` + strconv.Quote(variantVendor)

// linkExprVendorEntityType matches the organization's vendor entity type for the vendor mapping variant
var linkExprVendorEntityType = "target.name == " + strconv.Quote(variantVendor)

// mapExprEntity is the CEL mapping expression for stored vendor records mapped to Entity; the record's
// systemInternalID is both the upsert key and the retained system identifier
var mapExprEntity = providerkit.CelMapExpr(
	entityops.EntityFields.ExternalID.Expr(`payload.systemInternalID`),
	entityops.EntityFields.SystemInternalID.Expr(`payload.systemInternalID`),
	entityops.EntityFields.Name.Expr(`payload.name`),
	entityops.EntityFields.DisplayName.Expr(`'displayName' in payload ? payload.displayName : null`),
	entityops.EntityFields.Description.Expr(`'description' in payload ? payload.description : null`),
	entityops.EntityFields.Domains.Expr(`'domains' in payload ? payload.domains : null`),
	entityops.EntityFields.Tags.Expr(`'tags' in payload ? payload.tags : null`),
	entityops.EntityFields.LogoRemoteURL.Expr(`'logoRemoteURL' in payload ? payload.logoRemoteURL : null`),
)
