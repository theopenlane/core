package oci

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Oracle Cloud Infrastructure definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Oracle Cloud Infrastructure",
			DisplayName:  "Oracle Cloud Infrastructure",
			Description:  "Collect Oracle Cloud Infrastructure Cloud Guard problems for security posture reporting.",
			Category:     "security-posture",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/oci",
			Tags:         []string{"findings"},
			Active:       false,
			Visible:      true,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				apiKey.
					Name("OCI API Key").
					Description("Configure Oracle Cloud Infrastructure access using an API signing key registered to a tenancy user.").
					Provides(buildClient).
					Verified(verify).
					Disconnects("Removes the stored API signing key from Openlane. If the key is no longer needed, delete it from the user's API keys in the OCI console.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[FindingsSync]().
					Ingests(runFindingsSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaFinding.Name}).
					Permissions("read cloud-guard-problems in tenancy").
					Description("Collect OCI Cloud Guard problems as findings").
					Registration(),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaFinding.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprFinding,
					},
				},
			},
		}, nil
	})
}
