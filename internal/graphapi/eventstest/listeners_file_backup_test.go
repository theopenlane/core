//go:build test

package eventstest_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/gala"
	mock_shared "github.com/theopenlane/core/v2/pkg/objects/mocks"
)

// backupSourceProvider is the provider the test files are recorded against and the backup key
const backupSourceProvider = storagetypes.ProviderType("mock")

func TestFileBackupListener(t *testing.T) {
	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule)
	ctx := th.SetContext(user.UserCtx, suite.Client.DB)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.FileBackupListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	// the listener reads the object manager off the ent client, so swap in one with a backup target
	original := suite.Client.DB.ObjectManager

	t.Cleanup(func() {
		suite.Client.DB.ObjectManager = original
	})

	backupState := func(t *testing.T, id string) models.FileBackupState {
		t.Helper()

		f, err := suite.Client.DB.File.Get(ctx, id)
		assert.NilError(t, err)

		return f.BackupState
	}

	expectDownload := func(source *mock_shared.MockProvider, body []byte) {
		source.EXPECT().
			Download(mock.Anything, mock.Anything, mock.Anything).
			Return(&storagetypes.DownloadedFileMetadata{
				File: body,
				Size: int64(len(body)),
			}, nil).
			Maybe()
	}

	t.Run("create replicates to the backup provider", func(t *testing.T) {
		svc, source, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, false)
		assert.NilError(t, err)

		suite.Client.DB.ObjectManager = svc

		body := []byte("backup me")
		expectDownload(source, body)

		backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
		backup.EXPECT().
			Upload(mock.Anything, mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, reader io.Reader, _ *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
				got, readErr := io.ReadAll(reader)
				assert.NilError(t, readErr)
				assert.Check(t, bytes.Equal(got, body), "replica should carry the source bytes")

				return &storagetypes.UploadedFileMetadata{
					FileMetadata: storagetypes.FileMetadata{
						Key:     "backups/replica.txt",
						Bucket:  "backup-bucket",
						FullURI: "mock://backup-bucket/backups/replica.txt",
					},
				}, nil
			}).
			Once()

		f := newBackupFile(t, ctx, user.OrganizationID, "listener-backup-create.txt", nil)

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusCompleted
		}, "create should replicate the file to its backup provider")

		state := backupState(t, f.ID)
		assert.Equal(t, state.Key, "backups/replica.txt")
		assert.Equal(t, state.Bucket, "backup-bucket")
		assert.Equal(t, state.Attempts, 1)
		assert.Check(t, state.CompletedAt != nil)
	})

	t.Run("replication failure on the last attempt exhausts the backup", func(t *testing.T) {
		svc, source, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, false)
		assert.NilError(t, err)

		suite.Client.DB.ObjectManager = svc

		expectDownload(source, []byte("backup me"))

		backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
		backup.EXPECT().
			Upload(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("destination unavailable")).
			Maybe()

		// seeded one short of the cap so the failure exhausts rather than retrying into a later subtest
		f := newBackupFile(t, ctx, user.OrganizationID, "listener-backup-failure.txt", &models.FileBackupState{
			Status:   enums.FileBackupStatusFailed,
			Attempts: hooks.MaxFileBackupAttempts - 1,
		})

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusExhausted
		}, "a failure on the last attempt should exhaust the backup")

		state := backupState(t, f.ID)
		assert.Equal(t, state.Attempts, hooks.MaxFileBackupAttempts)
		assert.Check(t, state.Error != "")
		assert.Check(t, state.CompletedAt == nil)
	})

	t.Run("read from backup suppresses replication", func(t *testing.T) {
		svc, _, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, true)
		assert.NilError(t, err)

		suite.Client.DB.ObjectManager = svc

		backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()

		f := newBackupFile(t, ctx, user.OrganizationID, "listener-backup-readfrombackup.txt", nil)

		waitForGala(t, setup.Runtime)

		// no upload expected, so any replication attempt fails the mock
		state := backupState(t, f.ID)
		assert.Equal(t, state.Status, enums.FileBackupStatus(""))
		assert.Equal(t, state.Attempts, 0)
	})

	t.Run("explicit backup request replicates the file", func(t *testing.T) {
		// create while reads come from the backup so the create listener no-ops, leaving the file
		// unreplicated for the request path to pick up
		idle, _, idleBackup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, true)
		assert.NilError(t, err)

		idleBackup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
		suite.Client.DB.ObjectManager = idle

		f := newBackupFile(t, ctx, user.OrganizationID, "listener-backup-requested.txt", nil)

		waitForGala(t, setup.Runtime)
		assert.Equal(t, backupState(t, f.ID).Status, enums.FileBackupStatus(""))

		svc, source, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, false)
		assert.NilError(t, err)

		suite.Client.DB.ObjectManager = svc

		expectDownload(source, []byte("backup me"))

		backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
		backup.EXPECT().
			Upload(mock.Anything, mock.Anything, mock.Anything).
			Return(&storagetypes.UploadedFileMetadata{
				FileMetadata: storagetypes.FileMetadata{
					Key:    "backups/requested.txt",
					Bucket: "backup-bucket",
				},
			}, nil).
			Once()

		_, err = suite.GalaRuntime.EmitWithHeaders(ctx, hooks.FileBackupTopic.Name, hooks.FileBackupRequest{FileID: f.ID}, gala.Headers{})
		assert.NilError(t, err)

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusCompleted
		}, "an explicit backup request should replicate the file")

		state := backupState(t, f.ID)
		assert.Equal(t, state.Key, "backups/requested.txt")
		assert.Equal(t, state.Attempts, 1)
	})

	t.Run("file created by an org member replicates", func(t *testing.T) {
		setupReplicatingBackup(t, "backups/member.txt")

		f := newBackupFile(t, user.UserCtx, user.OrganizationID, "listener-backup-member.txt", nil)

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusCompleted
		}, "a file uploaded by an org member should replicate")
	})

	t.Run("file created by an anonymous questionnaire respondent replicates", func(t *testing.T) {
		setupReplicatingBackup(t, "backups/respondent.txt")

		// covers the listener under the anonymous respondent caller which is why the privacy.Allow is here
		respondent := auth.NewQuestionnaireCaller(user.OrganizationID, ulids.New().String(), "Anonymous Respondent", "")
		respondentCtx := privacy.DecisionContext(generated.NewContext(auth.WithCaller(context.Background(), respondent), suite.Client.DB), privacy.Allow)
		f := newBackupFile(t, respondentCtx, user.OrganizationID, "listener-backup-respondent.txt", nil)

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusCompleted
		}, "a file uploaded by an anonymous respondent should replicate")
	})

	t.Run("unrelated update does not re-replicate", func(t *testing.T) {
		svc, source, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, false)
		assert.NilError(t, err)

		suite.Client.DB.ObjectManager = svc

		expectDownload(source, []byte("backup me"))

		backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
		backup.EXPECT().
			Upload(mock.Anything, mock.Anything, mock.Anything).
			Return(&storagetypes.UploadedFileMetadata{
				FileMetadata: storagetypes.FileMetadata{
					Key:    "backups/replica.txt",
					Bucket: "backup-bucket",
				},
			}, nil).
			Once()

		f := newBackupFile(t, ctx, user.OrganizationID, "listener-backup-unrelated.txt", nil)

		waitForCondition(t, func() bool {
			return backupState(t, f.ID).Status == enums.FileBackupStatusCompleted
		}, "create should replicate the file")

		completedAt := backupState(t, f.ID).CompletedAt

		// storage_provider untouched, so the field gate should reject this
		assert.NilError(t, suite.Client.DB.File.UpdateOneID(f.ID).SetProvidedFileName("renamed.txt").Exec(ctx))

		waitForGala(t, setup.Runtime)

		assert.Equal(t, backupState(t, f.ID).CompletedAt.Equal(*completedAt), true)
		assert.Equal(t, backupState(t, f.ID).Attempts, 1)
	})
}

func newBackupFile(t *testing.T, ctx context.Context, orgID, name string, state *models.FileBackupState) *generated.File {
	t.Helper()

	create := suite.Client.DB.File.Create()
	if state != nil {
		create = create.SetBackupState(*state)
	}

	f, err := create.
		SetProvidedFileName(name).
		SetProvidedFileExtension(".txt").
		SetDetectedContentType("text/plain").
		SetStorageProvider(backupSourceProvider.String()).
		SetStoragePath(orgID + "/" + name).
		SetStorageVolume("source-bucket").
		Save(ctx)
	assert.NilError(t, err)

	return f
}

func setupReplicatingBackup(t *testing.T, key string) {
	t.Helper()

	svc, source, backup, err := coreutils.MockStorageServiceWithBackup(t, backupSourceProvider, false)
	assert.NilError(t, err)

	suite.Client.DB.ObjectManager = svc

	source.EXPECT().
		Download(mock.Anything, mock.Anything, mock.Anything).
		Return(&storagetypes.DownloadedFileMetadata{File: []byte("backup me"), Size: int64(len("backup me"))}, nil).
		Maybe()

	backup.EXPECT().ProviderType().Return(backupSourceProvider).Maybe()
	backup.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything).
		Return(&storagetypes.UploadedFileMetadata{
			FileMetadata: storagetypes.FileMetadata{
				Key:    key,
				Bucket: "backup-bucket",
			},
		}, nil).
		Once()
}
