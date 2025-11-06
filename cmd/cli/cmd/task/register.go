//go:build cli

package task

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
		"openlaneclient.CreateTaskInput": reflect.TypeOf(openlaneclient.CreateTaskInput{}),
		"openlaneclient.UpdateTaskInput": reflect.TypeOf(openlaneclient.UpdateTaskInput{}),
		"openlaneclient.TaskWhereInput":  reflect.TypeOf(openlaneclient.TaskWhereInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks: map[string]speccli.CreateHookFactory{
			"taskCSV": func(_ *speccli.CreateSpec) speccli.CreatePreHook { return taskCSVHook },
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register task command: %v", err))
	}
}
