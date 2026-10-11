package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// githubTestCredential is the credential type stored by the GitHub disconnect test definition
type githubTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

// githubTestClient is the client the GitHub disconnect test connection provides
type githubTestClient struct{}

// githubTestInstallation is the installation metadata layout of the GitHub disconnect test definition
type githubTestInstallation struct{}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.ConnectionOf[githubTestCredential]().Connection().Credential.Name
)

// withDefinitionRuntime returns a restore function that resets IntegrationsConfig
func (suite *HandlerTestSuite) withDefinitionRuntime(_ *testing.T, _ []registry.Builder) func() {
	originalConfig := suite.h.IntegrationsConfig

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// withGitHubAppIntegrationRuntime sets the handler's GitHubApp config for the test and returns a restore function that resets it
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()

	originalConfig := suite.h.IntegrationsConfig
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// githubTestDefinitionBuilder returns a minimal test definition used for disconnect tests
func githubTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		connection := types.ConnectionOf[githubTestCredential]().
			Name("GitHub Test Connection").
			Description("Test connection used for handler disconnect flows.").
			Provides(func(context.Context, types.ConnectionRequest[githubTestCredential]) (*githubTestClient, error) {
				return &githubTestClient{}, nil
			}).
			Verified(func(context.Context, types.ConnectionRequest[githubTestCredential], *githubTestClient) (githubTestInstallation, error) {
				return githubTestInstallation{}, nil
			}).
			Disconnects("Remove the persisted GitHub test credential and disconnect this installation.", nil)

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "GitHub",
				Active:      true,
				Visible:     true,
			},
			Installation: types.InstallationOf[githubTestInstallation]().Registration(),
			Connections:  []types.Connector{connection},
		}, nil
	})
}

const (
	webhookTestDefinitionID = "def_01K0TESTWBHK0000000000001"
	webhookTestPath         = "/v1/integrations/webhooks/:endpointID"
	webhookTestSecret       = "webhook-test-secret"
)

// WebhookTestHealthCheck is the config type for the webhook test health check operation
type WebhookTestHealthCheck struct{}

// webhookTestAlertEnvelope is the payload type for the test webhook events
type webhookTestAlertEnvelope struct{}

// webhookTestCredential is the credential type stored by the webhook test definition
type webhookTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

// webhookTestClient is the client the webhook test connection provides
type webhookTestClient struct{}

// webhookTestInstallation is the installation metadata layout of the webhook test definition
type webhookTestInstallation struct{}

var (
	webhookTestCredentialRef    = types.ConnectionOf[webhookTestCredential]().Connection().Credential.Name
	webhookHealthCheckOperation = types.OperationPayloadOf[WebhookTestHealthCheck]().Policy(types.ExecutionPolicy{Inline: true})
	webhookAlertCreatedEvent    = types.NewWebhookEventRef[webhookTestAlertEnvelope]("alert.created")
)

func webhookTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		connection := types.ConnectionOf[webhookTestCredential]().
			Name("Webhook Test Connection").
			Provides(func(context.Context, types.ConnectionRequest[webhookTestCredential]) (*webhookTestClient, error) {
				return &webhookTestClient{}, nil
			}).
			Verified(func(context.Context, types.ConnectionRequest[webhookTestCredential], *webhookTestClient) (webhookTestInstallation, error) {
				return webhookTestInstallation{}, nil
			})

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Webhook Test",
				Active:      true,
				Visible:     true,
			},
			Installation: types.InstallationOf[webhookTestInstallation]().Registration(),
			Connections:  []types.Connector{connection},
			Webhooks: []types.WebhookRegistration{
				{
					Name: "inbound.events",
					Event: func(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
						var envelope struct {
							Event      string `json:"event"`
							DeliveryID string `json:"delivery_id"`
						}
						if err := json.Unmarshal(req.Payload, &envelope); err != nil {
							return types.WebhookReceivedEvent{}, err
						}

						if envelope.Event == "" {
							return types.WebhookReceivedEvent{}, nil
						}

						return types.WebhookReceivedEvent{
							Name:       envelope.Event,
							DeliveryID: envelope.DeliveryID,
							Payload:    req.Payload,
						}, nil
					},
					Events: []types.WebhookEventRegistration{
						webhookAlertCreatedEvent.Registration(types.WebhookEventRegistration{
							Handle: func(context.Context, types.WebhookHandleRequest) error {
								return nil
							},
						}),
					},
				},
			},
			Operations: []types.OperationRegistration{
				webhookHealthCheckOperation.Description("Health check").HandlesRequest(func(context.Context, types.OperationRequest, WebhookTestHealthCheck) (json.RawMessage, error) {
					return json.RawMessage(`{"ok":true}`), nil
				}).Registration(),
			},
		}, nil
	})
}

const (
	operationTestDefinitionID       = "def_01K0TESTOPS00000000000001"
	operationTestInlineDefinitionID = "def_01K0TESTOPS00000000000002"
	operationTestPath               = "/v1/integrations/:definitionID/operations"
)

// OperationTestHealthCheck is the config type for the test health check operation
type OperationTestHealthCheck struct{}

// OperationTestRepoSync is the config type for the test repository sync operation
type OperationTestRepoSync struct{}

// OperationTestValidated is the config type for the test validated operation
type OperationTestValidated struct {
	types.OperationSettings
	// Target is the required target field
	Target string `json:"target" jsonschema:"required"`
}

// operationTestCredential is the credential type stored by the operation test definition
type operationTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

// operationTestClient is the client the operation test connection provides
type operationTestClient struct{}

// operationTestInstallation is the installation metadata layout of the operation test definition
type operationTestInstallation struct{}

var (
	operationTestCredentialRef = types.ConnectionOf[operationTestCredential]().Connection().Credential.Name
	opTestHealthCheckOperation = types.OperationPayloadOf[OperationTestHealthCheck]().Policy(types.ExecutionPolicy{Inline: true})
	opTestRepoSyncOperation    = types.OperationPayloadOf[OperationTestRepoSync]()
	opTestValidatedOperation   = types.OperationRefOf[OperationTestValidated]().Policy(types.ExecutionPolicy{Inline: true})
)

func operationTestDefinitionBuilder(definitionID string, inlineNonHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		connection := types.ConnectionOf[operationTestCredential]().
			Name("Op Test Connection").
			Provides(func(context.Context, types.ConnectionRequest[operationTestCredential]) (*operationTestClient, error) {
				return &operationTestClient{}, nil
			}).
			Verified(func(context.Context, types.ConnectionRequest[operationTestCredential], *operationTestClient) (operationTestInstallation, error) {
				return operationTestInstallation{}, nil
			})

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Operation Test",
				Active:      true,
				Visible:     true,
			},
			Installation: types.InstallationOf[operationTestInstallation]().Registration(),
			Connections:  []types.Connector{connection},
			Operations: []types.OperationRegistration{
				opTestHealthCheckOperation.Description("Validate the test credential").HandlesRequest(func(context.Context, types.OperationRequest, OperationTestHealthCheck) (json.RawMessage, error) {
					return json.RawMessage(`{"ok":true}`), nil
				}).Registration(),
				opTestRepoSyncOperation.Policy(types.ExecutionPolicy{Inline: inlineNonHealth}).Description("Sync repositories").HandlesRequest(func(context.Context, types.OperationRequest, OperationTestRepoSync) (json.RawMessage, error) {
					return json.RawMessage(`{"synced":true}`), nil
				}).Registration(),
				opTestValidatedOperation.Description("Operation with config schema").HandlesRequest(func(context.Context, types.OperationRequest, OperationTestValidated) (json.RawMessage, error) {
					return json.RawMessage(`{"validated":true}`), nil
				}).Registration(),
			},
		}, nil
	})
}

const (
	configTestProviderID           = "def_01K0TESTCFG00000000000001"
	configTestFailHealthProviderID = "def_01K0TESTCFG00000000000002"
	configTestVendorProviderID     = "def_01K0TESTCFG00000000000003"
	configTestVendorFamily         = "configtestvendor"
	configTestProbeProviderID      = "def_01K0TESTCFG00000000000004"
	configTestProbeFailProviderID  = "def_01K0TESTCFG00000000000005"
)

var errConfigTestProbeFailed = errors.New("probe prerequisites missing")

// configTestCredential is the credential type stored by the config test definition
type configTestCredential struct {
	// ProjectID is the required project identifier
	ProjectID string `json:"projectId" jsonschema:"required"`
	// ServiceAccountEmail is the required service account email
	ServiceAccountEmail string `json:"serviceAccountEmail" jsonschema:"required"`
}

// configTestUserInput is the installation-scoped user input layout stored by the config test definitions
type configTestUserInput struct {
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty"`
}

// configTestClient is the client the config test connection provides
type configTestClient struct{}

// configTestInstallation is the installation metadata layout of the config test definition
type configTestInstallation struct {
	// ProjectID is the verified project identifier
	ProjectID string `json:"projectId,omitempty"`
}

// ConfigTestProbe is the payload type of the config test operation that carries a health probe
type ConfigTestProbe struct{}

// ConfigTestUnprobed is the payload type of the config test operation without a health probe
type ConfigTestUnprobed struct{}

var (
	configTestCredentialRef  = types.ConnectionOf[configTestCredential]().Connection().Credential.Name
	configTestUserInputRef   = types.UserInputRefOf[configTestUserInput]()
	configTestProbeOp        = types.OperationPayloadOf[ConfigTestProbe]()
	configTestUnprobedOp     = types.OperationPayloadOf[ConfigTestUnprobed]()
	configTestProbeOperation = configTestProbeOp.Name()
)

func configTestProbeDefinitionBuilder(definitionID string, probeErr error) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def, err := configTestDefinitionBuilder(definitionID, false)()
		if err != nil {
			return types.Definition{}, err
		}

		def.Operations = append(def.Operations,
			configTestProbeOp.
				HealthCheck(func(context.Context, types.OperationRequest, *configTestClient) error {
					return probeErr
				}).
				HandlesRequest(func(context.Context, types.OperationRequest, ConfigTestProbe) (json.RawMessage, error) {
					return json.RawMessage(`{}`), nil
				}).
				Registration(),
			configTestUnprobedOp.
				HandlesRequest(func(context.Context, types.OperationRequest, ConfigTestUnprobed) (json.RawMessage, error) {
					return json.RawMessage(`{}`), nil
				}).
				Registration(),
		)

		return def, nil
	})
}

func configTestFamilyDefinitionBuilder(definitionID, family string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def, err := configTestDefinitionBuilder(definitionID, false)()
		if err != nil {
			return types.Definition{}, err
		}

		def.Family = family

		return def, nil
	})
}

func configTestDefinitionBuilder(definitionID string, failHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		connection := types.ConnectionOf[configTestCredential]().
			Name("Config Test Connection").
			Description("Connect the config test definition using the configured credential payload.").
			Disconnects("Remove the persisted config test credential and disconnect this installation.", nil).
			Provides(func(context.Context, types.ConnectionRequest[configTestCredential]) (*configTestClient, error) {
				return &configTestClient{}, nil
			}).
			Verified(func(_ context.Context, request types.ConnectionRequest[configTestCredential], _ *configTestClient) (configTestInstallation, error) {
				if failHealth {
					return configTestInstallation{}, errors.New("health failed")
				}

				return configTestInstallation{ProjectID: request.Credential.ProjectID}, nil
			})

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Config Test",
				Active:      true,
				Visible:     true,
			},
			UserInput:    configTestUserInputRef.Registration(),
			Installation: types.InstallationOf[configTestInstallation]().Registration(),
			Connections:  []types.Connector{connection},
		}, nil
	})
}

func userInputOnlyTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "User Input Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: configTestUserInputRef.Registration(),
		}, nil
	})
}
