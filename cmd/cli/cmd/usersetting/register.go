//go:build cli

package usersetting

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
		"openlaneclient.UpdateUserSettingInput": reflect.TypeOf(openlaneclient.UpdateUserSettingInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"userSettingUpdate": userSettingUpdateHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"userSettingGet": userSettingGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register usersetting command: %v", err))
	}
}
