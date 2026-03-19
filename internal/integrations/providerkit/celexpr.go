package providerkit

import (
	"strconv"
	"strings"
)

// CelMapEntry holds one key-expression pair for building CEL object literal mapping expressions
type CelMapEntry struct {
	// Key is the target field name in the mapped output document
	Key string
	// Expr is the CEL expression that produces the value for Key
	Expr string
}

// CelMapExpr renders a slice of CelMapEntry values into a CEL object literal string
func CelMapExpr(entries []CelMapEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	var b strings.Builder

	b.WriteString("{\n")

	for i, entry := range entries {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(entry.Key))
		b.WriteString(": ")
		b.WriteString(entry.Expr)

		if i < len(entries)-1 {
			b.WriteString(",")
		}

		b.WriteString("\n")
	}

	b.WriteString("}")

	return b.String()
}
