package workflows

import (
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
)

// DeriveTriggerPrefilter extracts normalized trigger operations and fields/edges for query prefiltering.
// If any trigger omits fields/edges, field prefiltering is disabled
func DeriveTriggerPrefilter(doc models.WorkflowDefinitionDocument) ([]string, []string) {
	operations := make([]string, 0, len(doc.Triggers))
	fields := make([]string, 0)
	hasAnyFieldTrigger := false

	for _, trigger := range doc.Triggers {
		if trigger.Operation != "" {
			operations = append(operations, strings.ToUpper(trigger.Operation))
		}

		filteredFields := lo.Filter(trigger.Fields, func(f string, _ int) bool { return f != "" })
		filteredEdges := lo.Filter(trigger.Edges, func(e string, _ int) bool { return e != "" })

		if len(filteredFields) == 0 && len(filteredEdges) == 0 {
			hasAnyFieldTrigger = true
			continue
		}

		fields = append(fields, filteredFields...)
		fields = append(fields, filteredEdges...)
	}

	operations = lo.Uniq(operations)
	sort.Strings(operations)

	if hasAnyFieldTrigger {
		return operations, nil
	}

	fields = lo.Uniq(fields)
	sort.Strings(fields)

	return operations, fields
}
