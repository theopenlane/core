package googleworkspace

import (
	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0GWKSP000000000000000001")
	WorkspaceClient        = types.NewClientRef[*admin.Service]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

const Slug = "google_workspace"
