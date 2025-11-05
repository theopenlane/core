//go:build cli

package speccli

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type columnFormatter func(any) (any, error)

var columnFormatters = map[string]columnFormatter{
	"timeOrNever":   formatTimeOrNever,
	"edgeNames":     formatEdgeNames,
	"joinedStrings": formatJoinedStrings,
}

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

func formatEdgeNames(input any) (any, error) {
	if input == nil {
		return "", nil
	}

	edges, ok := input.([]any)
	if !ok {
		return nil, errors.New("edgeNames formatter expects []any")
	}

	names := make([]string, 0, len(edges))
	for _, edge := range edges {
		if edgeMap, ok := edge.(map[string]any); ok {
			if node, ok := edgeMap["node"].(map[string]any); ok {
				if name, ok := node["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
	}

	return strings.Join(names, ", "), nil
}

func formatJoinedStrings(input any) (any, error) {
	switch v := input.(type) {
	case []any:
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprint(item)
		}
		return strings.Join(strs, ", "), nil
	case []string:
		return strings.Join(v, ", "), nil
	default:
		return input, nil
	}
}
