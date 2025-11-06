//go:build cli

package trustcentersubprocessors

import (
	"embed"
	"fmt"
	"reflect"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

//go:embed spec.json
var specFS embed.FS

func init() {
	typeMap := map[string]reflect.Type{
		"openlaneclient.CreateTrustCenterSubprocessorInput": reflect.TypeOf(openlaneclient.CreateTrustCenterSubprocessorInput{}),
		"openlaneclient.UpdateTrustCenterSubprocessorInput": reflect.TypeOf(openlaneclient.UpdateTrustCenterSubprocessorInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register trust center subprocessor command: %v", err))
	}
}
