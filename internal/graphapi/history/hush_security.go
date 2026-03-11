package graphapihistory

import (
	"github.com/theopenlane/core/internal/ent/historygenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

func redactHushHistoryCredentialSet(value *historygenerated.HushHistory) {
	if value == nil {
		return
	}

	value.CredentialSet = integrationtypes.CredentialSet{}
}

func redactHushHistoryConnection(conn *historygenerated.HushHistoryConnection) {
	if conn == nil {
		return
	}

	for _, edge := range conn.Edges {
		if edge == nil {
			continue
		}

		redactHushHistoryCredentialSet(edge.Node)
	}
}
