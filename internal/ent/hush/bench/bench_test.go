package bench_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"time"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"

	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	hushpkg "github.com/theopenlane/core/internal/ent/hush"
	"github.com/theopenlane/core/internal/ent/schema"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
	coreutils "github.com/theopenlane/core/pkg/testutils"

	// import generated runtime which is required to prevent cyclical dependencies
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

const (
	benchmarkSecretValue = "this-is-a-secret-value-that-will-be-encrypted-and-should-be-reasonably-long-to-simulate-real-world-data"
	benchmarkPlainValue  = "this-is-a-plain-value-that-will-not-be-encrypted-but-should-be-the-same-length-as-the-secret-value"
)

type benchmarkContext struct {
	client *generated.Client
	tf     *testutils.TestFixture
	ctx    context.Context
	userID string
	orgID  string
}

// testUserDetails is a struct that holds the details of a test user
type testUserDetails struct {
	// ID is the ID of the user
	ID string
	// UserInfo contains all the details of the user
	UserInfo *generated.User
	// PersonalOrgID is the ID of the personal organization of the user
	PersonalOrgID string
	// OrganizationID is the ID of the organization of the user
	OrganizationID string
	// UserCtx is the context of the user that can be used for the test requests that require authentication
	UserCtx context.Context
}

type userInput struct {
	email         string
	password      string
	confirmedUser bool
}

func setupBenchmarkClient(b *testing.B) *benchmarkContext {
	b.Helper()

	// Setup encryption key for benchmarking
	_, err := hooks.GenerateTinkKeyset()
	require.NoError(b, err)

	// Create PostgreSQL test fixture
	tf := entdb.NewTestFixture()

	// Create FGA test fixture
	ofgaTF := fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile("../../../fga/model/model.fga"))

	ctx := context.Background()

	// setup fga client
	fgaClient, err := ofgaTF.NewFgaClient(ctx)
	require.NoError(b, err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute)
	require.NoError(b, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)

	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	// setup otp manager
	otpOpts := []totp.ConfigOption{
		totp.WithCodeLength(6),
		totp.WithIssuer("authenticator.local"),
		totp.WithSecret(totp.Secret{
			Version: 0,
			Key:     "9f0c6da662f018b58b04a093e2dbb2e1",
		}),
		totp.WithRedis(rc),
	}

	otpMan := totp.NewOTP(otpOpts...)

	pool := soiree.NewPondPool(
		soiree.WithMaxWorkers(100),
		soiree.WithName("ent_client_pool"),
	)

	opts := []generated.Option{
		generated.Authz(*fgaClient),
		generated.Emailer(&emailtemplates.Config{
			CompanyName: "Benchmark Test Inc.",
		}),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.EntConfig(&entconfig.Config{}),
		generated.TOTP(&totp.Client{
			Manager: otpMan,
		}),
		generated.PondPool(pool),
	}

	// create database connection
	jobOpts := []riverqueue.Option{riverqueue.WithConnectionURI(tf.URI)}

	db, err := entdb.NewTestClient(ctx, tf, jobOpts, opts)
	require.NoError(b, err, "failed opening connection to database")

	// truncate river tables
	err = db.Job.TruncateRiverTables(ctx)
	require.NoError(b, err)

	// Create test user with organization following the handlers test pattern
	testUser := createTestUser(b, ctx, db)

	return &benchmarkContext{
		client: db,
		tf:     tf,
		ctx:    testUser.UserCtx,
		userID: testUser.ID,
		orgID:  testUser.OrganizationID,
	}
}

// createTestUser creates a test user with organization following the handlers test pattern
func createTestUser(b *testing.B, ctx context.Context, db *generated.Client) testUserDetails {
	b.Helper()

	testUser := testUserDetails{}

	ctx = generated.NewContext(ctx, db)

	// Allow all operations for test setup
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Create a test user with confirmed status
	userSetting, err := db.UserSetting.Create().
		SetEmailConfirmed(true).
		SetIsTfaEnabled(true).
		Save(ctx)
	require.NoError(b, err)

	email := "benchmarkuser@example.com"
	password := "benchmark123!"

	testUser.UserInfo, err = db.User.Create().
		SetEmail(email).
		SetFirstName("Benchmark").
		SetLastName("User").
		SetPassword(password).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSetting(userSetting).
		Save(ctx)
	require.NoError(b, err)

	testUser.ID = testUser.UserInfo.ID

	// Get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	require.NoError(b, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// Setup user context with the personal org
	userCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// Set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	userCtx = privacy.DecisionContext(userCtx, privacy.Allow)

	// Add client to context, required for hooks that expect the client to be in the context
	userCtx = generated.NewContext(userCtx, db)

	// Create a non-personal test organization
	orgSetting, err := db.OrganizationSetting.Create().
		SetBillingEmail(testUser.UserInfo.Email).
		Save(userCtx)
	require.NoError(b, err)

	testOrg, err := db.Organization.Create().
		SetName("Benchmark Test Organization").
		SetSettingID(orgSetting.ID).
		Save(userCtx)
	require.NoError(b, err)

	testUser.OrganizationID = testOrg.ID

	// Setup user context with the org (and not the personal org)
	testUser.UserCtx = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)

	// Set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	testUser.UserCtx = privacy.DecisionContext(testUser.UserCtx, privacy.Allow)

	// Add client to context, required for hooks that expect the client to be in the context
	testUser.UserCtx = generated.NewContext(testUser.UserCtx, db)

	return testUser
}

func (bc *benchmarkContext) cleanup() {
	bc.client.CloseAll()
	testutils.TeardownFixture(bc.tf)
}

// BenchmarkEncryptedFieldCreate measures performance of creating entities with encrypted fields
func BenchmarkEncryptedFieldCreate(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bc.client.Hush.Create().
			SetName(fmt.Sprintf("benchmark-entity-%d", i)).
			SetDescription("Benchmark entity for encrypted field testing").
			SetKind("benchmark").
			SetSecretName("benchmark-secret").
			SetSecretValue(benchmarkSecretValue). // This field gets encrypted
			Save(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNonEncryptedFieldCreate measures performance of creating entities with only non-encrypted fields
func BenchmarkNonEncryptedFieldCreate(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := bc.client.Hush.Create().
			SetName(fmt.Sprintf("benchmark-entity-%d", i)).
			SetDescription(benchmarkPlainValue). // Same length as secret but not encrypted
			SetKind("benchmark").
			SetSecretName("benchmark-secret").
			// Note: Not setting SecretValue to avoid encryption overhead
			Save(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncryptedFieldRead measures performance of reading entities with encrypted fields
func BenchmarkEncryptedFieldRead(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	// Pre-populate with test data
	const numEntities = 1000
	entities := make([]*generated.Hush, numEntities)

	for i := 0; i < numEntities; i++ {
		entity, err := bc.client.Hush.Create().
			SetName(fmt.Sprintf("benchmark-entity-%d", i)).
			SetDescription("Benchmark entity for encrypted field testing").
			SetKind("benchmark").
			SetSecretName("benchmark-secret").
			SetSecretValue(benchmarkSecretValue).
			Save(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entity, err := bc.client.Hush.Get(bc.ctx, entities[i%numEntities].ID)
		if err != nil {
			b.Fatal(err)
		}

		// Access the encrypted field to trigger decryption
		_ = entity.SecretValue
	}
}

// BenchmarkNonEncryptedFieldRead measures performance of reading entities with only non-encrypted fields
func BenchmarkNonEncryptedFieldRead(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	// Pre-populate with test data
	const numEntities = 1000
	entities := make([]*generated.Hush, numEntities)

	for i := 0; i < numEntities; i++ {
		entity, err := bc.client.Hush.Create().
			SetName(fmt.Sprintf("benchmark-entity-%d", i)).
			SetDescription(benchmarkPlainValue).
			SetKind("benchmark").
			SetSecretName("benchmark-secret").
			// Note: Not setting SecretValue to avoid encryption overhead
			Save(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entity, err := bc.client.Hush.Get(bc.ctx, entities[i%numEntities].ID)
		if err != nil {
			b.Fatal(err)
		}

		// Access non-encrypted fields
		_ = entity.Name
		_ = entity.Description
	}
}

// BenchmarkBulkEncryptedFieldCreate measures bulk creation performance with encrypted fields
func BenchmarkBulkEncryptedFieldCreate(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	const batchSize = 100

	for i := 0; i < b.N; i++ {
		// Create batch of entities
		bulk := make([]*generated.HushCreate, batchSize)
		for j := 0; j < batchSize; j++ {
			bulk[j] = bc.client.Hush.Create().
				SetName(fmt.Sprintf("bulk-entity-%d-%d", i, j)).
				SetDescription("Bulk benchmark entity for encrypted field testing").
				SetKind("bulk-benchmark").
				SetSecretName("bulk-secret").
				SetSecretValue(benchmarkSecretValue) // This field gets encrypted
		}

		_, err := bc.client.Hush.CreateBulk(bulk...).Save(bc.ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBulkNonEncryptedFieldCreate measures bulk creation performance with only non-encrypted fields
func BenchmarkBulkNonEncryptedFieldCreate(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	const batchSize = 100

	for i := 0; i < b.N; i++ {
		// Create batch of entities
		bulk := make([]*generated.HushCreate, batchSize)
		for j := 0; j < batchSize; j++ {
			bulk[j] = bc.client.Hush.Create().
				SetName(fmt.Sprintf("bulk-entity-%d-%d", i, j)).
				SetDescription(benchmarkPlainValue). // Same length as secret but not encrypted
				SetKind("bulk-benchmark").
				SetSecretName("bulk-secret")
			// Note: Not setting SecretValue to avoid encryption overhead
		}

		_, err := bc.client.Hush.CreateBulk(bulk...).Save(bc.ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBulkEncryptedFieldRead measures bulk read performance with encrypted fields
func BenchmarkBulkEncryptedFieldRead(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	// Pre-populate with test data
	const numEntities = 1000
	bulk := make([]*generated.HushCreate, numEntities)
	for i := 0; i < numEntities; i++ {
		bulk[i] = bc.client.Hush.Create().
			SetName(fmt.Sprintf("bulk-read-entity-%d", i)).
			SetDescription("Bulk read benchmark entity").
			SetKind("bulk-read").
			SetSecretName("bulk-read-secret").
			SetSecretValue(benchmarkSecretValue)
	}

	_, err := bc.client.Hush.CreateBulk(bulk...).Save(bc.ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entities, err := bc.client.Hush.Query().
			Where(hush.Kind("bulk-read")).
			Limit(100).
			All(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}

		// Access encrypted fields to trigger decryption
		for _, entity := range entities {
			_ = entity.SecretValue
		}
	}
}

// BenchmarkBulkNonEncryptedFieldRead measures bulk read performance with only non-encrypted fields
func BenchmarkBulkNonEncryptedFieldRead(b *testing.B) {
	bc := setupBenchmarkClient(b)
	defer bc.cleanup()

	// Pre-populate with test data
	const numEntities = 1000
	bulk := make([]*generated.HushCreate, numEntities)
	for i := 0; i < numEntities; i++ {
		bulk[i] = bc.client.Hush.Create().
			SetName(fmt.Sprintf("bulk-read-entity-%d", i)).
			SetDescription(benchmarkPlainValue).
			SetKind("bulk-read").
			SetSecretName("bulk-read-secret")
		// Note: Not setting SecretValue to avoid encryption overhead
	}

	_, err := bc.client.Hush.CreateBulk(bulk...).Save(bc.ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entities, err := bc.client.Hush.Query().
			Where(hush.Kind("bulk-read")).
			Limit(100).
			All(bc.ctx)

		if err != nil {
			b.Fatal(err)
		}

		// Access non-encrypted fields
		for _, entity := range entities {
			_ = entity.Name
			_ = entity.Description
		}
	}
}

// BenchmarkEncryptionOverhead measures the direct encryption overhead
func BenchmarkEncryptionOverhead(b *testing.B) {
	// Setup encryption key
	_, err := hooks.GenerateTinkKeyset()
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Direct encryption benchmark
		encrypted, err := hooks.Encrypt([]byte(benchmarkSecretValue))
		if err != nil {
			b.Fatal(err)
		}

		// Also benchmark decryption
		_, err = hooks.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFieldDetection measures the performance of automatic field detection
func BenchmarkFieldDetection(b *testing.B) {
	// Setup encryption key
	_, err := hooks.GenerateTinkKeyset()
	require.NoError(b, err)

	hushSchema := schema.Hush{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// This benchmarks the reflection-based field detection
		encryptionHooks := hushpkg.AutoEncryptionHook(hushSchema)

		// Ensure we got hooks back
		if len(encryptionHooks) == 0 {
			b.Fatal("Expected to get encryption hooks")
		}
	}
}
