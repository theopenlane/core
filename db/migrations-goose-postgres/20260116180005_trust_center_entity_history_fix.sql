-- +goose Up
-- drop legacy table created by early history migration
DROP TABLE "trustcenter_entity_history";
-- create history_time index on correct table
CREATE INDEX "trustcenterentityhistory_history_time" ON "trust_center_entity_history" ("history_time");

-- +goose Down
-- drop canonical index
DROP INDEX "trustcenterentityhistory_history_time";
-- recreate legacy table
CREATE TABLE "trustcenter_entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "logo_file_id" character varying NULL, "url" character varying NULL, "trust_center_id" character varying NULL, "name" character varying NOT NULL, "entity_type_id" character varying NULL, PRIMARY KEY ("id"));
-- recreate legacy index
CREATE INDEX "trustcenterentityhistory_history_time" ON "trustcenter_entity_history" ("history_time");
