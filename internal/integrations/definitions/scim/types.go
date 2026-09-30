package scim

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable reference for the SCIM integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	// SCIMAuthWebhook is the stable identity handle for the SCIM authentication webhook
	SCIMAuthWebhook = types.NewWebhookRef("scim.auth")
	// userInput is the installation user input layout, replacing the flat v1 layout
	userInput = types.NewUserInputRef[UserInput]("scim").Replacing(types.NewUserInputRef[oldUserInput]("scim-v1"), func(old oldUserInput) UserInput {
		return UserInput{Name: old.Name, PrimaryDirectory: old.PrimaryDirectory, DirectorySync: DirectorySync{FilterExpr: old.FilterExpr}}
	})
)

// UserInput captures optional user-provided configuration for the SCIM integration
type UserInput struct {
	// Name is the human-readable label for this SCIM directory (e.g. "Okta Production")
	Name string `json:"name,omitempty" jsonschema:"required,title=Directory Name,description=Human-readable label for this SCIM directory."`
	// PrimaryDirectory marks this installation as the authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
	// DirectorySync configures the directory sync operation
	DirectorySync DirectorySync `json:"directorySync,omitempty" jsonschema:"title=Directory Sync"`
}

// oldUserInput is the flat v1 installation user input layout
type oldUserInput struct {
	// Name is the human-readable label for this SCIM directory (e.g. "Okta Production")
	Name string `json:"name,omitempty" jsonschema:"required,title=Directory Name,description=Human-readable label for this SCIM directory."`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
	// PrimaryDirectory marks this installation as the authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}
