//go:build cli

package apitokens

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	fgamodel "github.com/theopenlane/core/fga/model"
)

var scopesCmd = &cobra.Command{
	Use:   "scopes",
	Short: "list scopes available for an api token",
	Run: func(cmd *cobra.Command, args []string) {
		err := allScopes()
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(scopesCmd)
}

const (
	boldStart = "\033[1m"
	boldEnd   = "\033[0m"
)

// allScopes returns all available scopes
func allScopes() error {
	scopes, err := fgamodel.ScopeOptions()
	if err != nil {
		return fmt.Errorf("failed to load service scopes: %v", err)
	}

	objects := make([]string, 0, len(scopes))
	for obj := range scopes {
		objects = append(objects, obj)
	}
	sort.Strings(objects)

	desc := fmt.Sprintf(boldStart + "Available Scopes: \n\n" + boldEnd)
	for _, obj := range objects {
		for _, v := range scopes[obj] {
			desc += fmt.Sprintf(boldStart+"%s"+boldEnd+":%s ", obj, v)
		}
		desc += fmt.Sprintf("\n")
	}

	fmt.Println(desc)

	return nil
}
