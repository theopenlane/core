package objectstore

import (
	"strconv"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// entityMapping maps every record under the import prefix to Entity and links it to the vendor entity type
var entityMapping = types.MappingRegistration{
	Schema: entityops.SchemaEntity.Name,
	Spec: types.MappingOverride{
		FilterExpr: "true",
		MapExpr:    mapExprEntity,
		Links: []types.LinkRule{
			{
				TargetSchema: entityops.SchemaEntityType.Name,
				Expression:   "target.name == " + strconv.Quote(vendorEntityTypeName),
			},
		},
	},
}

// mapExprEntity maps a vendor record to Entity; the record's systemInternalID is both the upsert key and the retained system identifier
var mapExprEntity = providerkit.CelMapExpr(
	entityops.EntityFields.ExternalID.Expr(`payload.systemInternalID`),
	entityops.EntityFields.SystemInternalID.Expr(`payload.systemInternalID`),
	entityops.EntityFields.Name.Expr(`payload.name`),
	entityops.EntityFields.DisplayName.Expr(`'displayName' in payload ? payload.displayName : null`),
	entityops.EntityFields.Description.Expr(`'description' in payload ? payload.description : null`),
	entityops.EntityFields.Domains.Expr(`'domains' in payload ? payload.domains : null`),
	entityops.EntityFields.Aliases.Expr(`'aliases' in payload ? payload.aliases : null`),
	entityops.EntityFields.Tags.Expr(`'tags' in payload ? payload.tags : null`),
	entityops.EntityFields.LogoRemoteURL.Expr(`'logoRemoteURL' in payload ? payload.logoRemoteURL : null`),
	entityops.EntityFields.ExternallyVisible.Expr(`'externallyVisible' in payload ? payload.externallyVisible : false`),
)
