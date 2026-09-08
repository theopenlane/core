package providerkit

import (
	"strconv"
	"strings"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// CelMapExpr renders field-descriptor mapping entries into a CEL object literal string
func CelMapExpr(entries ...entityops.MappingEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	var b strings.Builder

	b.WriteString("{\n")

	for i, entry := range entries {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(entry.Key))
		b.WriteString(": dyn(")
		b.WriteString(entry.Expr)
		b.WriteString(")")

		if i < len(entries)-1 {
			b.WriteString(",")
		}

		b.WriteString("\n")
	}

	b.WriteString("}")

	return b.String()
}
