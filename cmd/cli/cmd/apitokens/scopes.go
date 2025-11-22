//go:build cli

package apitokens

import (
    "fmt"
    "sort"
    "strings"

    fgamodel "github.com/theopenlane/core/fga/model"
)

// scopeFlagConfig returns a description suffix listing available scopes.
func scopeFlagConfig() string {
    scopes, err := fgamodel.DefaultServiceScopes()
    if err != nil {
        panic(fmt.Sprintf("failed to load service scopes: %v", err))
    }

	desc := fmt.Sprintf(" (available: %s)", strings.Join(scopes, ", "))

    aliases := fgamodel.ScopeAliases()
	if len(aliases) > 0 {
		aliasPairs := make([]string, 0, len(aliases))

		for alias, relation := range aliases {
			aliasPairs = append(aliasPairs, fmt.Sprintf("%s->%s", alias, relation))
		}

		sort.Strings(aliasPairs)

		desc = fmt.Sprintf("%s; aliases: %s", desc, strings.Join(aliasPairs, ", "))
	}

	return desc
}
