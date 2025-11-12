//go:build cli

package speccli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
)

type columnFormatter func(any) (any, error)

var columnFormatters = map[string]columnFormatter{
	"timeOrNever":   formatTimeOrNever,
	"edgeNames":     formatEdgeNames,
	"joinedStrings": formatJoinedStrings,
}

// formatTimeOrNever returns RFC3339 strings for times and "never" for empty values.
func formatTimeOrNever(input any) (any, error) {
	if input == nil {
		return "never", nil
	}

	switch v := input.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "never", nil
		}
		return v, nil
	case *string:
		if v == nil || strings.TrimSpace(*v) == "" {
			return "never", nil
		}
		return *v, nil
	case time.Time:
		return v.Format(time.RFC3339), nil
	case *time.Time:
		if v == nil {
			return "never", nil
		}
		return v.Format(time.RFC3339), nil
	default:
		return v, nil
	}
}

// formatEdgeNames extracts `node.name` from each edge and joins them.
func formatEdgeNames(input any) (any, error) {
	if input == nil {
		return "", nil
	}

	edges, ok := input.([]any)
	if !ok {
		return nil, errors.New("edgeNames formatter expects []any")
	}

	nodes, err := edgeNodesFromEdges(edges)
	if err != nil {
		return nil, err
	}

	names := lo.FilterMap(nodes, func(node map[string]any, _ int) (string, bool) {
		name, ok := node["name"].(string)
		return name, ok
	})

	return strings.Join(names, ", "), nil
}

// formatJoinedStrings stringifies slices and joins them with ", ".
func formatJoinedStrings(input any) (any, error) {
	switch v := input.(type) {
	case []any:
		strs := lo.Map(v, func(item any, _ int) string {
			return fmt.Sprint(item)
		})

		return strings.Join(strs, ", "), nil
	case []string:
		return strings.Join(v, ", "), nil
	default:
		return input, nil
	}
}
