package scim

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable reference for the SCIM integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	// SCIMAuthWebhook is the stable identity handle for the SCIM authentication webhook
	SCIMAuthWebhook = types.NewWebhookRef("scim.auth")
	// userInput is the installation user input layout
	userInput = types.UserInputRefOf[UserInput]()
)

// UserInput captures optional user-provided configuration for the SCIM integration
type UserInput struct {
	// Name is the human-readable label for this SCIM directory (e.g. "Okta Production")
	Name string `json:"name,omitempty" jsonschema:"required,title=Directory Name,description=Human-readable label for this SCIM directory."`
	// PrimaryDirectory marks this installation as the authoritative directory source
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}
