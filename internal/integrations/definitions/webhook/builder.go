package webhook

import (
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the outbound webhook definition builder; devMode allows http and non-public destinations
func Builder(devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           DefinitionID.ID(),
			Family:       "Webhook",
			DisplayName:  "Webhook",
			Description:  "Send JSON payloads to an HTTPS endpoint, with the headers and signing it requires.",
			Category:     "messaging",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/webhook",
			Tags:         []string{"webhook", "notifications"},
			Active:       false,
			Visible:      false,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				webhookConnection.
					Name("Webhook Destination").
					Description("Configure the HTTPS endpoint, any headers it requires such as an API key, and optional signing of each delivery.").
					Validated(credentialValidator(devMode)).
					Provides(clientBuilder(devMode)).
					Verified(verify).
					Disconnects("Removes the stored destination, headers, and signing secret from Openlane. Rotate any key or secret on the destination if it is no longer needed.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationPayloadOf[DeliverOperation]().
					Handles(deliver).
					Description("POST a JSON payload to the configured destination").
					Registration(),
			},
		}, nil
	})
}
