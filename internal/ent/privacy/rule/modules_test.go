package rule_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/token"
	"github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/permissioncache"
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

func createEventMutation(t *testing.T) ent.Mutation {
	t.Helper()
	return generated.NewClient(
		generated.EntConfig(
			&entconfig.Config{
				Modules: entconfig.Modules{
					Enabled: true,
				},
			},
		),
	).Event.Create().Mutation()
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

func setupContext(t *testing.T, feats []models.OrgModule) context.Context {
	t.Helper()
	org := ulids.New().String()
	ctx := context.Background()
	r := testutils.NewRedisClient()
	cache := permissioncache.NewCache(r, permissioncache.WithCacheTTL(time.Minute))

	_ = cache.SetFeatures(ctx, org, feats)
	ctx = permissioncache.WithCache(ctx, cache)
	ctx = auth.WithCaller(ctx, &auth.Caller{OrganizationID: org})
	return ctx
}

func TestHasFeature(t *testing.T) {
	ctx := setupContext(t, []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, err := rule.HasFeature(ctx, "base_module")
	assert.NilError(t, err)
	assert.Check(t, ok)

	ok, err = rule.HasFeature(ctx, "nonexistent_module")
	assert.NilError(t, err)
	assert.Check(t, !ok)
}

func TestHasAnyFeature(t *testing.T) {
	ctx := setupContext(t, []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})

	ok, _, err := rule.HasAnyFeature(ctx, models.CatalogBaseModule, models.CatalogEntityManagementModule)
	assert.NilError(t, err)
	assert.Check(t, ok)

	ok, _, err = rule.HasAnyFeature(ctx, models.CatalogEntityManagementModule)
	assert.NilError(t, err)
	assert.Check(t, !ok)
}

func TestHasAllFeatures(t *testing.T) {
	ctx := setupContext(t, []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule})

	tests := []struct {
		name            string
		modules         []models.OrgModule
		expected        bool
		expectedMissing models.OrgModule
	}{
		{
			name:     "base and compliance",
			modules:  []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule},
			expected: true,
		},
		{
			name:     "base and entity management",
			modules:  []models.OrgModule{models.CatalogBaseModule, models.CatalogEntityManagementModule},
			expected: true,
		},
		{
			name:     "compliance only",
			modules:  []models.OrgModule{models.CatalogComplianceModule},
			expected: true,
		},
		{
			name:     "all enabled modules",
			modules:  []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule},
			expected: true,
		},
		{
			name:            "trust center not enabled",
			modules:         []models.OrgModule{models.CatalogTrustCenterModule},
			expected:        false,
			expectedMissing: models.CatalogTrustCenterModule,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, missingModule, err := rule.HasAllFeatures(ctx, tt.modules...)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, ok))

			if !tt.expected {
				assert.Assert(t, missingModule != nil)
				assert.Check(t, is.Equal(tt.expectedMissing, *missingModule))
			}
		})
	}
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
			title: "Event features present should skip",
			createMutationFn: func() ent.Mutation {
				return createEventMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogBaseModule},
			shouldSkip: true,
		},
		{
			title: "Event features only requires base, should allow",
			createMutationFn: func() ent.Mutation {
				return createEventMutation(t)
			},
			modules:    []models.OrgModule{models.CatalogComplianceModule},
			shouldSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {

			ctx := setupContext(t, tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, tt.createMutationFn())

			if tt.shouldSkip {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), "skip rule"))
				return
			}

			if tt.shouldDeny {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), tt.expectedError))
				assert.Check(t, !strings.Contains(err.Error(), "skip rule"))
				return
			}

			assert.NilError(t, err)
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

			ctx := setupContext(t, tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, tt.createMutationFn())

			if tt.shouldSkip {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), "skip rule"))
				return
			}

			if tt.shouldDeny {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), tt.expectedError))
				assert.Check(t, !strings.Contains(err.Error(), "skip rule"))
				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestDenyIfMissingAllModules_BypassScenarios(t *testing.T) {
	baseCtx := setupContext(t, []models.OrgModule{})

	featureRule := rule.DenyIfMissingAllModules()

	testMutation := createExportMutation(t)

	t.Run("bypass with InternalOperationCaller", func(t *testing.T) {
		ctx := auth.WithCaller(baseCtx, &auth.Caller{Capabilities: auth.CapInternalOperation})
		err := featureRule.EvalMutation(ctx, testMutation)
		assert.Assert(t, err != nil)
		assert.Check(t, is.Contains(err.Error(), "skip rule"))
	})

	t.Run("bypass with EmailSignUpToken", func(t *testing.T) {
		ctx := token.NewContextWithSignUpToken(baseCtx, "test@example.com")
		err := featureRule.EvalMutation(ctx, testMutation)
		assert.Assert(t, err != nil)
		assert.Check(t, is.Contains(err.Error(), "skip rule"))
	})

	t.Run("bypass with ResetToken", func(t *testing.T) {
		ctx := token.NewContextWithResetToken(baseCtx, "test-reset-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		assert.Assert(t, err != nil)
		assert.Check(t, is.Contains(err.Error(), "skip rule"))
	})

	t.Run("bypass with VerifyToken", func(t *testing.T) {
		ctx := token.NewContextWithVerifyToken(baseCtx, "test-verify-token")
		err := featureRule.EvalMutation(ctx, testMutation)
		assert.Assert(t, err != nil)
		assert.Check(t, is.Contains(err.Error(), "skip rule"))
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
			ctx := setupContext(t, tt.modules)
			mutation := createEventMutation(t)

			rule := rule.DenyIfMissingAllModules()
			err := rule.EvalMutation(ctx, mutation)

			if tt.shouldAllow {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), "skip rule"))
			} else {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), tt.expectedErr))
				assert.Check(t, !strings.Contains(err.Error(), "skip rule"))
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
			title:       "Missing required module for Control",
			modules:     []models.OrgModule{models.CatalogBaseModule, models.CatalogEntityManagementModule},
			shouldAllow: false,
			expectedErr: "features are not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx := setupContext(t, tt.modules)
			mutation := createControlMutation(t)

			rule := rule.DenyIfMissingAllModules()
			err := rule.EvalMutation(ctx, mutation)

			if tt.shouldAllow {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), "skip rule"))
			} else {
				assert.Assert(t, err != nil)
				assert.Check(t, is.Contains(err.Error(), tt.expectedErr))
				assert.Check(t, !strings.Contains(err.Error(), "skip rule"))
			}
		})
	}
}

func TestAnyModuleFeature(t *testing.T) {
	tests := []struct {
		title    string
		modules  []models.OrgModule
		expected bool
	}{
		{
			title:    "no modules",
			modules:  []models.OrgModule{},
			expected: false,
		},
		{
			title:    "base module only",
			modules:  []models.OrgModule{models.CatalogBaseModule},
			expected: false,
		},
		{
			title:    "base and compliance",
			modules:  []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule},
			expected: true,
		},
		{
			title:    "addon only",
			modules:  []models.OrgModule{models.CatalogDomainScanningAddon},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx := setupContext(t, tt.modules)

			ok, err := rule.HasFeature(ctx, models.CatalogAnyModule.String())
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, ok))

			ok, _, err = rule.HasAnyFeature(ctx, models.CatalogAnyModule)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, ok))

			ok, missing, err := rule.HasAllFeatures(ctx, models.CatalogBaseModule, models.CatalogAnyModule)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.expected, ok))

			if !tt.expected {
				assert.Assert(t, missing != nil)
				assert.Check(t, is.Equal(models.CatalogAnyModule, *missing))
			}
		})
	}
}

func TestDenyIfMissingAllModulesAnyModule(t *testing.T) {
	tests := []struct {
		title       string
		modules     []models.OrgModule
		shouldAllow bool
	}{
		{
			title:       "no modules should deny",
			modules:     []models.OrgModule{},
			shouldAllow: false,
		},
		{
			title:       "base module only should deny",
			modules:     []models.OrgModule{models.CatalogBaseModule},
			shouldAllow: false,
		},
		{
			title:       "compliance module should allow",
			modules:     []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule},
			shouldAllow: true,
		},
		{
			title:       "trust center module should allow",
			modules:     []models.OrgModule{models.CatalogTrustCenterModule},
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctx := setupContext(t, tt.modules)

			err := rule.DenyIfMissingAllModules().EvalMutation(ctx, createExportMutation(t))
			assert.Assert(t, err != nil)

			if tt.shouldAllow {
				assert.Check(t, is.Contains(err.Error(), "skip rule"))
				return
			}

			assert.Check(t, is.Contains(err.Error(), "features are not enabled"))
			assert.Check(t, !strings.Contains(err.Error(), "skip rule"))
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

	ctx := setupContext(t, []models.OrgModule{})
	rule := rule.DenyIfMissingAllModules()

	err := rule.EvalMutation(ctx, mutation)

	assert.Assert(t, err != nil)
	assert.Check(t, is.Contains(err.Error(), "skip rule"))
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
