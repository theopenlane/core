# Object Storage

The `pkg/objects` module provides a modular, provider‑agnostic object storage layer. It exposes:

- A minimal set of top‑level aliases so consumers can work with a single package (`objects.File`, `objects.UploadOptions`, etc.)
- A provider interface and concrete providers for S3, Cloudflare R2, local disk, and database
- A small, focused `storage.ObjectService` that performs Upload/Download/Delete/Exists using any provider that implements the interface
- Utilities for MIME detection, safe buffering of uploads, and document parsing

This README is written for engineers integrating object storage into their services or CLIs, and explains how to:

- Choose and configure providers
- Upload, download, delete files, and generate presigned URLs
- Pass provider hints for dynamic selection
- Understand how dynamic and concurrent provider resolution works in this codebase

## Quickstart

Create a provider and use `storage.ObjectService` to upload and download files. The service is stateless and thread‑safe; it relies on the provider for IO.

```go
package main

import (
    "context"
    "strings"

    "github.com/theopenlane/shared/objects"
    storage "github.com/theopenlane/shared/objects/storage"
    disk "github.com/theopenlane/shared/objects/storage/providers/disk"
)

func main() {
    // 1) Choose a provider (disk example)
    opts := storage.NewProviderOptions(
        storage.WithBucket("./uploads"),    // local folder for disk provider
        storage.WithBasePath("./uploads"),
        storage.WithLocalURL("http://localhost:8080/files"), // optional, for presigned links
    )
    provider, err := disk.NewDiskProvider(opts)
    if err != nil { panic(err) }
    defer provider.Close()

    // 2) Use the object service with the provider
    svc := storage.NewObjectService()

    // 3) Upload
    content := strings.NewReader("hello world")
    file, err := svc.Upload(context.Background(), provider, content, &storage.UploadOptions{
        FileName:    "greeting.txt",
        ContentType: "text/plain; charset=utf-8",
        FileMetadata: storage.FileMetadata{ Key: "uploadFile" },
    })
    if err != nil { panic(err) }

    // 4) Download
    downloaded, err := svc.Download(context.Background(), provider, &storage.File{
        FileMetadata: storage.FileMetadata{ Key: file.Key, Bucket: file.Folder },
    }, &storage.DownloadOptions{})
    if err != nil { panic(err) }
    _ = downloaded // bytes and metadata
}
```

## Key Concepts

- `storage.Provider` is the interface every backend implements. It supports `Upload`, `Download`, `Delete`, `GetPresignedURL`, and `Exists`.
- `storage.ObjectService` is a thin layer over a given provider that implements core operations. It does not resolve providers itself.
- `objects.File` holds metadata about files (IDs, names, parent object, provider hints, etc.)
- `storage.UploadOptions` and `storage.DownloadOptions` capture request‑specific inputs (file name, content type, bucket/path, hints, and metadata)
- `storage.ProviderHints` lets you steer provider selection in environments that support dynamic resolution
- `ReaderToSeeker`, `NewBufferedReaderFromReader`, and `DetectContentType` help you safely stream and classify uploads
- `ParseDocument` extracts textual or structured content from common files (DOCX, JSON, YAML, plaintext)

## Built‑in Providers

All providers implement the same interface and are safe to use concurrently from goroutines. Construct them with provider‑specific options and credentials.

- Disk (`providers/disk`)
  - Stores files on local filesystem paths for development/testing
  - Options: `WithBucket`, `WithBasePath`, optional `WithLocalURL` for presigned URLs
  - Example: `disk.NewDiskProvider(options)`

- Amazon S3 (`providers/s3`)
  - Options: `WithBucket`, `WithRegion`, `WithEndpoint` (minio/alt endpoints)
  - Build via `NewS3Provider(options, ...)` or `NewS3ProviderResult(...).Get()`
  - Credentials can come from `ProviderOptions.Credentials` or environment

- Cloudflare R2 (`providers/r2`)
  - Similar to S3 in usage, supports account‑specific credentials and endpoints

- Database (`providers/database`)
  - Stores file bytes in the database; useful for low‑volume or migration scenarios

## Operations

### Upload

```go
seeker, _ := objects.ReaderToSeeker(reader)               // optional, ensures re‑readable stream
contentType, _ := storage.DetectContentType(seeker)       // optional, defaults to application/octet-stream for empty inputs

file, err := svc.Upload(ctx, provider, seeker, &storage.UploadOptions{
    FileName:    "report.pdf",
    ContentType: contentType,
    Bucket:      "reports",
    FileMetadata: storage.FileMetadata{
        Key:           "uploadFile",
        ProviderHints: &storage.ProviderHints{ PreferredProvider: storage.S3Provider },
    },
})
```

### Download

```go
meta, err := svc.Download(ctx, provider, &storage.File{
    ID: file.ID,
    FileMetadata: storage.FileMetadata{ Key: file.Key, Bucket: file.Folder },
}, &storage.DownloadOptions{ FileName: file.OriginalName })
```

### Presigned URL

```go
url, err := svc.GetPresignedURL(ctx, provider, &storage.File{ FileMetadata: storage.FileMetadata{ Key: file.Key, Bucket: file.Folder } }, &storage.PresignedURLOptions{ Duration: 15 * time.Minute })
```

### Delete & Exists

```go
_ = svc.Delete(ctx, provider, &storage.File{ FileMetadata: storage.FileMetadata{ Key: file.Key, Bucket: file.Folder } }, &storage.DeleteFileOptions{ Reason: "cleanup" })
exists, _ := provider.Exists(ctx, &storage.File{ FileMetadata: storage.FileMetadata{ Key: file.Key, Bucket: file.Folder } })
```

## Storage Structure

Files are stored with the following pattern:
```
s3://bucket-name/organization-id/file-id/filename.ext
r2://bucket-name/organization-id/file-id/filename.ext
database://default/file-id (database provider uses file ID, not paths)
```

### Example

For organization `01HYQZ5YTVJ0P2R2HF7N3W3MQZ`, and file record `01J1FILEXYZABCD5678` uploading `report.pdf`:
```
s3://my-bucket/01HYQZ5YTVJ0P2R2HF7N3W3MQZ/01J1FILEXYZABCD5678/report.pdf
```

### Implementation

When a file is uploaded through `HandleUploads`:

1. The organization ID is derived from the authenticated context or the persisted file record.
2. Metadata is persisted, returning the stored file record (including its database ID) and owning organization.
3. A folder path is built as `orgID/fileID`
4. The computed folder is passed as `FolderDestination` in upload options.
5. Storage providers use `path.Join(FolderDestination, FileName)` to construct the object key.

**Code:** `internal/objects/upload/handler.go`
```go
entFile, ownerOrgID, err := store.CreateFileRecord(ctx, file)
// ...
folderPath := buildStorageFolderPath(ownerOrgID, file, entFile.ID)
if folderPath != "" {
    uploadOpts.FolderDestination = folderPath
    file.Folder = folderPath
}
```

### Provider Implementation

**S3 Provider:** `pkg/objects/storage/providers/s3/provider.go`
```go
func (p *Provider) Upload(ctx context.Context, reader io.Reader, opts *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
    objectKey := opts.FileName
    if opts.FolderDestination != "" {
        objectKey = path.Join(opts.FolderDestination, opts.FileName)
    }
    // Upload with objectKey as the full path
}
```

**R2 Provider:** `pkg/objects/storage/providers/r2/provider.go`
- Same implementation as S3

**Database Provider:** `pkg/objects/storage/providers/database/provider.go`
- Stores files by file ID directly in database
- No folder structure concept (files are stored as binary blobs)

### Download Flow

When downloading files, the full key (including organization prefix) is stored in the database and used for retrieval:

```go
// File record in database
StoragePath: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ/01J1FILEXYZABCD5678/report.pdf"

// Used for download
provider.Download(ctx, &storagetypes.File{
    Key: file.StoragePath,  // Full path including organization
})
```

## Provider Hints and Dynamic Selection

`storage.ProviderHints` lets you carry metadata that a resolver can use to choose a storage backend at runtime. Common fields include:

- `KnownProvider` or `PreferredProvider` to suggest a backend (e.g., `s3`, `r2`, `disk`)
- `OrganizationID`, `IntegrationID`, `HushID` to route per tenant or integration
- `Module` and free‑form `Metadata` for feature‑level routing

Hints flow through `UploadOptions.FileMetadata.ProviderHints` and are copied into the resulting `objects.File`. Your resolver can read these hints and select an appropriate provider for each request.

## Dynamic & Concurrent Providers: How It Works Here

In this repository, dynamic provider selection and concurrency are handled by an orchestration layer that wraps `storage.ObjectService`:

- A resolver (built with the `eddy` library) looks at request context + `ProviderHints` to pick the right provider builder and runtime options
- A client pool caches and reuses provider clients per tenant and integration (`ProviderCacheKey`) to avoid reconnect churn
- The orchestrator (`internal/objects/Service`) then delegates actual IO to `storage.ObjectService` with the resolved provider
- This makes provider selection dynamic per request and safe under high concurrency

If you want a similar setup in your own project:

1) Define a `ProviderBuilder` for each backend you support, capable of reading configuration + hints and returning a `storage.Provider`

2) Use a resolver to map `(context, hints) -> (builder + runtime config)`

3) Use a thread‑safe client pool to cache provider instances by a stable key (e.g., org + integration), so concurrent requests reuse clients

4) Wrap `storage.ObjectService` to consume the resolved provider. The `ObjectService` is intentionally small and stateless to make this easy to compose

Within this repo, see `internal/objects/service.go` and `internal/objects/resolver` for a reference implementation.

## Validation & Readiness

You can validate providers at startup by calling `ListBuckets` or performing a trivial `Exists` check. In this repository, we expose helper validators that:

- Run best‑effort validation for all configured providers to surface misconfiguration early
- Optionally enforce provider‑specific availability via a per‑provider `ensureAvailable` flag in configuration

See `internal/objects/validators` for examples.

## Upload Utilities

- `objects.NewBufferedReaderFromReader` and `objects.ReaderToSeeker` wrap streams to support re‑reads for MIME detection and upload retries
- `storage.DetectContentType` detects MIME types and safely defaults to `application/octet-stream` for empty input
- `storage.ParseDocument` extracts content from DOCX/JSON/YAML/plaintext for downstream processing

## Error Handling

Provider implementations may return backend‑specific errors. The `ObjectService` surfaces these errors directly. For best UX:

- Detect content types before upload (or let the service detect for you)
- Use small in‑memory buffers for common cases; fall back to temp files for large streams
- Record provider hints so dynamic resolution can route requests consistently

## Configuration Reference

The repository includes a strongly typed configuration model (see `pkg/objects/storage/types.go`) suitable for wiring up providers in servers. Key fields:

- `ProviderConfig`: global flags (enabled, keys, size limits), plus a `Providers` section per backend
- `Providers.S3|CloudflareR2|GCS|Disk|Database`: enable flags, credentials, bucket/region/endpoint, and an `ensureAvailable` boolean for strict startup checks

## FAQ

- Can I upload to multiple providers at once
  - Yes, orchestrate multiple calls to `ObjectService.Upload` with different providers; the service is stateless and supports concurrent use

- How do I choose a provider per organization
  - Pass `ProviderHints` with an `OrganizationID` and build a resolver that maps orgs/integrations to providers. The internal orchestration layer demonstrates this pattern

- Do I need to use the internal orchestration layer
  - No. If you already know which provider to use, construct it directly and use `storage.ObjectService`. The dynamic bits are optional
