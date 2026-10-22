//go:build test

package microsoft

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

const (
	testManagingTenantID = "00000000-0000-0000-0000-00000000aaaa"
	testPrincipalID      = "00000000-0000-0000-0000-00000000bbbb"
)

var testOffer = LighthouseOffer{
	NamePrefix:           "Openlane Test",
	Description:          "test offer",
	PrincipalDisplayName: "Openlane",
	RoleDefinitionIDs:    []string{SecurityReaderRoleID, RegistrationAssignmentDeleteRoleID},
}

func TestLighthouseTemplate(t *testing.T) {
	cfg := LighthouseConfig{ManagingTenantID: testManagingTenantID, PrincipalID: testPrincipalID}

	out, err := cfg.Template(testOffer)
	assert.NilError(t, err)

	var template struct {
		Schema     string                  `json:"$schema"`
		Parameters map[string]armParameter `json:"parameters"`
		Variables  map[string]string       `json:"variables"`
		Resources  []struct {
			Type       string          `json:"type"`
			Properties json.RawMessage `json:"properties"`
		} `json:"resources"`
	}
	assert.NilError(t, json.Unmarshal([]byte(out), &template))

	assert.Check(t, template.Schema == lighthouseTemplateSchema)
	assert.Check(t, template.Parameters[connectionCodeParameter].Type == "string")
	assert.Check(t, template.Variables["mspOfferName"] == "[concat('Openlane Test ', parameters('connectionCode'))]")
	assert.Assert(t, len(template.Resources) == 2)
	assert.Check(t, template.Resources[0].Type == registrationDefinitionType)
	assert.Check(t, template.Resources[1].Type == registrationAssignmentType)

	var definition registrationDefinitionProperties
	assert.NilError(t, json.Unmarshal(template.Resources[0].Properties, &definition))
	assert.Check(t, definition.ManagedByTenantID == testManagingTenantID)
	assert.Assert(t, len(definition.Authorizations) == 2)

	for i, auth := range definition.Authorizations {
		assert.Check(t, auth.PrincipalID == testPrincipalID)
		assert.Check(t, auth.RoleDefinitionID == testOffer.RoleDefinitionIDs[i])
	}
}

func TestLighthouseOfferName(t *testing.T) {
	assert.Check(t, testOffer.OfferName("ABC123") == "Openlane Test ABC123")
}

func TestListAndDeleteDelegations(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetDelegations("sub-1",
		MockDelegation{OfferName: "Openlane Test ABC", ManagedByTenantID: testManagingTenantID},
		MockDelegation{OfferName: "Other Provider", ManagedByTenantID: "00000000-0000-0000-0000-00000000cccc"},
		MockDelegation{OfferName: "Openlane Test DEF", ManagedByTenantID: "00000000-0000-0000-0000-00000000AAAA"},
	)

	cfg := mock.AppConfig()

	cred, err := NewTokenCredential(cfg, testManagingTenantID)
	assert.NilError(t, err)

	client, err := NewARMClient(cfg, cred)
	assert.NilError(t, err)

	delegations, err := client.ListDelegations(context.Background(), "sub-1", testManagingTenantID)
	assert.NilError(t, err)
	assert.Assert(t, len(delegations) == 2)
	assert.Check(t, delegations[0].OfferName == "Openlane Test ABC" && delegations[1].OfferName == "Openlane Test DEF")

	match, ok := FindDelegation(delegations, "openlane test def")
	assert.Check(t, ok)
	assert.Check(t, match.AssignmentID == "/subscriptions/sub-1/providers/Microsoft.ManagedServices/registrationAssignments/assignment-2")

	_, ok = FindDelegation(delegations, "Other Provider")
	assert.Check(t, !ok)

	assert.NilError(t, client.DeleteDelegation(context.Background(), match.AssignmentID))
	assert.Check(t, len(mock.DeletedDelegations()) == 1 && mock.DeletedDelegations()[0] == match.AssignmentID)

	delegations, err = client.ListDelegations(context.Background(), "sub-1", testManagingTenantID)
	assert.NilError(t, err)
	assert.Check(t, len(delegations) == 1 && delegations[0].OfferName == "Openlane Test ABC")

	delegations, err = client.ListDelegations(context.Background(), "sub-2", testManagingTenantID)
	assert.NilError(t, err)
	assert.Check(t, len(delegations) == 0)
}
