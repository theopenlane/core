package notifications

import (
	"strings"

	"entgo.io/ent"
)

func isUpdateOperation(operation string) bool {
	switch strings.TrimSpace(operation) {
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return true
	default:
		return false
	}
}
