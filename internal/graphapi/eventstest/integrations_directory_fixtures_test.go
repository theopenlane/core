//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"maps"
	"sync/atomic"
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// directoryProfileBaseLastLogin is the profile lastLoginTime carried by every directorySnapshot fixture record before a churn permutation advances it
const directoryProfileBaseLastLogin = "2024-01-01T00:00:00Z"

// directoryProfileChurnedLastLogin is the advanced lastLoginTime a churn permutation writes into a fixture record's profile
const directoryProfileChurnedLastLogin = "2024-06-01T00:00:00Z"

// directoryAccountRecord is one DirectoryAccount create-input payload used by a directorySnapshot fixture
type directoryAccountRecord struct {
	// ExternalID is the stable identifier from the directory system, and the account's lookup key
	ExternalID string `json:"external_id"`
	// CanonicalEmail is the account's primary email address
	CanonicalEmail string `json:"canonical_email,omitempty"`
	// DisplayName is the provider supplied display name, a material (non-Volatile) field
	DisplayName string `json:"display_name,omitempty"`
	// Profile is the flattened attribute bag, and is Volatile
	Profile map[string]any `json:"profile,omitempty"`
}

// directoryGroupRecord is one DirectoryGroup create-input payload used by a directorySnapshot fixture
type directoryGroupRecord struct {
	// ExternalID is the stable identifier from the directory system, and the group's lookup key
	ExternalID string `json:"external_id"`
	// DisplayName is the directory supplied display name, a material (non-Volatile) field
	DisplayName string `json:"display_name,omitempty"`
	// Profile is the flattened attribute bag, and is Volatile
	Profile map[string]any `json:"profile,omitempty"`
}

// directoryMembershipRecord is one DirectoryMembership create-input payload used by a directorySnapshot fixture
type directoryMembershipRecord struct {
	// DirectoryAccountID references the member account by its external id
	DirectoryAccountID string `json:"directory_account_id"`
	// DirectoryGroupID references the group by its external id
	DirectoryGroupID string `json:"directory_group_id"`
	// Role is the membership role reported by the provider, a material (non-Volatile) field
	Role string `json:"role,omitempty"`
	// Metadata is the raw provider metadata, a Volatile field
	Metadata map[string]any `json:"metadata,omitempty"`
}

// directorySnapshot is one deep-copyable directory provider snapshot — accounts, groups, and memberships
type directorySnapshot struct {
	// Prefix scopes every external id and email in this snapshot
	Prefix string
	// Accounts are the snapshot's DirectoryAccount records
	Accounts []directoryAccountRecord
	// Groups are the snapshot's DirectoryGroup records
	Groups []directoryGroupRecord
	// Memberships are the snapshot's DirectoryMembership records, referencing Accounts and Groups by external id
	Memberships []directoryMembershipRecord
}

// newDirectoryAccountRecord builds one account record with a profile carrying stable keys plus the churny lastLoginTime key
func newDirectoryAccountRecord(externalID, displayName string) directoryAccountRecord {
	return directoryAccountRecord{
		ExternalID:     externalID,
		CanonicalEmail: externalID + "@example.com",
		DisplayName:    displayName,
		Profile: map[string]any{
			"id":            externalID,
			"displayName":   displayName,
			"department":    "Engineering",
			"lastLoginTime": directoryProfileBaseLastLogin,
		},
	}
}

// newDirectoryGroupRecord builds one group record with a profile carrying stable keys plus the churny lastLoginTime key
func newDirectoryGroupRecord(externalID, displayName string) directoryGroupRecord {
	return directoryGroupRecord{
		ExternalID:  externalID,
		DisplayName: displayName,
		Profile: map[string]any{
			"id":            externalID,
			"displayName":   displayName,
			"lastLoginTime": directoryProfileBaseLastLogin,
		},
	}
}

// newDirectoryMembershipRecord builds one membership record linking account to the group identified by groupExternalID
func newDirectoryMembershipRecord(account directoryAccountRecord, groupExternalID, role string) directoryMembershipRecord {
	return directoryMembershipRecord{
		DirectoryAccountID: account.ExternalID,
		DirectoryGroupID:   groupExternalID,
		Role:               role,
		Metadata: map[string]any{
			"member_profile": maps.Clone(account.Profile),
			"role":           role,
		},
	}
}

// newDirectorySnapshot builds the base directory snapshot every scenario test in this package derives its permutations from
func newDirectorySnapshot(prefix string) directorySnapshot {
	accounts := []directoryAccountRecord{
		newDirectoryAccountRecord(prefix+"-acct-1", "Alice Example"),
		newDirectoryAccountRecord(prefix+"-acct-2", "Bob Example"),
		newDirectoryAccountRecord(prefix+"-acct-3", "Carol Example"),
		newDirectoryAccountRecord(prefix+"-acct-4", "Dave Example"),
	}

	groups := []directoryGroupRecord{
		newDirectoryGroupRecord(prefix+"-grp-1", "Engineering"),
		newDirectoryGroupRecord(prefix+"-grp-2", "Security"),
	}

	return directorySnapshot{
		Prefix:   prefix,
		Accounts: accounts,
		Groups:   groups,
		Memberships: []directoryMembershipRecord{
			newDirectoryMembershipRecord(accounts[0], groups[0].ExternalID, enums.DirectoryMembershipRoleMember.String()),
			newDirectoryMembershipRecord(accounts[1], groups[0].ExternalID, enums.DirectoryMembershipRoleMember.String()),
			newDirectoryMembershipRecord(accounts[2], groups[0].ExternalID, enums.DirectoryMembershipRoleMaintainer.String()),
			newDirectoryMembershipRecord(accounts[3], groups[1].ExternalID, enums.DirectoryMembershipRoleMember.String()),
			newDirectoryMembershipRecord(accounts[0], groups[1].ExternalID, enums.DirectoryMembershipRoleOwner.String()),
		},
	}
}

// clone copies s's account, group, and membership slices into new backing arrays so a permutation method can edit them without mutating s
func (s directorySnapshot) clone() directorySnapshot {
	return directorySnapshot{
		Prefix:      s.Prefix,
		Accounts:    append([]directoryAccountRecord{}, s.Accounts...),
		Groups:      append([]directoryGroupRecord{}, s.Groups...),
		Memberships: append([]directoryMembershipRecord{}, s.Memberships...),
	}
}

// account returns the record in s whose external id matches externalID
func (s directorySnapshot) account(externalID string) directoryAccountRecord {
	record, _ := lo.Find(s.Accounts, func(a directoryAccountRecord) bool { return a.ExternalID == externalID })

	return record
}

// group returns the record in s whose external id matches externalID
func (s directorySnapshot) group(externalID string) directoryGroupRecord {
	record, _ := lo.Find(s.Groups, func(g directoryGroupRecord) bool { return g.ExternalID == externalID })

	return record
}

// identical returns a snapshot whose accounts, groups, and memberships are unchanged from s
func (s directorySnapshot) identical() directorySnapshot {
	return s.clone()
}

// withAccountProfileChurn returns a snapshot where the account matching externalID has advanced its profile's lastLoginTime and nothing else
func (s directorySnapshot) withAccountProfileChurn(externalID string) directorySnapshot {
	out := s.clone()

	for i, account := range out.Accounts {
		if account.ExternalID != externalID {
			continue
		}

		profile := maps.Clone(account.Profile)
		profile["lastLoginTime"] = directoryProfileChurnedLastLogin
		out.Accounts[i].Profile = profile
	}

	return out
}

// withAccountMaterialChange returns a snapshot where the account matching externalID has a new display_name
func (s directorySnapshot) withAccountMaterialChange(externalID string) directorySnapshot {
	out := s.clone()

	for i, account := range out.Accounts {
		if account.ExternalID != externalID {
			continue
		}

		out.Accounts[i].DisplayName = account.DisplayName + " Updated"
	}

	return out
}

// withGroupProfileChurn returns a snapshot where the group matching externalID has advanced its profile's lastLoginTime and nothing else
func (s directorySnapshot) withGroupProfileChurn(externalID string) directorySnapshot {
	out := s.clone()

	for i, group := range out.Groups {
		if group.ExternalID != externalID {
			continue
		}

		profile := maps.Clone(group.Profile)
		profile["lastLoginTime"] = directoryProfileChurnedLastLogin
		out.Groups[i].Profile = profile
	}

	return out
}

// withGroupMaterialChange returns a snapshot where the group matching externalID has a new display_name
func (s directorySnapshot) withGroupMaterialChange(externalID string) directorySnapshot {
	out := s.clone()

	for i, group := range out.Groups {
		if group.ExternalID != externalID {
			continue
		}

		out.Groups[i].DisplayName = group.DisplayName + " Updated"
	}

	return out
}

// withMembershipMetadataChurn returns a snapshot where the matching membership has advanced the lastLoginTime embedded in its metadata and nothing else
func (s directorySnapshot) withMembershipMetadataChurn(accountExternalID, groupExternalID string) directorySnapshot {
	out := s.clone()

	for i, membership := range out.Memberships {
		if membership.DirectoryAccountID != accountExternalID || membership.DirectoryGroupID != groupExternalID {
			continue
		}

		metadata := maps.Clone(membership.Metadata)

		if profile, ok := metadata["member_profile"].(map[string]any); ok {
			churned := maps.Clone(profile)
			churned["lastLoginTime"] = directoryProfileChurnedLastLogin
			metadata["member_profile"] = churned
		}

		out.Memberships[i].Metadata = metadata
	}

	return out
}

// withoutMembership returns a snapshot with the membership between accountExternalID and groupExternalID removed
func (s directorySnapshot) withoutMembership(accountExternalID, groupExternalID string) directorySnapshot {
	out := s.clone()

	out.Memberships = lo.Reject(out.Memberships, func(m directoryMembershipRecord, _ int) bool {
		return m.DirectoryAccountID == accountExternalID && m.DirectoryGroupID == groupExternalID
	})

	return out
}

// withoutAccount returns a snapshot with the account matching externalID, and every membership referencing it, removed
func (s directorySnapshot) withoutAccount(externalID string) directorySnapshot {
	out := s.clone()

	out.Accounts = lo.Reject(out.Accounts, func(a directoryAccountRecord, _ int) bool { return a.ExternalID == externalID })
	out.Memberships = lo.Reject(out.Memberships, func(m directoryMembershipRecord, _ int) bool { return m.DirectoryAccountID == externalID })

	return out
}

// withNewAccount returns a snapshot with one additional account, identified by externalID, and one membership linking it to the first group
func (s directorySnapshot) withNewAccount(externalID string) directorySnapshot {
	out := s.clone()

	account := newDirectoryAccountRecord(externalID, "New Account")
	out.Accounts = append(out.Accounts, account)
	out.Memberships = append(out.Memberships, newDirectoryMembershipRecord(account, out.Groups[0].ExternalID, enums.DirectoryMembershipRoleMember.String()))

	return out
}

// marshalDirectoryEnvelopes marshals each fixture record to its JSON create-input payload and wraps it as a mapping envelope
func marshalDirectoryEnvelopes[T any](records []T) []integrationtypes.MappingEnvelope {
	return lo.Map(records, func(record T, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: lo.Must(json.Marshal(record))}
	})
}

// payloadSets converts s into the three IngestPayloadSets a directory sync run submits together — account, group, and membership
func (s directorySnapshot) payloadSets(complete bool) []integrationtypes.IngestPayloadSet {
	return []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaDirectoryAccount.Name, Envelopes: marshalDirectoryEnvelopes(s.Accounts), SnapshotComplete: complete},
		{Schema: entityops.SchemaDirectoryGroup.Name, Envelopes: marshalDirectoryEnvelopes(s.Groups), SnapshotComplete: complete},
		{Schema: entityops.SchemaDirectoryMembership.Name, Envelopes: marshalDirectoryEnvelopes(s.Memberships), SnapshotComplete: complete},
	}
}

// ingestDirectorySnapshotFixture runs snap's account, group, and membership payload sets through the synchronous ingest path in one ProcessPayloadSets call
func ingestDirectorySnapshotFixture(ctx context.Context, t *testing.T, integration *ent.Integration, snap directorySnapshot, complete bool, options ...operations.IngestOptions) operations.IngestResult {
	t.Helper()

	def := directorySyncTestDefinition(integration.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
	}, directorySyncTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, snap.payloadSets(complete), lo.FirstOrEmpty(options))
	th.RequireNoError(t, err)

	return result
}

// directoryEventCounterSet tracks create/update mutation event counts per directory schema
type directoryEventCounterSet struct {
	// Runtime is the gala runtime the counting listeners are registered on, for waitForGala
	Runtime *gala.Gala
	// AccountCreates counts DirectoryAccount create mutation events
	AccountCreates *atomic.Int64
	// AccountUpdates counts DirectoryAccount update mutation events
	AccountUpdates *atomic.Int64
	// GroupCreates counts DirectoryGroup create mutation events
	GroupCreates *atomic.Int64
	// GroupUpdates counts DirectoryGroup update mutation events
	GroupUpdates *atomic.Int64
	// MembershipCreates counts DirectoryMembership create mutation events
	MembershipCreates *atomic.Int64
	// MembershipUpdates counts DirectoryMembership update mutation events
	MembershipUpdates *atomic.Int64
}

// directoryEventCounters registers counting listeners for directory schema create/update mutation events on the shared test runtime
func directoryEventCounters(t *testing.T) (directoryEventCounterSet, func()) {
	t.Helper()

	counters := directoryEventCounterSet{
		Runtime:           suite.GalaRuntime,
		AccountCreates:    &atomic.Int64{},
		AccountUpdates:    &atomic.Int64{},
		GroupCreates:      &atomic.Int64{},
		GroupUpdates:      &atomic.Int64{},
		MembershipCreates: &atomic.Int64{},
		MembershipUpdates: &atomic.Int64{},
	}

	countingListener := func(schema *entityops.Schema, creates, updates *atomic.Int64) gala.Registration {
		return entityops.MutationListener{
			Schema:     schema,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, payload entityops.MutationPayload) error {
				if payload.Operation == entityops.OpCreate {
					creates.Add(1)
				} else {
					updates.Add(1)
				}

				return nil
			},
		}
	}

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		countingListener(entityops.SchemaDirectoryAccount, counters.AccountCreates, counters.AccountUpdates),
		countingListener(entityops.SchemaDirectoryGroup, counters.GroupCreates, counters.GroupUpdates),
		countingListener(entityops.SchemaDirectoryMembership, counters.MembershipCreates, counters.MembershipUpdates),
	})
	th.RequireNoError(t, err)

	return counters, setup.Teardown
}

// cleanupDirectoryPrefix registers a t.Cleanup that removes every directory row created by a prefixed snapshot fixture
func cleanupDirectoryPrefix(t *testing.T, ctx context.Context, integrationIDs []string, prefix string) {
	t.Helper()

	t.Cleanup(func() {
		memberships, err := suite.Client.DB.DirectoryMembership.Query().Where(directorymembership.IntegrationIDIn(integrationIDs...)).All(ctx)
		th.RequireNoError(t, err)

		if len(memberships) > 0 {
			(&th.Cleanup[*ent.DirectoryMembershipDeleteOne]{Client: suite.Client.DB.DirectoryMembership, IDs: lo.Map(memberships, func(m *ent.DirectoryMembership, _ int) string { return m.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		accounts, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalIDHasPrefix(prefix)).All(ctx)
		th.RequireNoError(t, err)

		if len(accounts) > 0 {
			(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: lo.Map(accounts, func(da *ent.DirectoryAccount, _ int) string { return da.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		groups, err := suite.Client.DB.DirectoryGroup.Query().Where(directorygroup.ExternalIDHasPrefix(prefix)).All(ctx)
		th.RequireNoError(t, err)

		if len(groups) > 0 {
			(&th.Cleanup[*ent.DirectoryGroupDeleteOne]{Client: suite.Client.DB.DirectoryGroup, IDs: lo.Map(groups, func(dg *ent.DirectoryGroup, _ int) string { return dg.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		for _, integrationID := range integrationIDs {
			(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integrationID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}
	})
}
