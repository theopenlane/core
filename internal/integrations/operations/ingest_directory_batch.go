package operations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// directoryConfirmChunkSize bounds the id list of one bulk confirmation update
const directoryConfirmChunkSize = 500

// directoryScopeKey identifies one lookup scope by owner, directory instance, and integration
type directoryScopeKey struct {
	ownerID       string
	instanceID    string
	integrationID string
}

// directoryAccountScope indexes one scope's preloaded accounts by lookup key
type directoryAccountScope struct {
	byExternalID map[string][]*ent.DirectoryAccount
	byEmail      map[string][]*ent.DirectoryAccount
}

// directoryGroupScope indexes one scope's preloaded groups by lookup key
type directoryGroupScope struct {
	byExternalID map[string][]*ent.DirectoryGroup
	byEmail      map[string][]*ent.DirectoryGroup
}

// membershipKey identifies one active membership by resolved account and group id
type membershipKey struct {
	accountID string
	groupID   string
}

// directorySyncBatch holds one ingest pass's lookup caches and pending bulk confirmations
type directorySyncBatch struct {
	changed                int
	accountScopes          map[directoryScopeKey]*directoryAccountScope
	ownerAccountIDs        map[string]map[string]struct{}
	groupScopes            map[directoryScopeKey]*directoryGroupScope
	ownerGroupIDs          map[string]map[string]struct{}
	membershipIndexes      map[string]map[membershipKey][]*ent.DirectoryMembership
	seenAccountIDs         []string
	confirmedMembershipIDs []string
	relinkIDs              map[string][]string
	linkedIntegrationIDs   map[string]map[string]struct{}
}

var directorySyncBatchKey = contextx.NewKey[*directorySyncBatch]()

// newDirectorySyncBatch builds the empty batch state for one ingest pass
func newDirectorySyncBatch() *directorySyncBatch {
	return &directorySyncBatch{
		accountScopes:        map[directoryScopeKey]*directoryAccountScope{},
		ownerAccountIDs:      map[string]map[string]struct{}{},
		groupScopes:          map[directoryScopeKey]*directoryGroupScope{},
		ownerGroupIDs:        map[string]map[string]struct{}{},
		membershipIndexes:    map[string]map[membershipKey][]*ent.DirectoryMembership{},
		relinkIDs:            map[string][]string{},
		linkedIntegrationIDs: map[string]map[string]struct{}{},
	}
}

// withDirectorySyncBatch installs the batch state on the context
func withDirectorySyncBatch(ctx context.Context, batch *directorySyncBatch) context.Context {
	return directorySyncBatchKey.Set(ctx, batch)
}

// directorySyncBatchFromContext returns the installed batch state, or nil
func directorySyncBatchFromContext(ctx context.Context) *directorySyncBatch {
	return directorySyncBatchKey.GetOr(ctx, nil)
}

// recordIngestChange counts one record whose persistence created or modified a row
func recordIngestChange(ctx context.Context) {
	if batch := directorySyncBatchFromContext(ctx); batch != nil {
		batch.changed++
	}
}

// directoryChangeSet computes the provider-field delta between a resolved directory row and its
// prepared create input, honoring the schema's volatile-field exclusions so a locally set
// confirmation timestamp does not by itself count as a change
func directoryChangeSet(ctx context.Context, db *ent.Client, schema *entityops.Schema, existing any, createInput any) (entityops.ChangeSet, error) {
	row, err := json.Marshal(existing)
	if err != nil {
		return entityops.ChangeSet{}, err
	}

	payload, err := json.Marshal(createInput)
	if err != nil {
		return entityops.ChangeSet{}, err
	}

	return schema.IngestChangeSet(ctx, db, row, payload)
}

// lookupScopeKey derives the lookup scope for one record's owner, instance, and integration
func lookupScopeKey(ownerID string, instanceID string, integrationID string) directoryScopeKey {
	return directoryScopeKey{ownerID: ownerID, instanceID: instanceID, integrationID: integrationID}
}

// accountScope lazily loads and indexes every account in one lookup scope
func (b *directorySyncBatch) accountScope(ctx context.Context, db *ent.Client, key directoryScopeKey) (*directoryAccountScope, error) {
	if scope, ok := b.accountScopes[key]; ok {
		return scope, nil
	}

	pred := directoryaccount.IntegrationID(key.integrationID)
	if key.instanceID != "" {
		pred = directoryaccount.Or(
			directoryaccount.SourceInstanceID(key.instanceID),
			directoryaccount.And(directoryaccount.IntegrationID(key.integrationID), directoryaccount.SourceInstanceIDIsNil()),
		)
	}

	rows, err := db.DirectoryAccount.Query().Where(directoryaccount.OwnerID(key.ownerID), pred).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account scope load failed")

		return nil, err
	}

	scope := &directoryAccountScope{
		byExternalID: map[string][]*ent.DirectoryAccount{},
		byEmail:      map[string][]*ent.DirectoryAccount{},
	}

	for _, row := range rows {
		scope.index(row)
	}

	b.accountScopes[key] = scope

	return scope, nil
}

// index adds one account row to the scope's lookup maps
func (s *directoryAccountScope) index(row *ent.DirectoryAccount) {
	if row.ExternalID != "" {
		s.byExternalID[row.ExternalID] = append(s.byExternalID[row.ExternalID], row)
	}

	if email := lo.FromPtr(row.CanonicalEmail); email != "" {
		s.byEmail[email] = append(s.byEmail[email], row)
	}
}

// addAccount indexes one newly created account into the loaded caches
func (b *directorySyncBatch) addAccount(row *ent.DirectoryAccount, key directoryScopeKey) {
	if scope, ok := b.accountScopes[key]; ok {
		scope.index(row)
	}

	if ids, ok := b.ownerAccountIDs[row.OwnerID]; ok {
		ids[row.ID] = struct{}{}
	}
}

// ownerAccountIDSet lazily loads the owner's directory account id set
func (b *directorySyncBatch) ownerAccountIDSet(ctx context.Context, db *ent.Client, ownerID string) (map[string]struct{}, error) {
	if ids, ok := b.ownerAccountIDs[ownerID]; ok {
		return ids, nil
	}

	rows, err := db.DirectoryAccount.Query().Where(directoryaccount.OwnerID(ownerID)).IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory account id load failed")

		return nil, err
	}

	ids := make(map[string]struct{}, len(rows))
	for _, id := range rows {
		ids[id] = struct{}{}
	}

	b.ownerAccountIDs[ownerID] = ids

	return ids, nil
}

// groupScope lazily loads and indexes every group in one lookup scope
func (b *directorySyncBatch) groupScope(ctx context.Context, db *ent.Client, key directoryScopeKey) (*directoryGroupScope, error) {
	if scope, ok := b.groupScopes[key]; ok {
		return scope, nil
	}

	pred := directorygroup.IntegrationID(key.integrationID)
	if key.instanceID != "" {
		pred = directorygroup.Or(
			directorygroup.SourceInstanceID(key.instanceID),
			directorygroup.And(directorygroup.IntegrationID(key.integrationID), directorygroup.SourceInstanceIDIsNil()),
		)
	}

	rows, err := db.DirectoryGroup.Query().Where(directorygroup.OwnerID(key.ownerID), pred).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group scope load failed")

		return nil, err
	}

	scope := &directoryGroupScope{
		byExternalID: map[string][]*ent.DirectoryGroup{},
		byEmail:      map[string][]*ent.DirectoryGroup{},
	}

	for _, row := range rows {
		scope.index(row)
	}

	b.groupScopes[key] = scope

	return scope, nil
}

// index adds one group row to the scope's lookup maps
func (s *directoryGroupScope) index(row *ent.DirectoryGroup) {
	if row.ExternalID != "" {
		s.byExternalID[row.ExternalID] = append(s.byExternalID[row.ExternalID], row)
	}

	if email := lo.FromPtr(row.Email); email != "" {
		s.byEmail[email] = append(s.byEmail[email], row)
	}
}

// addGroup indexes one newly created group into the loaded caches
func (b *directorySyncBatch) addGroup(row *ent.DirectoryGroup, key directoryScopeKey) {
	if scope, ok := b.groupScopes[key]; ok {
		scope.index(row)
	}

	if ids, ok := b.ownerGroupIDs[row.OwnerID]; ok {
		ids[row.ID] = struct{}{}
	}
}

// ownerGroupIDSet lazily loads the owner's directory group id set
func (b *directorySyncBatch) ownerGroupIDSet(ctx context.Context, db *ent.Client, ownerID string) (map[string]struct{}, error) {
	if ids, ok := b.ownerGroupIDs[ownerID]; ok {
		return ids, nil
	}

	rows, err := db.DirectoryGroup.Query().Where(directorygroup.OwnerID(ownerID)).IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory group id load failed")

		return nil, err
	}

	ids := make(map[string]struct{}, len(rows))
	for _, id := range rows {
		ids[id] = struct{}{}
	}

	b.ownerGroupIDs[ownerID] = ids

	return ids, nil
}

// membershipIndex lazily loads the owner's active memberships keyed by account and group
func (b *directorySyncBatch) membershipIndex(ctx context.Context, db *ent.Client, ownerID string) (map[membershipKey][]*ent.DirectoryMembership, error) {
	if index, ok := b.membershipIndexes[ownerID]; ok {
		return index, nil
	}

	rows, err := db.DirectoryMembership.Query().
		Where(directorymembership.OwnerID(ownerID)).
		Where(directorymembership.RemovedAtIsNil()).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory membership index load failed")

		return nil, err
	}

	index := make(map[membershipKey][]*ent.DirectoryMembership, len(rows))
	for _, row := range rows {
		key := membershipKey{accountID: row.DirectoryAccountID, groupID: row.DirectoryGroupID}
		index[key] = append(index[key], row)
	}

	b.membershipIndexes[ownerID] = index

	return index, nil
}

// addMembership indexes one newly created membership into the loaded index
func (b *directorySyncBatch) addMembership(row *ent.DirectoryMembership) {
	if index, ok := b.membershipIndexes[row.OwnerID]; ok {
		key := membershipKey{accountID: row.DirectoryAccountID, groupID: row.DirectoryGroupID}
		index[key] = append(index[key], row)
	}
}

// resolveAccountFromCache resolves an account reference from the caches; false means fall back to the live lookup
func (b *directorySyncBatch) resolveAccountFromCache(ctx context.Context, db *ent.Client, ownerID string, instanceID string, integrationID string, value string) (string, bool, error) {
	ownerIDs, err := b.ownerAccountIDSet(ctx, db, ownerID)
	if err != nil {
		return "", false, err
	}

	if _, ok := ownerIDs[value]; ok {
		return value, true, nil
	}

	scope, err := b.accountScope(ctx, db, lookupScopeKey(ownerID, instanceID, integrationID))
	if err != nil {
		return "", false, err
	}

	candidates := append([]*ent.DirectoryAccount{}, scope.byExternalID[value]...)
	candidates = append(candidates, scope.byEmail[value]...)

	if legacy, ok := legacyScientificKey(value); ok {
		candidates = append(candidates, scope.byExternalID[legacy]...)
	}

	if newest := newestDirectoryAccount(candidates); newest != nil {
		return newest.ID, true, nil
	}

	return "", false, nil
}

// resolveGroupFromCache resolves a group reference from the caches; false means fall back to the live lookup
func (b *directorySyncBatch) resolveGroupFromCache(ctx context.Context, db *ent.Client, ownerID string, instanceID string, integrationID string, value string) (string, bool, error) {
	ownerIDs, err := b.ownerGroupIDSet(ctx, db, ownerID)
	if err != nil {
		return "", false, err
	}

	if _, ok := ownerIDs[value]; ok {
		return value, true, nil
	}

	scope, err := b.groupScope(ctx, db, lookupScopeKey(ownerID, instanceID, integrationID))
	if err != nil {
		return "", false, err
	}

	candidates := append([]*ent.DirectoryGroup{}, scope.byExternalID[value]...)
	candidates = append(candidates, scope.byEmail[value]...)

	if legacy, ok := legacyScientificKey(value); ok {
		candidates = append(candidates, scope.byExternalID[legacy]...)
	}

	if newest := newestDirectoryGroup(candidates); newest != nil {
		return newest.ID, true, nil
	}

	return "", false, nil
}

// chooseScopedDirectoryRow prefers the integration's own row among scope candidates, else the newest
func chooseScopedDirectoryRow[T any](rows []*T, owned func(*T) bool, newest func([]*T) *T) (*T, error) {
	own := lo.Filter(rows, func(row *T, _ int) bool { return owned(row) })

	switch {
	case len(own) == 1:
		return own[0], nil
	case len(own) > 1:
		return nil, ErrIngestUpsertConflict
	default:
		return newest(rows), nil
	}
}

// newestDirectoryAccount picks the most recently created candidate
func newestDirectoryAccount(candidates []*ent.DirectoryAccount) *ent.DirectoryAccount {
	return lo.MaxBy(candidates, func(a *ent.DirectoryAccount, b *ent.DirectoryAccount) bool {
		return a.CreatedAt.After(b.CreatedAt)
	})
}

// newestDirectoryGroup picks the most recently created candidate
func newestDirectoryGroup(candidates []*ent.DirectoryGroup) *ent.DirectoryGroup {
	return lo.MaxBy(candidates, func(a *ent.DirectoryGroup, b *ent.DirectoryGroup) bool {
		return a.CreatedAt.After(b.CreatedAt)
	})
}

// flushDirectoryConfirmations advances last_seen_at and last_confirmed_run_id for unchanged rows in chunked bulk updates
func flushDirectoryConfirmations(ctx context.Context, db *ent.Client, batch *directorySyncBatch, runID string) error {
	if batch == nil {
		return nil
	}

	now := time.Now()
	vetoCtx := entityops.WithEmissionVetoed(ctx)

	for _, chunk := range lo.Chunk(batch.seenAccountIDs, directoryConfirmChunkSize) {
		if err := db.DirectoryAccount.Update().
			Where(directoryaccount.IDIn(chunk...)).
			SetLastSeenAt(now).
			Exec(vetoCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Int("accounts", len(chunk)).Msg("directory account confirmation flush failed")

			return err
		}
	}

	for _, chunk := range lo.Chunk(batch.confirmedMembershipIDs, directoryConfirmChunkSize) {
		if err := db.DirectoryMembership.Update().
			Where(
				directorymembership.IDIn(chunk...),
				directorymembership.Or(
					directorymembership.LastConfirmedRunIDIsNil(),
					directorymembership.LastConfirmedRunIDLTE(runID),
				),
			).
			SetLastSeenAt(now).
			SetLastConfirmedRunID(runID).
			Exec(vetoCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Int("memberships", len(chunk)).Msg("directory membership confirmation flush failed")

			return err
		}
	}

	return nil
}
