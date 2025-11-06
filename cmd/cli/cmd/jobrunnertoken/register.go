//go:build cli

package jobrunnertoken

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
		"openlaneclient.CreateJobRunnerTokenInput": reflect.TypeOf(openlaneclient.CreateJobRunnerTokenInput{}),
		"openlaneclient.UpdateJobRunnerTokenInput": reflect.TypeOf(openlaneclient.UpdateJobRunnerTokenInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register jobrunnertoken command: %v", err))
	}
}
