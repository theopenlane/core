//go:build cli

package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/internal/constants"
)

// VersionCmd is the version command
var command = &cobra.Command{
	Use:   "version",
	Short: "print the CLI version",
	Long:  `The version command prints the version of the CLI`,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Println(constants.VerboseCLIVersion)
		cmd.Printf("User Agent: %s\n", getUserAgent())
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

func getUserAgent() string {
	product := "openlane-cli"
	productVersion := constants.CLIVersion

	userAgent := fmt.Sprintf("%s/%s (%s) %s (%s)",
		product, productVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())

	return userAgent
}
