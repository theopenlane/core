package rule_test

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/permissioncache"
	"github.com/theopenlane/core/pkg/testutils"
)

func createExportMutation(t *testing.T) ent.Mutation {
	t.Helper()
	return generated.NewClient(
		generated.EntConfig(
			&entconfig.Config{
				Modules: entconfig.Modules{
					Enabled: true,
				},
			},
		),
	).Export.Create().Mutation()
}

func createControlMutation(t *testing.T) ent.Mutation {
	t.Helper()
	return generated.NewClient(
		generated.EntConfig(
			&entconfig.Config{
				Modules: entconfig.Modules{
					Enabled: true,
				},
			},
		),
	).Control.Create().Mutation()
}

func setupContext(t *testing.T, org string, feats []models.OrgModule) context.Context {
	t.Helper()
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext(t, "org1", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, err := rule.HasFeature(ctx, "base_module")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rule.HasFeature(ctx, "nonexistent_module")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext(t, "org2", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, _, err := rule.HasAnyFeature(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAnyFeature(ctx, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHasAllFeatures(t *testing.T) {
	ctx := setupContext(t, "org3", []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule})

	ok, _, err := rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogComplianceModule)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, missingModule, err := rule.HasAllFeatures(ctx, models.CatalogTrustCenterModule)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, models.CatalogTrustCenterModule, *missingModule)

	ok, _, err = rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestDenyIfMissingAllModulesBase(t *testing.T) {
	tests := []struct {
		title            string
		createMutationFn func() ent.Mutation
		modules          []models.OrgModule
		shouldSkip       bool
		shouldDeny       bool
		expectedError    string
	}{
		{
			title: "Export features present should skip",
			createMutationFn: func() ent.Mutation {
				return createExportMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogBaseModule},
			shouldSkip: true,
		},
		{
			title: "Export features only requires base, should allow",
			createMutationFn: func() ent.Mutation {
				return createExportMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogComplianceModule},
			shouldSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {

			ctx := setupContext(t, "test-org", tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, tt.createMutationFn())

			if tt.shouldSkip {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
				return
			}

			if tt.shouldDeny {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NotContains(t, err.Error(), "skip rule")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDenyIfMissingAllModules(t *testing.T) {
	tests := []struct {
		title            string
		createMutationFn func() ent.Mutation
		modules          []models.OrgModule
		shouldSkip       bool
		shouldDeny       bool
		expectedError    string
	}{
		{
			title: "Missing compliance module",
			createMutationFn: func() ent.Mutation {
				return createControlMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogBaseModule},
			shouldDeny: true,
		},
		{
			title: "Allowed, has compliance module",
			createMutationFn: func() ent.Mutation {
				return createControlMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogComplianceModule},
			shouldSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {

			ctx := setupContext(t, "test-org", tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, tt.createMutationFn())

			if tt.shouldSkip {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
				return
			}

			if tt.shouldDeny {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NotContains(t, err.Error(), "skip rule")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDenyIfMissingAllModules_BypassScenarios(t *testing.T) {
	baseCtx := setupContext(t, "test-org", []models.OrgModule{})

	featureRule := rule.DenyIfMissingAllModules()

	testMutation := createExportMutation(t)

	t.Run("bypass with privacy decision context", func(t *testing.T) {
		ctx := privacy.DecisionContext(baseCtx, privacy.Allow)
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrgSubscriptionContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrgSubscriptionContextKey{})
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with OrganizationCreationContextKey", func(t *testing.T) {
		ctx := contextx.With(baseCtx, auth.OrganizationCreationContextKey{})
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with EmailSignUpToken", func(t *testing.T) {
		ctx := token.NewContextWithSignUpToken(baseCtx, "test@example.com")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with ResetToken", func(t *testing.T) {
		ctx := token.NewContextWithResetToken(baseCtx, "test-reset-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with VerifyToken", func(t *testing.T) {
		ctx := token.NewContextWithVerifyToken(baseCtx, "test-verify-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})

	t.Run("bypass with JobRunnerRegistrationToken", func(t *testing.T) {
		ctx := token.NewContextWithJobRunnerRegistrationToken(baseCtx, "test-job-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "skip rule")
	})
}

func TestModulesEnabledBase(t *testing.T) {
	tests := []struct {
		title       string
		modules     []models.OrgModule
		shouldAllow bool
		expectedErr string
	}{
		{
			title:       "Base module enabled should allow",
			modules:     []models.OrgModule{models.CatalogBaseModule},
			shouldAllow: true,
		},
		{
			title:       "Multiple modules enabled should allow",
			modules:     []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule},
			shouldAllow: true,
		},
		{
			title:       "No modules enabled but base (by default), should allow",
			modules:     []models.OrgModule{},
			shouldAllow: true,
		},
		{
			title:       "Base module is always allowed",
			modules:     []models.OrgModule{models.CatalogComplianceModule}, // base module always allowed
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx := setupContext(t, "test-org", tt.modules)
			mutation := createExportMutation(t)

			rule := rule.DenyIfMissingAllModules()
			err := rule.EvalMutation(ctx, mutation)

			if tt.shouldAllow {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.NotContains(t, err.Error(), "skip rule")
			}
		})
	}
}

func TestModulesEnabled(t *testing.T) {
	tests := []struct {
		title       string
		modules     []models.OrgModule
		shouldAllow bool
		expectedErr string
	}{
		{
			title:       "Compliance module enabled should allow",
			modules:     []models.OrgModule{models.CatalogComplianceModule},
			shouldAllow: true,
		},
		{
			title:       "Multiple modules enabled should allow",
			modules:     []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogExtraEvidenceStorageAddon},
			shouldAllow: true,
		},
		{
			title:       "No modules enabled should allow enabled should deny",
			modules:     []models.OrgModule{},
			shouldAllow: false,
			expectedErr: "features are not enabled",
		},
		{
			title:       "Missing Compliance module",
			modules:     []models.OrgModule{models.CatalogBaseModule, models.CatalogTrustCenterModule},
			shouldAllow: false,
			expectedErr: "features are not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx := setupContext(t, "test-org", tt.modules)
			mutation := createControlMutation(t)

			rule := rule.DenyIfMissingAllModules()
			err := rule.EvalMutation(ctx, mutation)

			if tt.shouldAllow {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "skip rule")
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.NotContains(t, err.Error(), "skip rule")
			}
		})
	}
}

func TestModulesDisabled(t *testing.T) {
	client := generated.NewClient()
	client.EntConfig = &entconfig.Config{
		Modules: entconfig.Modules{
			Enabled: false,
		},
	}

	mutation := &mockMutation{
		client:       client,
		mutationType: "Export",
	}

	ctx := setupContext(t, "test-org", []models.OrgModule{})
	rule := rule.DenyIfMissingAllModules()

	err := rule.EvalMutation(ctx, mutation)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip rule")
}

type mockMutation struct {
	client       *generated.Client
	mutationType string
}

func (m *mockMutation) Op() ent.Op                                                   { return ent.OpCreate }
func (m *mockMutation) Type() string                                                 { return m.mutationType }
func (m *mockMutation) Fields() []string                                             { return []string{} }
func (m *mockMutation) Field(name string) (ent.Value, bool)                          { return nil, false }
func (m *mockMutation) OldField(ctx context.Context, name string) (ent.Value, error) { return nil, nil }
func (m *mockMutation) SetField(name string, value ent.Value) error                  { return nil }
func (m *mockMutation) AddedFields() []string                                        { return []string{} }
func (m *mockMutation) AddedField(name string) (ent.Value, bool)                     { return nil, false }
func (m *mockMutation) AddField(name string, value ent.Value) error                  { return nil }
func (m *mockMutation) ClearedFields() []string                                      { return []string{} }
func (m *mockMutation) FieldCleared(name string) bool                                { return false }
func (m *mockMutation) ClearField(name string) error                                 { return nil }
func (m *mockMutation) RemovedEdges() []string                                       { return []string{} }
func (m *mockMutation) RemovedIDs(name string) []ent.Value                           { return []ent.Value{} }
func (m *mockMutation) ClearedEdges() []string                                       { return []string{} }
func (m *mockMutation) EdgeCleared(name string) bool                                 { return false }
func (m *mockMutation) ClearEdge(name string) error                                  { return nil }
func (m *mockMutation) AddedEdges() []string                                         { return []string{} }
func (m *mockMutation) AddedIDs(name string) []ent.Value                             { return []ent.Value{} }
func (m *mockMutation) Where(ps ...func(s *sql.Selector))                            {}
func (m *mockMutation) WhereP(...func(*sql.Selector))                                {}
func (m *mockMutation) ResetEdge(name string) error                                  { return nil }
func (m *mockMutation) ResetField(name string) error                                 { return nil }
func (m *mockMutation) Client() *generated.Client                                    { return m.client }
