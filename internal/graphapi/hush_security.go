package graphapi

import (
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

func validateCreateHushInput(input generated.CreateHushInput) error {
	if input.SecretValue != nil {
		return rout.InvalidField("secretValue")
	}

	if input.CredentialSet != nil {
		return rout.InvalidField("credentialSet")
	}

	return nil
}

func validateUpdateHushInput(input generated.UpdateHushInput) error {
	if input.CredentialSet != nil || input.ClearCredentialSet {
		return rout.InvalidField("credentialSet")
	}

	return nil
}

func redactHushCredentialSet(value *generated.Hush) {
	if value == nil {
		return
	}

	value.CredentialSet = integrationtypes.CredentialSet{}
}

func redactHushConnection(conn *generated.HushConnection) {
	if conn == nil {
		return
	}

	for _, edge := range conn.Edges {
		if edge == nil {
			continue
		}

		redactHushCredentialSet(edge.Node)
	}
}
