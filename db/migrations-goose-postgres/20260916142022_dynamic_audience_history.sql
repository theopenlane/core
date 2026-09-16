-- +goose Up
-- create "audience_history" table
CREATE TABLE "audience_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "name" character varying NOT NULL, "description" character varying NULL, "audience_type" character varying NOT NULL DEFAULT 'MANUAL', "filters" jsonb NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "audiencehistory_history_time" to table: "audience_history"
CREATE INDEX "audiencehistory_history_time" ON "audience_history" ("history_time");
-- create "audience_member_history" table
CREATE TABLE "audience_member_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "audience_id" character varying NOT NULL, "contact_id" character varying NULL, "user_id" character varying NULL, "group_id" character varying NULL, "identity_holder_id" character varying NULL, "subscriber_id" character varying NULL, "email" character varying NOT NULL, "full_name" character varying NULL, "metadata" jsonb NULL, PRIMARY KEY ("id"));
-- create index "audiencememberhistory_history_time" to table: "audience_member_history"
CREATE INDEX "audiencememberhistory_history_time" ON "audience_member_history" ("history_time");

-- +goose Down
-- reverse: create index "audiencememberhistory_history_time" to table: "audience_member_history"
DROP INDEX "audiencememberhistory_history_time";
-- reverse: create "audience_member_history" table
DROP TABLE "audience_member_history";
-- reverse: create index "audiencehistory_history_time" to table: "audience_history"
DROP INDEX "audiencehistory_history_time";
-- reverse: create "audience_history" table
DROP TABLE "audience_history";
