package definition

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Assembler is the interface each definition package satisfies to compose a types.Definition.
// Each section of the definition is expressed as an independent method so callers can
// reason about each concern in isolation.
type Assembler interface {
	// Spec returns the catalog-visible metadata for the definition
	Spec() types.DefinitionSpec
	// OperatorConfig returns the operator-owned configuration registration, or nil
	OperatorConfig() *types.OperatorConfigRegistration
	// UserInput returns the installation-scoped user input registration, or nil
	UserInput() *types.UserInputRegistration
	// Credentials returns the credential registration, or nil
	Credentials() *types.CredentialRegistration
	// Auth returns the auth flow registration, or nil
	Auth() *types.AuthRegistration
	// Clients returns the client registrations for the definition
	Clients() []types.ClientRegistration
	// Operations returns the operation registrations for the definition
	Operations() []types.OperationRegistration
	// Mappings returns the default mapping registrations for the definition
	Mappings() []types.MappingRegistration
	// Webhooks returns the webhook registrations for the definition
	Webhooks() []types.WebhookRegistration
}

// FromAssembler builds a Builder from an Assembler, composing the types.Definition
// by calling each section method in turn
func FromAssembler(a Assembler) Builder {
	return BuilderFunc(func(context.Context) (types.Definition, error) {
		return types.Definition{
			Spec:           a.Spec(),
			OperatorConfig: a.OperatorConfig(),
			UserInput:      a.UserInput(),
			Credentials:    a.Credentials(),
			Auth:           a.Auth(),
			Clients:        a.Clients(),
			Operations:     a.Operations(),
			Mappings:       a.Mappings(),
			Webhooks:       a.Webhooks(),
		}, nil
	})
}
