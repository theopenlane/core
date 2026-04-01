package graphapihistory

import (
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/historygenerated"
)

func redactHushHistoryCredentialSet(value *historygenerated.HushHistory) {
	if value == nil {
		return
	}

	value.CredentialSet = models.CredentialSet{}
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
