//go:build examples

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/examples/common"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/providers/disk"
	s3provider "github.com/theopenlane/core/pkg/objects/storage/providers/s3"
	"github.com/theopenlane/eddy"
)

func multiProviderCommand() *cli.Command {
	const skipSetupFlag = "skip-setup"
	const skipTeardownFlag = "skip-teardown"

	return &cli.Command{
		Name:  "multi-provider",
		Usage: "Demonstrate provider resolution across disk and S3 backends with optional benchmarking",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: skipSetupFlag, Usage: "Assume infrastructure is already running and provisioned"},
			&cli.BoolFlag{Name: skipTeardownFlag, Usage: "Leave supporting services running after the example"},
		},
		Commands: []*cli.Command{
			{
				Name:  "setup",
				Usage: "Start docker services and seed credentials",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					out := cmd.Writer
					if out == nil {
						out = os.Stdout
					}
					return multiProviderSetup(ctx, out)
				},
			},
			{
				Name:  "teardown",
				Usage: "Stop docker services and remove containers",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					out := cmd.Writer
					if out == nil {
						out = os.Stdout
					}
					return multiProviderTeardown(ctx, out)
				},
			},
			{
				Name:  "benchmark",
				Usage: "Run high-throughput benchmark across multiple tenants",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "ops", Usage: "Number of operations per tenant", Value: 100},
					&cli.IntFlag{Name: "concurrent", Usage: "Number of concurrent workers", Value: 10},
					&cli.IntFlag{Name: "tenants", Usage: "Number of tenants to provision", Value: 10},
					&cli.IntFlag{Name: "parallel", Usage: "Number of parallel provisioning workers", Value: 5},
					&cli.StringFlag{Name: "config", Usage: "Path to tenant configuration file", Value: "tenants.json"},
				},
				Commands: []*cli.Command{
					{
						Name:  "setup",
						Usage: "Provision tenants for benchmarking",
						Flags: []cli.Flag{
							&cli.IntFlag{Name: "tenants", Usage: "Number of tenants to provision", Value: 10},
							&cli.IntFlag{Name: "parallel", Usage: "Number of parallel provisioning workers", Value: 5},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							out := cmd.Writer
							if out == nil {
								out = os.Stdout
							}
							return benchmarkSetup(ctx, out, benchmarkSetupConfig{
								TenantCount: cmd.Int("tenants"),
								Parallel:    cmd.Int("parallel"),
							})
						},
					},
					{
						Name:  "setup-1000",
						Usage: "Provision 1000 tenants for large-scale benchmarking",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							out := cmd.Writer
							if out == nil {
								out = os.Stdout
							}
							return benchmarkSetup(ctx, out, benchmarkSetupConfig{TenantCount: 1000, Parallel: 20})
						},
					},
					{
						Name:  "teardown",
						Usage: "Remove generated tenants and docker services",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							out := cmd.Writer
							if out == nil {
								out = os.Stdout
							}
							return benchmarkTeardown(ctx, out)
						},
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := benchmarkRunConfig{
						OpsPerTenant:  cmd.Int("ops"),
						Concurrent:    cmd.Int("concurrent"),
						TenantCfgPath: cmd.String("config"),
					}
					out := cmd.Writer
					if out == nil {
						out = os.Stdout
					}
					return runBenchmark(ctx, out, cfg)
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := multiProviderRunConfig{
				SkipSetup:    cmd.Bool(skipSetupFlag),
				SkipTeardown: cmd.Bool(skipTeardownFlag),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runMultiProvider(ctx, out, cfg)
		},
	}
}

type multiProviderRunConfig struct {
	SkipSetup    bool
	SkipTeardown bool
}

type multiProviderStorageClient struct {
	Provider storagetypes.Provider
	Type     string
}

type multiProviderCacheKey struct {
	TenantID   string
	ProviderID string
}

func (k multiProviderCacheKey) String() string {
	return fmt.Sprintf("%s:%s", k.TenantID, k.ProviderID)
}

func runMultiProvider(ctx context.Context, out io.Writer, cfg multiProviderRunConfig) error {
	if !cfg.SkipSetup {
		if err := multiProviderSetup(ctx, out); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "=== Multi-Provider Object Storage Example ===")

	pool := eddy.NewClientPool[*multiProviderStorageClient](10 * time.Minute)
	service := eddy.NewClientService(pool, eddy.WithConfigClone[
		*multiProviderStorageClient,
		storage.ProviderCredentials](cloneProviderOptions))

	resolver := createMultiProviderResolver()

	scenarios := []struct {
		heading  string
		provider string
		tenantID string
		userID   string
	}{
		{"1. Testing Disk Provider...", "disk", "tenant1", "provider1"},
		{"\n2. Testing S3 Provider (MinIO - Provider 1)...", "s3-provider1", "tenant1", "provider1"},
		{"\n3. Testing S3 Provider (MinIO - Provider 2)...", "s3-provider2", "tenant2", "provider2"},
		{"\n4. Testing S3 Provider (MinIO - Provider 3)...", "s3-provider3", "tenant3", "provider3"},
	}

	for _, scenario := range scenarios {
		fmt.Fprintln(out, scenario.heading)
		if err := exerciseProvider(ctx, out, service, resolver, scenario.provider, scenario.tenantID, scenario.userID); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "\n5. Concurrent Operations Across All Providers...")
	if err := multiProviderConcurrent(ctx, out, service, resolver); err != nil {
		return err
	}

	fmt.Fprintln(out, "\n6. Pool Statistics...")
	printMultiProviderPoolStats(out)
	fmt.Fprintln(out, "\n=== Example completed successfully ===")

	if !cfg.SkipTeardown {
		if err := multiProviderTeardown(ctx, out); err != nil {
			return err
		}
	}

	return nil
}

func createMultiProviderResolver() *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := eddy.NewResolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &multiProviderDiskBuilder{providerType: "disk"}
	resolver.AddRule(eddy.NewRule[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
		WhenFunc(func(ctx context.Context) bool {
			provider, _ := ctx.Value(multiProviderTypeKey{}).(string)
			return provider == "disk"
		}).
		Resolve(func(context.Context) (*eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: diskBuilder,
				Output:  storage.ProviderCredentials{},
				Config: storage.NewProviderOptions(
					storage.WithBucket("./tmp/disk-storage"),
					storage.WithBasePath("./tmp/disk-storage"),
					storage.WithLocalURL("http://localhost:8080/files"),
				),
			}, nil
		}))

	s3Configs := []struct {
		name      string
		accessKey string
		secretKey string
		bucket    string
		endpoint  string
	}{
		{"s3-provider1", "provider1", "provider1secret", "provider1-bucket", "http://localhost:19000"},
		{"s3-provider2", "provider2", "provider2secret", "provider2-bucket", "http://localhost:19000"},
		{"s3-provider3", "provider3", "provider3secret", "provider3-bucket", "http://localhost:19000"},
	}

	for _, cfg := range s3Configs {
		providerName := cfg.name
		builder := &multiProviderS3Builder{providerType: providerName}
		resolver.AddRule(eddy.NewRule[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				provider, _ := ctx.Value(multiProviderTypeKey{}).(string)
				return provider == providerName
			}).
			Resolve(func(context.Context) (*eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &eddy.ResolvedProvider[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]{
					Builder: builder,
					Output: storage.ProviderCredentials{
						AccessKeyID:     cfg.accessKey,
						SecretAccessKey: cfg.secretKey,
					},
					Config: storage.NewProviderOptions(
						storage.WithBucket(cfg.bucket),
						storage.WithRegion("us-east-1"),
						storage.WithEndpoint(cfg.endpoint),
					),
				}, nil
			}))
	}

	return resolver
}

type multiProviderTypeKey struct{}

type multiProviderTenantKey struct{}

func exerciseProvider(ctx context.Context, out io.Writer, service *eddy.ClientService[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], providerType, tenantID, userID string) error {
	ctx = context.WithValue(ctx, multiProviderTypeKey{}, providerType)
	ctx = context.WithValue(ctx, multiProviderTenantKey{}, tenantID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		return fmt.Errorf("provider resolution failed for %s", providerType)
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := multiProviderCacheKey{
		TenantID:   tenantID,
		ProviderID: resolution.Builder.ProviderType(),
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
	if !clientOpt.IsPresent() {
		return fmt.Errorf("client acquisition failed for %s", providerType)
	}
	client := clientOpt.MustGet()

	objService := storage.NewObjectService()
	content := strings.NewReader(fmt.Sprintf("Test content for %s-%s-%s", providerType, tenantID, userID))

	uploadOpts := &storage.UploadOptions{
		FileName:    fmt.Sprintf("test-%s.txt", userID),
		ContentType: "text/plain",
	}

	uploaded, err := objService.Upload(ctx, client.Provider, content, uploadOpts)
	if err != nil {
		return fmt.Errorf("upload failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Uploaded to %s: %s\n", providerType, uploaded.Key)

	storageFile := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: uploaded.Key, Size: uploaded.Size, ContentType: uploaded.ContentType}}

	downloaded, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{})
	if err != nil {
		return fmt.Errorf("download failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Downloaded from %s: %d bytes\n", providerType, len(downloaded.File))

	if err := objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
		return fmt.Errorf("delete failed for %s: %w", providerType, err)
	}
	fmt.Fprintf(out, "Deleted from %s\n", providerType)

	return nil
}

func multiProviderConcurrent(ctx context.Context, out io.Writer, service *eddy.ClientService[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*multiProviderStorageClient, storage.ProviderCredentials, *storage.ProviderOptions]) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	providers := []string{"disk", "s3-provider1", "s3-provider2", "s3-provider3"}
	errCh := make(chan error, len(providers))

	for i, provider := range providers {
		wg.Add(1)
		go func(idx int, prov string) {
			defer wg.Done()

			testCtx := context.WithValue(ctx, multiProviderTypeKey{}, prov)
			testCtx = context.WithValue(testCtx, multiProviderTenantKey{}, fmt.Sprintf("tenant%d", idx))

			resolutionOpt := resolver.Resolve(testCtx)
			if !resolutionOpt.IsPresent() {
				errCh <- fmt.Errorf("resolve failed for %s", prov)
				return
			}
			resolution := resolutionOpt.MustGet()

			cacheKey := multiProviderCacheKey{
				TenantID:   fmt.Sprintf("tenant%d", idx),
				ProviderID: resolution.Builder.ProviderType(),
			}

			clientOpt := service.GetClient(testCtx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
			if !clientOpt.IsPresent() {
				errCh <- fmt.Errorf("client retrieval failed for %s", prov)
				return
			}
			client := clientOpt.MustGet()

			objService := storage.NewObjectService()
			content := strings.NewReader(fmt.Sprintf("Concurrent test %d", idx))

			uploadOpts := &storage.UploadOptions{
				FileName:    "concurrent.txt",
				ContentType: "text/plain",
			}

			if _, err := objService.Upload(testCtx, client.Provider, content, uploadOpts); err != nil {
				errCh <- fmt.Errorf("concurrent upload failed for %s: %w", prov, err)
				return
			}

			mu.Lock()
			fmt.Fprintf(out, "Concurrent upload to %s completed\n", prov)
			mu.Unlock()
		}(i, provider)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "All concurrent operations completed")
	return nil
}

func printMultiProviderPoolStats(out io.Writer) {
	fmt.Fprintln(out, "   Pool statistics:")
	fmt.Fprintln(out, "   - Providers configured: 4 (1 disk + 3 S3)")
	fmt.Fprintln(out, "   - Clients cached: Based on tenant+provider combinations")
	fmt.Fprintln(out, "   - TTL: 10 minutes")
}

type multiProviderDiskBuilder struct {
	providerType string
}

func (b *multiProviderDiskBuilder) Build(_ context.Context, _ storage.ProviderCredentials, options *storage.ProviderOptions) (*multiProviderStorageClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	} else {
		opts = storage.NewProviderOptions(storage.WithBucket("./tmp/disk-storage"), storage.WithBasePath("./tmp/disk-storage"))
	}
	if err := os.MkdirAll(opts.Bucket, 0o755); err != nil {
		return nil, err
	}

	provider, err := disk.NewDiskProvider(opts)
	if err != nil {
		return nil, err
	}

	return &multiProviderStorageClient{
		Provider: provider,
		Type:     b.providerType,
	}, nil
}

func (b *multiProviderDiskBuilder) ProviderType() string {
	return b.providerType
}

type multiProviderS3Builder struct {
	providerType string
}

func (b *multiProviderS3Builder) Build(_ context.Context, credentials storage.ProviderCredentials, options *storage.ProviderOptions) (*multiProviderStorageClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	}
	opts.Apply(storage.WithCredentials(credentials))

	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true))
	if err != nil {
		return nil, err
	}

	return &multiProviderStorageClient{
		Provider: provider,
		Type:     b.providerType,
	}, nil
}

func (b *multiProviderS3Builder) ProviderType() string {
	return b.providerType
}

func multiProviderSetup(ctx context.Context, out io.Writer, composeArgs ...string) error {
	fmt.Fprintln(out, "Starting docker services...")
	args := append([]string{"-f", composeFilePath(), "up", "-d"}, composeArgs...)
	if err := common.RunCommand(ctx, out, "docker-compose", args...); err != nil {
		return err
	}

	fmt.Fprintln(out, "Waiting for MinIO...")
	if err := common.WaitForMinIO(ctx, "objects-examples-minio", "admin", "adminsecretpassword"); err != nil {
		return err
	}

	fmt.Fprintln(out, "Configuring MinIO users and buckets...")
	users := []common.MinIOUser{
		{Username: "provider1", Password: "provider1secret", Bucket: "provider1-bucket"},
		{Username: "provider2", Password: "provider2secret", Bucket: "provider2-bucket"},
		{Username: "provider3", Password: "provider3secret", Bucket: "provider3-bucket"},
	}

	if err := common.SetupMinIOUsers(ctx, "objects-examples-minio", users); err != nil {
		return err
	}

	fmt.Fprintln(out, "Infrastructure ready.")
	return nil
}

func multiProviderTeardown(ctx context.Context, out io.Writer) error {
	fmt.Fprintln(out, "Stopping docker services...")
	return common.RunCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "down", "--remove-orphans")
}

// Benchmark functionality (formerly multi-tenant)

type benchmarkSetupConfig struct {
	TenantCount int
	Parallel    int
}

type benchmarkRunConfig struct {
	OpsPerTenant  int
	Concurrent    int
	TenantCfgPath string
}

type benchmarkTenant struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type benchmarkClient struct {
	Provider storagetypes.Provider
	TenantID string
}

type benchmarkStats struct {
	uploads     atomic.Int64
	downloads   atomic.Int64
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64
	errors      atomic.Int64
}

type benchmarkTenantKey struct{}

type benchmarkCacheKey struct {
	TenantID   string
	ProviderID string
}

func (k benchmarkCacheKey) String() string {
	return fmt.Sprintf("%s:%s", k.TenantID, k.ProviderID)
}

func benchmarkSetup(ctx context.Context, out io.Writer, cfg benchmarkSetupConfig) error {
	if cfg.TenantCount <= 0 {
		cfg.TenantCount = 10
	}
	if cfg.Parallel <= 0 {
		cfg.Parallel = 5
	}

	fmt.Fprintln(out, "Starting docker services for benchmark...")
	if err := common.RunCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "up", "-d", "minio"); err != nil {
		return err
	}

	fmt.Fprintln(out, "Waiting for MinIO...")
	if err := common.WaitForMinIO(ctx, "objects-examples-minio", "admin", "adminsecretpassword"); err != nil {
		return err
	}

	tenants, err := provisionBenchmarkTenants(ctx, cfg)
	if err != nil {
		return err
	}

	path := resolvePath("tenants.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(tenants); err != nil {
		f.Close()
		return err
	}
	f.Close()

	fmt.Fprintf(out, "Provisioned %d tenants and wrote %s\n", len(tenants), path)
	return nil
}

func provisionBenchmarkTenants(ctx context.Context, cfg benchmarkSetupConfig) ([]benchmarkTenant, error) {
	type result struct {
		tenant benchmarkTenant
		err    error
	}

	results := make(chan result, cfg.TenantCount)
	sem := make(chan struct{}, cfg.Parallel)

	for i := 0; i < cfg.TenantCount; i++ {
		sem <- struct{}{}
		go func(idx int) {
			defer func() { <-sem }()
			r := result{tenant: benchmarkTenant{ID: idx}}

			bucket := fmt.Sprintf("tenant-%04d", idx)
			username := fmt.Sprintf("tenant-user-%04d", idx)
			password := fmt.Sprintf("tenant-secret-%04d", idx)

			if err := common.CreateMinIOUser(ctx, "objects-examples-minio", username, password); err != nil {
				r.err = err
			} else if err := common.CreateMinIOBucket(ctx, "objects-examples-minio", bucket); err != nil {
				r.err = err
			}

			r.tenant.Username = username
			r.tenant.Password = password
			r.tenant.Bucket = bucket
			r.tenant.AccessKey = username
			r.tenant.SecretKey = password
			results <- r
		}(i)
	}

	for i := 0; i < cfg.Parallel; i++ {
		sem <- struct{}{}
	}
	close(results)

	var tenants []benchmarkTenant
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		tenants = append(tenants, r.tenant)
	}

	return tenants, nil
}

func benchmarkTeardown(ctx context.Context, out io.Writer) error {
	fmt.Fprintln(out, "Stopping services and cleaning files...")
	_ = common.RunCommand(ctx, out, "docker-compose", "-f", composeFilePath(), "down", "--remove-orphans")
	os.Remove(resolvePath("tenants.json"))
	return nil
}

func runBenchmark(ctx context.Context, out io.Writer, cfg benchmarkRunConfig) error {
	if cfg.OpsPerTenant <= 0 {
		cfg.OpsPerTenant = 100
	}
	if cfg.Concurrent <= 0 {
		cfg.Concurrent = 10
	}

	fmt.Fprintln(out, "=== Multi-Provider High-Throughput Benchmark ===")

	tenants, err := loadBenchmarkTenants(resolvePath(cfg.TenantCfgPath))
	if err != nil {
		return fmt.Errorf("load tenants: %w", err)
	}
	fmt.Fprintf(out, "Loaded %d tenants\n", len(tenants))

	pool := eddy.NewClientPool[*benchmarkClient](30 * time.Minute)
	service := eddy.NewClientService(pool, eddy.WithConfigClone[
		*benchmarkClient,
		storage.ProviderCredentials](cloneProviderOptions))

	resolver := createBenchmarkResolver(tenants)

	var s benchmarkStats
	totalOps := cfg.OpsPerTenant * len(tenants)
	fmt.Fprintf(out, "\nRunning %d operations across %d tenants with %d workers...\n", totalOps, len(tenants), cfg.Concurrent)

	startMem := getMemStats()
	start := time.Now()

	if err := performBenchmarkOperations(ctx, service, resolver, tenants, cfg, &s); err != nil {
		return err
	}

	elapsed := time.Since(start)
	endMem := getMemStats()
	printBenchmarkResults(out, tenants, &s, elapsed, startMem, endMem)
	return nil
}

func getMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

func loadBenchmarkTenants(path string) ([]benchmarkTenant, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tenants []benchmarkTenant
	if err := json.NewDecoder(f).Decode(&tenants); err != nil {
		return nil, err
	}
	return tenants, nil
}

func createBenchmarkResolver(tenants []benchmarkTenant) *eddy.Resolver[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := eddy.NewResolver[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions]()

	for _, t := range tenants {
		tenantRule := t
		providerType := fmt.Sprintf("s3-tenant-%d", tenantRule.ID)
		builder := &benchmarkS3Builder{providerType: providerType}

		resolver.AddRule(eddy.NewRule[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions]().
			WhenFunc(func(ctx context.Context) bool {
				id, _ := ctx.Value(benchmarkTenantKey{}).(int)
				return id == tenantRule.ID
			}).
			Resolve(func(context.Context) (*eddy.ResolvedProvider[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions], error) {
				return &eddy.ResolvedProvider[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions]{
					Builder: builder,
					Output: storage.ProviderCredentials{
						AccessKeyID:     tenantRule.AccessKey,
						SecretAccessKey: tenantRule.SecretKey,
					},
					Config: storage.NewProviderOptions(
						storage.WithBucket(tenantRule.Bucket),
						storage.WithRegion("us-east-1"),
						storage.WithEndpoint("http://localhost:19000"),
					),
				}, nil
			}))
	}

	return resolver
}

func performBenchmarkOperations(ctx context.Context, service *eddy.ClientService[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions], tenants []benchmarkTenant, cfg benchmarkRunConfig, s *benchmarkStats) error {
	var wg sync.WaitGroup
	work := make(chan benchmarkTenant, len(tenants))
	errCh := make(chan error, cfg.Concurrent)

	for i := 0; i < cfg.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tenant := range work {
				if err := benchmarkWorkflow(ctx, service, resolver, tenant, cfg.OpsPerTenant, s); err != nil {
					errCh <- err
				}
			}
		}()
	}

	for _, t := range tenants {
		work <- t
	}
	close(work)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func benchmarkWorkflow(ctx context.Context, service *eddy.ClientService[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions], resolver *eddy.Resolver[*benchmarkClient, storage.ProviderCredentials, *storage.ProviderOptions], t benchmarkTenant, ops int, s *benchmarkStats) error {
	ctx = context.WithValue(ctx, benchmarkTenantKey{}, t.ID)

	resolutionOpt := resolver.Resolve(ctx)
	if !resolutionOpt.IsPresent() {
		s.errors.Add(1)
		return fmt.Errorf("resolve failed for tenant %d", t.ID)
	}
	resolution := resolutionOpt.MustGet()

	cacheKey := benchmarkCacheKey{
		TenantID:   fmt.Sprintf("tenant-%d", t.ID),
		ProviderID: resolution.Builder.ProviderType(),
	}

	clientOpt := service.GetClient(ctx, cacheKey, resolution.Builder, resolution.Output, resolution.Config)
	if !clientOpt.IsPresent() {
		s.errors.Add(1)
		s.cacheMisses.Add(1)
		return fmt.Errorf("client acquisition failed for tenant %d", t.ID)
	}
	client := clientOpt.MustGet()
	s.cacheHits.Add(1)

	objService := storage.NewObjectService()

	for i := 0; i < ops; i++ {
		content := strings.NewReader(fmt.Sprintf("Test data for tenant %d operation %d", t.ID, i))

		uploadOpts := &storage.UploadOptions{
			FileName:    fmt.Sprintf("file-%d.txt", i),
			ContentType: "text/plain",
		}

		uploaded, err := objService.Upload(ctx, client.Provider, content, uploadOpts)
		if err != nil {
			s.errors.Add(1)
			continue
		}
		s.uploads.Add(1)

		storageFile := &storagetypes.File{
			FileMetadata: storagetypes.FileMetadata{
				Key:         uploaded.Key,
				Size:        uploaded.Size,
				ContentType: uploaded.ContentType,
			},
		}

		if _, err := objService.Download(ctx, client.Provider, storageFile, &storage.DownloadOptions{}); err != nil {
			s.errors.Add(1)
			continue
		}
		s.downloads.Add(1)

		if err := objService.Delete(ctx, client.Provider, storageFile, &storagetypes.DeleteFileOptions{}); err != nil {
			s.errors.Add(1)
		}
	}

	return nil
}

func printBenchmarkResults(out io.Writer, tenants []benchmarkTenant, s *benchmarkStats, elapsed time.Duration, startMem, endMem runtime.MemStats) {
	fmt.Fprintln(out, "\n=== Results ===")
	fmt.Fprintf(out, "\nTenants: %d\n", len(tenants))
	fmt.Fprintf(out, "Total Operations: %d\n", s.uploads.Load()+s.downloads.Load())
	fmt.Fprintf(out, "  Uploads: %d\n", s.uploads.Load())
	fmt.Fprintf(out, "  Downloads: %d\n", s.downloads.Load())
	fmt.Fprintf(out, "  Errors: %d\n", s.errors.Load())

	fmt.Fprintln(out, "\nCache Statistics:")
	totalCache := s.cacheHits.Load() + s.cacheMisses.Load()
	hitRate := 0.0
	if totalCache > 0 {
		hitRate = float64(s.cacheHits.Load()) / float64(totalCache) * 100
	}
	fmt.Fprintf(out, "  Hits: %d\n", s.cacheHits.Load())
	fmt.Fprintf(out, "  Misses: %d\n", s.cacheMisses.Load())
	fmt.Fprintf(out, "  Hit Rate: %.2f%%\n", hitRate)

	fmt.Fprintln(out, "\nPerformance:")
	fmt.Fprintf(out, "  Total Time: %v\n", elapsed)
	totalOps := s.uploads.Load() + s.downloads.Load()
	if totalOps > 0 {
		fmt.Fprintf(out, "  Operations/sec: %.2f\n", float64(totalOps)/elapsed.Seconds())
		fmt.Fprintf(out, "  Avg Time/op: %v\n", elapsed/time.Duration(totalOps))
	}

	fmt.Fprintln(out, "\nMemory:")
	const bytesToMB = 1024 * 1024
	fmt.Fprintf(out, "  Start Alloc: %.2f MB\n", float64(startMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  End Alloc: %.2f MB\n", float64(endMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  Delta: %.2f MB\n", float64(endMem.Alloc-startMem.Alloc)/bytesToMB)
	fmt.Fprintf(out, "  Sys: %.2f MB\n", float64(endMem.Sys)/bytesToMB)
	fmt.Fprintf(out, "  NumGC: %d\n", endMem.NumGC-startMem.NumGC)

	fmt.Fprintln(out, "\n=== Benchmark completed successfully ===")
}

type benchmarkS3Builder struct {
	providerType string
}

func (b *benchmarkS3Builder) Build(_ context.Context, credentials storage.ProviderCredentials, options *storage.ProviderOptions) (*benchmarkClient, error) {
	opts := storage.NewProviderOptions()
	if options != nil {
		opts = options.Clone()
	}
	opts.Apply(storage.WithCredentials(credentials))

	provider, err := s3provider.NewS3Provider(opts, s3provider.WithUsePathStyle(true))
	if err != nil {
		return nil, err
	}

	return &benchmarkClient{
		Provider: provider,
		TenantID: opts.Bucket,
	}, nil
}

func (b *benchmarkS3Builder) ProviderType() string {
	return b.providerType
}
