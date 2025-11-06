//go:build cli

package trustcenterwatermarkconfig

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
		"openlaneclient.CreateTrustCenterWatermarkConfigInput": reflect.TypeOf(openlaneclient.CreateTrustCenterWatermarkConfigInput{}),
		"openlaneclient.UpdateTrustCenterWatermarkConfigInput": reflect.TypeOf(openlaneclient.UpdateTrustCenterWatermarkConfigInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register trustcenterwatermarkconfig command: %v", err))
	}
}
