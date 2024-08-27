package version

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/internal/constants"
	"github.com/theopenlane/core/pkg/utils/cli/useragent"
)

// VersionCmd is the version command
var command = &cobra.Command{
	Use:   "version",
	Short: "print the CLI version",
	Long:  `The version command prints the version of the CLI`,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Println(constants.VerboseCLIVersion)
		cmd.Printf("User Agent: %s\n", useragent.GetUserAgent())
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
