package scim

import "github.com/theopenlane/core/internal/integrations/types"

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

const Slug = "scim_directory_sync"
