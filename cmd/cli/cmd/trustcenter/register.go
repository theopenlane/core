//go:build cli

package trustcenter

import (
	"embed"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

//go:embed spec.json
var specFS embed.FS

var rootCmd *cobra.Command

func init() {
	typeMap := map[string]reflect.Type{
		"openlaneclient.CreateTrustCenterInput": reflect.TypeOf(openlaneclient.CreateTrustCenterInput{}),
		"openlaneclient.UpdateTrustCenterInput": reflect.TypeOf(openlaneclient.UpdateTrustCenterInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register trustcenter command: %v", err))
	}

	cmd := findTrustCenterCommand()
	if cmd == nil {
		panic("failed to locate registered trustcenter command")
	}

	rootCmd = cmd
	attachTrustCenterExtras(rootCmd)
}
