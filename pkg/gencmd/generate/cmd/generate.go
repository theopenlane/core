package cmd

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/pkg/gencmd"
	"github.com/theopenlane/core/pkg/gencmd/generate/prompts"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate is the command to generate the stub files for a given cli cmd",
	Run: func(_ *cobra.Command, _ []string) {
		err := generateStubFiles()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("name", "n", "", "name of the command to generate")
	generateCmd.Flags().StringP("dir", "d", "cmd", "root directory location to generate the files")
	generateCmd.Flags().BoolP("read-only", "r", false, "only generate the read only commands, no create, update or delete commands")
	generateCmd.Flags().BoolP("interactive", "i", true, "interactive prompt, set to false to disable")
	generateCmd.Flags().Bool("spec", false, "generate spec-driven command files instead of legacy cobra stubs")
	generateCmd.Flags().BoolP("force", "f", false, "force overwrite of existing files")
}

func generateStubFiles() (err error) {
	interactive := Config.Bool("interactive")

	cmdName := Config.String("name")

	if interactive {
		cmdName, err = prompts.Name(cmdName)
		cobra.CheckErr(err)
	}

	dirName := Config.String("dir")
	readOnly := Config.Bool("read-only")
	spec := Config.Bool("spec")
	force := Config.Bool("force")

	return gencmd.Generate(cmdName, dirName, readOnly, spec, force)
}
