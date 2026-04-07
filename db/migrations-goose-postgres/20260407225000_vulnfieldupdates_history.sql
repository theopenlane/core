-- +goose Up
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "cwe_ids" jsonb NULL, ADD COLUMN "vulnerable_version_range" character varying NULL, ADD COLUMN "first_patched_version" character varying NULL, ADD COLUMN "package_name" character varying NULL, ADD COLUMN "package_ecosystem" character varying NULL, ADD COLUMN "manifest_path" character varying NULL, ADD COLUMN "dependency_scope" character varying NULL, ADD COLUMN "dismissed_at" timestamptz NULL, ADD COLUMN "dismissed_reason" character varying NULL, ADD COLUMN "dismissed_comment" text NULL, ADD COLUMN "fixed_at" timestamptz NULL, ADD COLUMN "auto_dismissed_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "auto_dismissed_at", DROP COLUMN "fixed_at", DROP COLUMN "dismissed_comment", DROP COLUMN "dismissed_reason", DROP COLUMN "dismissed_at", DROP COLUMN "dependency_scope", DROP COLUMN "manifest_path", DROP COLUMN "package_ecosystem", DROP COLUMN "package_name", DROP COLUMN "first_patched_version", DROP COLUMN "vulnerable_version_range", DROP COLUMN "cwe_ids";
