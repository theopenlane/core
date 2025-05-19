-- Create "job_runner_history" table
CREATE TABLE "job_runner_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'OFFLINE', "ip_address" character varying NOT NULL, PRIMARY KEY ("id"));
-- Create index "jobrunnerhistory_history_time" to table: "job_runner_history"
CREATE INDEX "jobrunnerhistory_history_time" ON "job_runner_history" ("history_time");
-- Create "job_runners" table
CREATE TABLE "job_runners" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_owned" boolean NULL DEFAULT false, "name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'OFFLINE', "ip_address" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "job_runners_organizations_job_runners" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "job_runners_ip_address_key" to table: "job_runners"
CREATE UNIQUE INDEX "job_runners_ip_address_key" ON "job_runners" ("ip_address");
-- Create index "jobrunner_display_id_owner_id" to table: "job_runners"
CREATE UNIQUE INDEX "jobrunner_display_id_owner_id" ON "job_runners" ("display_id", "owner_id");
-- Create index "jobrunner_id" to table: "job_runners"
CREATE UNIQUE INDEX "jobrunner_id" ON "job_runners" ("id");
-- Create "job_runner_tokens" table
CREATE TABLE "job_runner_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NULL, "last_used_at" timestamptz NULL, "is_active" boolean NULL DEFAULT true, "revoked_reason" character varying NULL, "revoked_by" character varying NULL, "revoked_at" timestamptz NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "job_runner_tokens_organizations_job_runner_tokens" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "job_runner_tokens_token_key" to table: "job_runner_tokens"
CREATE UNIQUE INDEX "job_runner_tokens_token_key" ON "job_runner_tokens" ("token");
-- Create index "jobrunnertoken_id" to table: "job_runner_tokens"
CREATE UNIQUE INDEX "jobrunnertoken_id" ON "job_runner_tokens" ("id");
-- Create index "jobrunnertoken_token_expires_at_is_active" to table: "job_runner_tokens"
CREATE INDEX "jobrunnertoken_token_expires_at_is_active" ON "job_runner_tokens" ("token", "expires_at", "is_active");
-- Create "job_runner_job_runner_tokens" table
CREATE TABLE "job_runner_job_runner_tokens" ("job_runner_id" character varying NOT NULL, "job_runner_token_id" character varying NOT NULL, PRIMARY KEY ("job_runner_id", "job_runner_token_id"), CONSTRAINT "job_runner_job_runner_tokens_job_runner_id" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "job_runner_job_runner_tokens_job_runner_token_id" FOREIGN KEY ("job_runner_token_id") REFERENCES "job_runner_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "job_runner_registration_tokens" table
CREATE TABLE "job_runner_registration_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "token" character varying NOT NULL, "expires_at" timestamptz NOT NULL, "last_used_at" timestamptz NULL, "job_runner_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "job_runner_registration_tokens_daddf3e078805108b2d174df258ddb4b" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "job_runner_registration_tokens_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "job_runner_registration_tokens_token_key" to table: "job_runner_registration_tokens"
CREATE UNIQUE INDEX "job_runner_registration_tokens_token_key" ON "job_runner_registration_tokens" ("token");
-- Create index "jobrunnerregistrationtoken_id" to table: "job_runner_registration_tokens"
CREATE UNIQUE INDEX "jobrunnerregistrationtoken_id" ON "job_runner_registration_tokens" ("id");
