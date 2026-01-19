-- Modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "title" character varying NULL;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "remove_branding" boolean NULL DEFAULT false, ADD COLUMN "company_domain" character varying NULL, ADD COLUMN "security_contact" character varying NULL, ADD COLUMN "nda_approval_required" boolean NULL DEFAULT false;
-- Create "trust_center_nda_request_history" table
CREATE TABLE "trust_center_nda_request_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "trust_center_id" character varying NULL, "first_name" character varying NOT NULL, "last_name" character varying NOT NULL, "email" character varying NOT NULL, "company_name" character varying NULL, "reason" character varying NULL, "access_level" character varying NULL DEFAULT 'FULL', "status" character varying NULL DEFAULT 'REQUESTED', PRIMARY KEY ("id"));
-- Create index "trustcenterndarequesthistory_history_time" to table: "trust_center_nda_request_history"
CREATE INDEX "trustcenterndarequesthistory_history_time" ON "trust_center_nda_request_history" ("history_time");
