//go:build cli

package orgsetting

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
		"openlaneclient.UpdateOrganizationSettingInput": reflect.TypeOf(openlaneclient.UpdateOrganizationSettingInput{}),
		"openlaneclient.OrganizationSettingWhereInput":  reflect.TypeOf(openlaneclient.OrganizationSettingWhereInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"organizationSettingUpdate": updateOrganizationSettingHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register organization setting command: %v", err))
	}
}
