-- +goose Up
-- create "impersonation_event_history" table
CREATE TABLE "impersonation_event_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "impersonation_type" character varying NOT NULL, "action" character varying NOT NULL, "reason" character varying NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "scopes" jsonb NULL, "user_id" character varying NOT NULL, "organization_id" character varying NOT NULL, "target_user_id" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "impersonationeventhistory_history_time" to table: "impersonation_event_history"
CREATE INDEX "impersonationeventhistory_history_time" ON "impersonation_event_history" ("history_time");
-- create "impersonation_events" table
CREATE TABLE "impersonation_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "impersonation_type" character varying NOT NULL, "action" character varying NOT NULL, "reason" character varying NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "scopes" jsonb NULL, "organization_id" character varying NOT NULL, "user_id" character varying NOT NULL, "target_user_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "impersonation_events_organizat_e9d1807913709f8160cb96a7d2f2627f" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "impersonation_events_users_impersonation_events" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "impersonation_events_users_targeted_impersonations" FOREIGN KEY ("target_user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);

-- +goose Down
-- reverse: create "impersonation_events" table
DROP TABLE "impersonation_events";
-- reverse: create index "impersonationeventhistory_history_time" to table: "impersonation_event_history"
DROP INDEX "impersonationeventhistory_history_time";
-- reverse: create "impersonation_event_history" table
DROP TABLE "impersonation_event_history";
