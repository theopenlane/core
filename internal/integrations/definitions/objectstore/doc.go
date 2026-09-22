// Package objectstore provides the object storage integration definition: customer installations read and
// write objects in their own Google Cloud Storage, Amazon S3 or Cloudflare R2 bucket and import JSON records
// under a prefix, while the runtime integration imports platform-owned records as system-owned rows
package objectstore
