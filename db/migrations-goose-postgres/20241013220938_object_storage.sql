-- +goose Up
-- modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "file_name", DROP COLUMN "file_extension", DROP COLUMN "file_size", DROP COLUMN "content_type", ALTER COLUMN "store_key" DROP NOT NULL, DROP COLUMN "category", DROP COLUMN "annotation", ADD COLUMN "provided_file_name" character varying NOT NULL, ADD COLUMN "provided_file_extension" character varying NOT NULL, ADD COLUMN "provided_file_size" bigint NULL, ADD COLUMN "persisted_file_size" bigint NULL, ADD COLUMN "detected_mime_type" character varying NULL, ADD COLUMN "md5_hash" character varying NULL, ADD COLUMN "detected_content_type" character varying NOT NULL, ADD COLUMN "category_type" character varying NULL, ADD COLUMN "uri" character varying NULL, ADD COLUMN "storage_scheme" character varying NULL, ADD COLUMN "storage_volume" character varying NULL, ADD COLUMN "storage_path" character varying NULL, ADD COLUMN "file_contents" bytea NULL;
-- modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "file_id" jsonb NULL;
-- modify "files" table
ALTER TABLE "files" DROP COLUMN "file_name", DROP COLUMN "file_extension", DROP COLUMN "file_size", DROP COLUMN "content_type", ALTER COLUMN "store_key" DROP NOT NULL, DROP COLUMN "category", DROP COLUMN "annotation", DROP COLUMN "user_files", ADD COLUMN "provided_file_name" character varying NOT NULL, ADD COLUMN "provided_file_extension" character varying NOT NULL, ADD COLUMN "provided_file_size" bigint NULL, ADD COLUMN "persisted_file_size" bigint NULL, ADD COLUMN "detected_mime_type" character varying NULL, ADD COLUMN "md5_hash" character varying NULL, ADD COLUMN "detected_content_type" character varying NOT NULL, ADD COLUMN "category_type" character varying NULL, ADD COLUMN "uri" character varying NULL, ADD COLUMN "storage_scheme" character varying NULL, ADD COLUMN "storage_volume" character varying NULL, ADD COLUMN "storage_path" character varying NULL, ADD COLUMN "file_contents" bytea NULL;
-- create "contact_files" table
CREATE TABLE "contact_files" ("contact_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("contact_id", "file_id"), CONSTRAINT "contact_files_contact_id" FOREIGN KEY ("contact_id") REFERENCES "contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "contact_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "document_data_files" table
CREATE TABLE "document_data_files" ("document_data_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("document_data_id", "file_id"), CONSTRAINT "document_data_files_document_data_id" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "document_data_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "file_events" table
CREATE TABLE "file_events" ("file_id" character varying NOT NULL, "event_id" character varying NOT NULL, PRIMARY KEY ("file_id", "event_id"), CONSTRAINT "file_events_event_id" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "file_events_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "organization_setting_files" table
CREATE TABLE "organization_setting_files" ("organization_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("organization_setting_id", "file_id"), CONSTRAINT "organization_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "organization_setting_files_organization_setting_id" FOREIGN KEY ("organization_setting_id") REFERENCES "organization_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "template_files" table
CREATE TABLE "template_files" ("template_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("template_id", "file_id"), CONSTRAINT "template_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "template_files_template_id" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "file_id" jsonb NULL;
-- create "user_files" table
CREATE TABLE "user_files" ("user_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_id", "file_id"), CONSTRAINT "user_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_files_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "user_setting_files" table
CREATE TABLE "user_setting_files" ("user_setting_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("user_setting_id", "file_id"), CONSTRAINT "user_setting_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_setting_files_user_setting_id" FOREIGN KEY ("user_setting_id") REFERENCES "user_settings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "user_setting_files" table
DROP TABLE "user_setting_files";
-- reverse: create "user_files" table
DROP TABLE "user_files";
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "file_id";
-- reverse: create "template_files" table
DROP TABLE "template_files";
-- reverse: create "organization_setting_files" table
DROP TABLE "organization_setting_files";
-- reverse: create "file_events" table
DROP TABLE "file_events";
-- reverse: create "document_data_files" table
DROP TABLE "document_data_files";
-- reverse: create "contact_files" table
DROP TABLE "contact_files";
-- reverse: modify "files" table
ALTER TABLE "files" DROP COLUMN "file_contents", DROP COLUMN "storage_path", DROP COLUMN "storage_volume", DROP COLUMN "storage_scheme", DROP COLUMN "uri", DROP COLUMN "category_type", DROP COLUMN "detected_content_type", DROP COLUMN "md5_hash", DROP COLUMN "detected_mime_type", DROP COLUMN "persisted_file_size", DROP COLUMN "provided_file_size", DROP COLUMN "provided_file_extension", DROP COLUMN "provided_file_name", ADD COLUMN "user_files" character varying NULL, ADD COLUMN "annotation" character varying NULL, ADD COLUMN "category" character varying NULL, ALTER COLUMN "store_key" SET NOT NULL, ADD COLUMN "content_type" character varying NOT NULL, ADD COLUMN "file_size" bigint NULL, ADD COLUMN "file_extension" character varying NOT NULL, ADD COLUMN "file_name" character varying NOT NULL;
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "file_id";
-- reverse: modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "file_contents", DROP COLUMN "storage_path", DROP COLUMN "storage_volume", DROP COLUMN "storage_scheme", DROP COLUMN "uri", DROP COLUMN "category_type", DROP COLUMN "detected_content_type", DROP COLUMN "md5_hash", DROP COLUMN "detected_mime_type", DROP COLUMN "persisted_file_size", DROP COLUMN "provided_file_size", DROP COLUMN "provided_file_extension", DROP COLUMN "provided_file_name", ADD COLUMN "annotation" character varying NULL, ADD COLUMN "category" character varying NULL, ALTER COLUMN "store_key" SET NOT NULL, ADD COLUMN "content_type" character varying NOT NULL, ADD COLUMN "file_size" bigint NULL, ADD COLUMN "file_extension" character varying NOT NULL, ADD COLUMN "file_name" character varying NOT NULL;
